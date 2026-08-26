package models

import (
	"regexp"
)

type Command struct {
	Action         TokenType
	File           string
	Pattern        *regexp.Regexp
	NegateContent  bool
	Replace        string
	Replacements   []Replacement
	Deletions      []Deletion
	IsBatch        bool
	SelectTargets  []TokenType
	WhereTarget    TokenType
	WherePattern   *regexp.Regexp
	NegateFileName bool
	ExtraPattern   *regexp.Regexp
	ExtraNegate    bool
	OrderBy        []OrderBy
	Limit          int
	OutputFormat   OutputFormat
}

type OrderBy struct {
	Column    TokenType
	Direction TokenType
}

type Replacement struct {
	Pattern *regexp.Regexp
	Negate  bool
	Replace string
}

type Deletion struct {
	Pattern *regexp.Regexp
	Negate  bool
}
