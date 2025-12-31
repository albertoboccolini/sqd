package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/services"
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.BoolVar(versionFlag, "v", false, "Print version information")
	transactionFlag := flag.Bool("transaction", false, "Enable transaction mode with rollback on failure")
	flag.BoolVar(transactionFlag, "t", false, "Enable transaction mode with rollback on failure")
	dryRunFlag := flag.Bool("dry-run", false, "Show what would be done without making changes")
	flag.BoolVar(dryRunFlag, "d", false, "Show what would be done without making changes")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", services.SQD_VERSION)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: sqd 'query'")
		fmt.Println("\nCommands:")
		fmt.Println("  SELECT - Display matching lines")
		fmt.Println("  UPDATE - Replace content in matching lines")
		fmt.Println("  DELETE - Remove matching lines")
		fmt.Println("\nExamples:")
		fmt.Println("  sqd 'SELECT * FROM file.txt WHERE content LIKE pattern'")
		fmt.Println("  sqd 'UPDATE file.txt SET old TO new WHERE content = match, SET foo TO bar WHERE content = other'")
		fmt.Println("  sqd 'DELETE FROM file.txt WHERE content = exact_match'")
		fmt.Println("\nFlags:")
		fmt.Println("  -d, --dry-run\t\tShow what would be done without making changes")
		fmt.Println("  -t, --transaction	Enable transaction mode with rollback on failure")
		fmt.Println("  -v, --version		Show the version information")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")

	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse(sql)

	fileFinder := services.NewFileFinder()
	utils := services.NewUtils()

	files := fileFinder.FindFiles(command.File)
	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(1)
	}

	fileOperator := services.NewFileOperator(utils)
	fileOperator.ExecuteCommand(command, files, *transactionFlag, *dryRunFlag)
}
