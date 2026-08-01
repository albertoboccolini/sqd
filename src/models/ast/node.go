package ast

import (
	"github.com/overthinkinglabs/sqd/src/models"
)

type Node interface {
	Accept(visitor Visitor) (models.Command, error)
}
