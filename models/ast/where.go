package ast

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/models"
)

type Where struct {
	Target  models.TokenType
	Pattern *regexp.Regexp
	Negate  bool
	And     *Where
}
