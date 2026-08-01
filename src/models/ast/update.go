package ast

import "github.com/overthinkinglabs/sqd/src/models"

type Update struct {
	Source       string
	Replacements []models.Replacement
	WhereClause  *Where
	IsBatch      bool
}

func (statement *Update) Accept(visitor Visitor) (models.Command, error) {
	return visitor.VisitUpdate(statement)
}
