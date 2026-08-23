package ast

import "github.com/overthinkinglabs/sqd/src/models"

type Select struct {
	Target      models.TokenType
	Targets     []models.TokenType
	Source      string
	WhereClause *Where
	OrderBy     []models.OrderBy
	IsCount     bool
	Limit       int
}

func (statement *Select) Accept(visitor Visitor) (models.Command, error) {
	return visitor.VisitSelect(statement)
}
