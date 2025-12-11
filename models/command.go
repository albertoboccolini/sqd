package models

import (
	"regexp"
)

type Command struct {
	Action       Action
	File         string
	Pattern      *regexp.Regexp
	Replace      string
	MatchExact   bool
	Replacements []Replacement
	IsBatch      bool
}

type Replacement struct {
	Pattern    *regexp.Regexp
	Replace    string
	MatchExact bool
}
