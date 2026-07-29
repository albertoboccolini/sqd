package models

import (
	"regexp"
)

type WhereCondition struct {
	Pattern    *regexp.Regexp
	Negate     bool
	Substrings []string
}

type Command struct {
	Action          TokenType
	File            string
	Pattern         *regexp.Regexp
	NegateContent   bool
	Replace         string
	Replacements    []Replacement
	Deletions       []Deletion
	IsBatch         bool
	SelectTarget    TokenType
	WhereTarget     TokenType
	WherePattern    *regexp.Regexp
	NegateFileName  bool
	ExtraPattern    *regexp.Regexp
	ExtraNegate     bool
	Substrings      []string
	ExtraSubstrings []string
	OrderBy         []OrderBy
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
