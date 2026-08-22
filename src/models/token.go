package models

type TokenType int

const (
	IDENTIFIER TokenType = iota
	STRING
	SELECT
	FROM
	WHERE
	UPDATE
	SET
	DELETE
	COUNT
	LIKE
	NAME
	CONTENT
	EQUALS
	NOT_EQUALS
	LPAREN
	RPAREN
	ASTERISK
	COMMA
	ORDER
	BY
	ASC
	DESC
	AND
	LIMIT
	LINE
	EOF
)

type Token struct {
	Type    TokenType
	Literal string
}
