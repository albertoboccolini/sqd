package models

type Action string

const (
	SELECT Action = "SELECT"
	COUNT  Action = "COUNT"
	UPDATE Action = "UPDATE"
	DELETE Action = "DELETE"
)
