package sql

import (
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/ast"
)

type CommandBuilder struct {
	fromAliases map[string]string
}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{fromAliases: make(map[string]string)}
}

func NewCommandBuilderWithAliases(fromAliases map[string]string) *CommandBuilder {
	if fromAliases == nil {
		fromAliases = make(map[string]string)
	}

	return &CommandBuilder{fromAliases: fromAliases}
}

func (commandBuilder *CommandBuilder) resolveSource(source string) string {
	if alias, exists := commandBuilder.fromAliases[source]; exists {
		return alias
	}

	firstSlash := strings.Index(source, "/")
	if firstSlash < 0 {
		return source
	}

	aliasKey := source[:firstSlash]
	if alias, exists := commandBuilder.fromAliases[aliasKey]; exists {
		return alias + source[firstSlash:]
	}

	return source
}

func (commandBuilder *CommandBuilder) populateWhere(command *models.Command, whereClause *ast.Where) {
	if whereClause == nil {
		return
	}

	command.WhereTarget = whereClause.Target

	if whereClause.Target == models.CONTENT {
		command.Pattern = whereClause.Pattern
		command.NegateContent = whereClause.Negate
	}

	if whereClause.Target == models.NAME {
		command.WherePattern = whereClause.Pattern
		command.NegateFileName = whereClause.Negate
	}

	if whereClause.And != nil {
		command.ExtraPattern = whereClause.And.Pattern
		command.ExtraNegate = whereClause.And.Negate
	}
}

func (commandBuilder *CommandBuilder) VisitSelect(statement *ast.Select) (models.Command, error) {
	command := models.Command{
		Action:        models.SELECT,
		SelectTarget:  statement.Target,
		SelectTargets: statement.Targets,
		File:          commandBuilder.resolveSource(statement.Source),
		OrderBy:       statement.OrderBy,
		Limit:         statement.Limit,
		WhereTarget:   models.CONTENT,
	}

	if statement.IsCount {
		command.Action = models.COUNT
	}

	commandBuilder.populateWhere(&command, statement.WhereClause)

	return command, nil
}

func (commandBuilder *CommandBuilder) VisitUpdate(statement *ast.Update) (models.Command, error) {
	command := models.Command{
		Action:       models.UPDATE,
		File:         commandBuilder.resolveSource(statement.Source),
		IsBatch:      statement.IsBatch,
		Replacements: statement.Replacements,
		WhereTarget:  models.CONTENT,
	}

	if statement.WhereClause != nil && !statement.IsBatch {
		commandBuilder.populateWhere(&command, statement.WhereClause)

		if len(statement.Replacements) > 0 {
			command.Replace = statement.Replacements[0].Replace
		}
	}

	return command, nil
}

func (commandBuilder *CommandBuilder) VisitDelete(statement *ast.Delete) (models.Command, error) {
	command := models.Command{
		Action:      models.DELETE,
		File:        commandBuilder.resolveSource(statement.Source),
		IsBatch:     statement.IsBatch,
		Deletions:   statement.Deletions,
		WhereTarget: models.CONTENT,
	}

	if statement.WhereClause != nil && !statement.IsBatch {
		commandBuilder.populateWhere(&command, statement.WhereClause)
	}

	return command, nil
}
