package models

type OutputFormat int

const (
	TextOutput OutputFormat = iota
	JSONOutput
	CSVOutput
)
