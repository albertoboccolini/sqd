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
	Deletions    []Deletion
	IsBatch      bool
}

type Replacement struct {
	Pattern    *regexp.Regexp
	Replace    string
	MatchExact bool
}

type Deletion struct {
	Pattern    *regexp.Regexp
	MatchExact bool
}
