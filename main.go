package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/commands"
	"github.com/overthinkinglabs/sqd/src/services/default_config"
	"github.com/overthinkinglabs/sqd/src/services/dry_mode"
	"github.com/overthinkinglabs/sqd/src/services/files"
	"github.com/overthinkinglabs/sqd/src/services/sql"
	"github.com/overthinkinglabs/sqd/src/services/variables"
)

func splitQueries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := range data {
		if data[i] == ';' {
			return i + 1, data[:i], nil
		}
	}

	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func handleError(errorHandler *services.ErrorHandler, err error) {
	errorHandler.HandleError(err)
	os.Exit(1)
}

func parseVariable(value string) (string, string, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("variable must be in key=value format")
	}

	return parts[0], parts[1], nil
}

func registerVariableFlag(flagSet *flag.FlagSet, variables map[string]string) {
	flagSet.Func("var", "Define a variable as key=value (can be repeated)", func(value string) error {
		key, val, err := parseVariable(value)
		if err != nil {
			return err
		}

		variables[key] = val
		return nil
	})
}

func executeQuery(
	query string,
	definedVariables map[string]string,
	useTransaction,
	dryRun bool,
	showDetailedOutputInDryMode bool,
	outputFormat models.OutputFormat,
	defaultConfig *models.DefaultConfig,
) error {
	expander := variables.NewExpander()
	expandedQuery := expander.Expand(query, definedVariables)

	validator := sql.NewValidator()
	if err := validator.Validate(expandedQuery); err != nil {
		return err
	}

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	commandBuilder := sql.NewCommandBuilderWithAliases(defaultConfig.FromAliases)
	parser := sql.NewParser(extractor, batchParser, commandBuilder)
	command, err := parser.Parse(expandedQuery)
	if err != nil {
		return err
	}

	utils := services.NewUtils(defaultConfig)

	ignoreList, ignoreErr := files.LoadIgnoreList(".sqdignore")
	if ignoreErr != nil {
		return ignoreErr
	}

	finder := files.NewFinder(ignoreList)
	processor := files.NewProcessor(utils)
	parallelizer := files.NewParallelizer(utils)

	foundFiles, walkWarnings := finder.FindFiles(command.File)

	if len(foundFiles) == 0 {
		if walkWarnings != nil {
			return walkWarnings
		}

		return displayable_errors.NewNoFilesFoundError(command.File)
	}

	dryModeFileReader := dry_mode.NewFileReader(utils)
	dryModeChangeProcessor := dry_mode.NewChangeProcessor(dryModeFileReader, utils)
	dryModeRunner := dry_mode.NewRunner(dryModeChangeProcessor, utils)

	transactioner := commands.NewTransactioner(utils)
	sorter := commands.NewSorter()
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)
	dispatcher := commands.NewDispatcher(
		searcher,
		counter,
		updater,
		deleter,
		transactioner,
		dryModeRunner,
		utils,
		parallelizer,
	)

	dispatchErr := dispatcher.Execute(
		command,
		foundFiles,
		useTransaction,
		dryRun,
		showDetailedOutputInDryMode,
		outputFormat,
	)

	if dispatchErr != nil {
		var errorCollection *models.ErrorCollection
		if errors.As(dispatchErr, &errorCollection) {
			utils.AddWalkWarnings(errorCollection, walkWarnings)
			return errorCollection
		}

		if walkWarnings != nil {
			errorCollection := models.NewErrorCollection()
			errorCollection.Add(dispatchErr)
			utils.AddWalkWarnings(errorCollection, walkWarnings)
			return errorCollection
		}

		return dispatchErr
	}

	return walkWarnings
}

func executeQueriesFromFile(
	filePath string,
	definedVariables map[string]string,
	useTransaction,
	dryRun bool,
	showDetailedOutputInDryMode bool,
	outputFormat models.OutputFormat,
	defaultConfig *models.DefaultConfig,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return displayable_errors.NewFileReadError(filePath, err)
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(splitQueries)

	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}

		fmt.Printf("%s\n", query)
		if err := executeQuery(
			query,
			definedVariables,
			useTransaction,
			dryRun,
			showDetailedOutputInDryMode,
			outputFormat,
			defaultConfig,
		); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return displayable_errors.NewFileReadError(filePath, err)
	}

	return nil
}

func registerCommonFlags(flagSet *flag.FlagSet, variables map[string]string, defaultTransaction bool) (*bool, *string) {
	transactionFlag := flagSet.Bool("transaction", defaultTransaction, "Enable transaction mode with rollback on failure")
	flagSet.BoolVar(transactionFlag, "t", defaultTransaction, "Enable transaction mode with rollback on failure")
	queryFile := flagSet.String("file", "", "Path to a file containing queries to execute")
	flagSet.StringVar(queryFile, "f", "", "Path to a file containing queries to execute")
	registerVariableFlag(flagSet, variables)

	return transactionFlag, queryFile
}

func handleInitCommand(loader *default_config.Loader, errorHandler *services.ErrorHandler) {
	cwd, err := os.Getwd()
	if err != nil {
		handleError(errorHandler, err)
	}

	initializer := default_config.NewInitializer(loader)

	configPath, configCreated, ignorePath, ignoreCreated, err := initializer.InitResult(cwd)
	if err != nil {
		handleError(errorHandler, err)
	}

	if configCreated {
		fmt.Printf("Created global config: %s\n", configPath)
	} else {
		fmt.Printf("Global config already exists: %s\n", configPath)
	}

	if ignoreCreated {
		fmt.Printf("Created local ignore file: %s\n", ignorePath)
	} else {
		fmt.Printf("Local ignore file already exists: %s\n", ignorePath)
	}
}

func handleDryModeCommand(args []string, errorHandler *services.ErrorHandler, defaultConfig *models.DefaultConfig) {
	dryFlagSet := flag.NewFlagSet("dry", flag.ExitOnError)
	completeFlag := dryFlagSet.Bool("complete", false, "Show file names with modified lines")
	dryFlagSet.BoolVar(completeFlag, "c", false, "Show file names with modified lines")

	definedVariables := make(map[string]string)
	transactionFlag, queryFile := registerCommonFlags(dryFlagSet, definedVariables, false)

	if err := dryFlagSet.Parse(args); err != nil {
		handleError(errorHandler, err)
	}

	if *queryFile != "" {
		if err := executeQueriesFromFile(
			*queryFile,
			definedVariables,
			*transactionFlag,
			true, *completeFlag,
			models.TextOutput,
			defaultConfig,
		); err != nil {
			handleError(errorHandler, err)
		}

		return
	}

	if len(dryFlagSet.Args()) == 0 {
		fmt.Println("Usage: sqd dry [flags] 'query'")
		fmt.Println("\nFlags:")
		fmt.Println("  -c, --complete    Show file names with modified lines")
		fmt.Println("  -t, --transaction Enable transaction mode with rollback on failure")
		fmt.Println("  -f, --file        Path to a file containing queries to execute")
		fmt.Println("      --var         Define a variable as key=value (can be repeated)")
		os.Exit(1)
	}

	query := strings.Join(dryFlagSet.Args(), " ")

	if err := executeQuery(
		query,
		definedVariables,
		*transactionFlag,
		true,
		*completeFlag,
		models.TextOutput,
		defaultConfig,
	); err != nil {
		handleError(errorHandler, err)
	}
}

func main() {
	errorHandler := services.NewErrorHandler()
	loader := default_config.NewLoader()

	if len(os.Args) > 1 && os.Args[1] == "init" {
		handleInitCommand(loader, errorHandler)
		return
	}

	defaultConfig, configErr := loader.LoadConfig()
	if configErr != nil {
		handleError(errorHandler, configErr)
	}

	if len(os.Args) > 1 && os.Args[1] == "dry" {
		handleDryModeCommand(os.Args[2:], errorHandler, defaultConfig)
		return
	}

	versionFlag := flag.Bool("version", false, "Print version information")
	flag.BoolVar(versionFlag, "v", false, "Print version information")

	definedVariables := make(map[string]string)
	transactionFlag, queryFile := registerCommonFlags(flag.CommandLine, definedVariables, false)

	jsonFlag := flag.Bool("json", false, "Print query output as JSON")
	csvFlag := flag.Bool("csv", false, "Print query output as CSV")

	flag.Parse()

	if *jsonFlag && *csvFlag {
		handleError(errorHandler, displayable_errors.NewInvalidQueryError("Cannot use both --json and --csv"))
	}

	outputFormat := models.TextOutput
	if *jsonFlag {
		outputFormat = models.JSONOutput
	}

	if *csvFlag {
		outputFormat = models.CSVOutput
	}

	if *versionFlag {
		fmt.Printf("v%s\n", models.VERSION)
		os.Exit(0)
	}

	if *queryFile != "" {
		if err := executeQueriesFromFile(
			*queryFile,
			definedVariables,
			*transactionFlag,
			false,
			false,
			outputFormat,
			defaultConfig,
		); err != nil {
			handleError(errorHandler, err)
		}

		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("sqd | A SQL-like document editor")
		fmt.Println("\nUsage: sqd [flags] 'query'")
		fmt.Println("       sqd dry [flags] 'query'")
		fmt.Println("       sqd init")
		fmt.Println("\nStatements:")
		fmt.Println("  SELECT	Display matching lines")
		fmt.Println("  UPDATE	Replace content in matching lines")
		fmt.Println("  DELETE	Remove matching lines")
		fmt.Println("  COUNT		Count matching lines (using *, name, or content)")
		fmt.Println("\nClauses:")
		fmt.Println("  FROM		Specify the target file or file pattern")
		fmt.Println("  WHERE		Define matching criteria")
		fmt.Println("  SET		Define replacement content for UPDATE statements (only for content)")
		fmt.Println("  ORDER BY 	Sort matching lines (using name or content)")
		fmt.Println("  LIMIT n\tReturn only the first n results")
		fmt.Println("\nOperators:")
		fmt.Println("  =		Exact match")
		fmt.Println("  !=		Negation of exact match")
		fmt.Println("  LIKE		Pattern match with wildcards (%)")
		fmt.Println("\nExamples:")
		fmt.Println("  sqd 'SELECT *|name|content FROM *.txt WHERE content LIKE pattern ORDER BY name|content ASC|DESC'")
		fmt.Println("  sqd dry 'UPDATE file.txt SET old TO new WHERE content = match, SET foo TO bar WHERE content = other'")
		fmt.Println("  sqd dry -c 'UPDATE file.txt SET old TO new WHERE content = match'")
		fmt.Println("  sqd -t 'DELETE FROM file.txt WHERE content = exact_match'")
		fmt.Println("  sqd -f path/to/file")
		fmt.Println("  sqd --var old=foo --var new=bar 'UPDATE *.md SET content=\"$new\" WHERE content = \"$old\"'")
		fmt.Println("\nCommands:")
		fmt.Println("  init              Create global config and local .sqdignore")
		fmt.Println("  dry               Show what would be done without making changes")
		fmt.Println("    -c, --complete  Show file names with modified lines")
		fmt.Println("\nFlags:")
		fmt.Println("  -f, --file        Path to a file containing queries to execute")
		fmt.Println("  -t, --transaction Enable transaction mode with rollback on failure")
		fmt.Println("  -v, --version     Show the version information")
		fmt.Println("      --json        Print SELECT/COUNT output as JSON")
		fmt.Println("      --csv         Print SELECT/COUNT output as CSV")
		fmt.Println("      --var         Define a variable as key=value (can be repeated)")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")

	if err := executeQuery(sql, definedVariables, *transactionFlag, false, false, outputFormat, defaultConfig); err != nil {
		handleError(errorHandler, err)
	}
}
