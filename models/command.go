package models

import "regexp"

type Command struct {
	Action  string
	File    string
	Pattern *regexp.Regexp
	Replace string
}
