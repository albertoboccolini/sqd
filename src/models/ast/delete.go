package ast

import "github.com/overthinkinglabs/sqd/src/models"

type Delete struct {
	Source      string
	Deletions   []models.Deletion
	WhereClause *Where
	IsBatch     bool
}

func (statement *Delete) Accept(visitor Visitor) (models.Command, error) {
	return visitor.VisitDelete(statement)
}
