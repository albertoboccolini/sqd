package sql

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/ast"
)

type CommandBuilder struct{}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{}
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
		Action:       models.SELECT,
		SelectTarget: statement.Target,
		File:         statement.Source,
		OrderBy:      statement.OrderBy,
		WhereTarget:  models.CONTENT,
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
		File:         statement.Source,
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
		File:        statement.Source,
		IsBatch:     statement.IsBatch,
		Deletions:   statement.Deletions,
		WhereTarget: models.CONTENT,
	}

	if statement.WhereClause != nil && !statement.IsBatch {
		commandBuilder.populateWhere(&command, statement.WhereClause)
	}

	return command, nil
}
