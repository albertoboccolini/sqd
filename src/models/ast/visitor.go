package ast

import "github.com/overthinkinglabs/sqd/src/models"

type Visitor interface {
	VisitSelect(statement *Select) (models.Command, error)
	VisitUpdate(statement *Update) (models.Command, error)
	VisitDelete(statement *Delete) (models.Command, error)
}
