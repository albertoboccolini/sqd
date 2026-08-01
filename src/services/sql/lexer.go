package sql

import (
	"strings"
	"unicode"

	"github.com/overthinkinglabs/sqd/src/models"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	currentChar  byte
}

func NewLexer(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.currentChar = 0
		lexer.position = lexer.readPosition
		return
	}

	lexer.currentChar = lexer.input[lexer.readPosition]
	lexer.position = lexer.readPosition
	lexer.readPosition++
}

func (lexer *Lexer) peekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	}

	return lexer.input[lexer.readPosition]
}

func (lexer *Lexer) skipWhitespace() {
	for lexer.currentChar == ' ' || lexer.currentChar == '\t' || lexer.currentChar == '\n' || lexer.currentChar == '\r' {
		lexer.readChar()
	}
}

func (lexer *Lexer) readString() string {
	quote := lexer.currentChar
	lexer.readChar()

	start := lexer.position
	for lexer.currentChar != quote && lexer.currentChar != 0 {
		lexer.readChar()
	}

	result := lexer.input[start:lexer.position]
	lexer.readChar()

	return result
}

func (lexer *Lexer) readIdentifier() string {
	start := lexer.position
	for lexer.isLetter(lexer.currentChar) || lexer.isDigit(lexer.currentChar) || lexer.currentChar == '_' || lexer.currentChar == '.' || lexer.currentChar == '*' || lexer.currentChar == '%' || lexer.currentChar == '-' || lexer.currentChar == '/' {
		lexer.readChar()
	}

	return lexer.input[start:lexer.position]
}

func (lexer *Lexer) NextToken() models.Token {
	var token models.Token

	lexer.skipWhitespace()

	switch lexer.currentChar {
	case '=':
		token = models.Token{Type: models.EQUALS, Literal: string(lexer.currentChar)}
		lexer.readChar()
		return token
	case '!':
		if lexer.peekChar() == '=' {
			lexer.readChar()
			token = models.Token{Type: models.NOT_EQUALS, Literal: "!="}
			lexer.readChar()
			return token
		}
		token.Literal = string(lexer.currentChar)
		token.Type = models.IDENTIFIER
		lexer.readChar()
		return token
	case ',':
		token = models.Token{Type: models.COMMA, Literal: string(lexer.currentChar)}
		lexer.readChar()
		return token
	case '(':
		token = models.Token{Type: models.LPAREN, Literal: string(lexer.currentChar)}
		lexer.readChar()
		return token
	case ')':
		token = models.Token{Type: models.RPAREN, Literal: string(lexer.currentChar)}
		lexer.readChar()
		return token
	case '\'', '"':
		token.Type = models.STRING
		token.Literal = lexer.readString()
		return token
	case 0:
		token.Type = models.EOF
		token.Literal = ""
		return token
	case '*':
		if lexer.isLetter(lexer.peekChar()) || lexer.peekChar() == '.' || lexer.peekChar() == '/' {
			token.Literal = lexer.readIdentifier()
			token.Type = models.IDENTIFIER
			return token
		}
		token = models.Token{Type: models.ASTERISK, Literal: string(lexer.currentChar)}
		lexer.readChar()
		return token
	}

	if lexer.isLetter(lexer.currentChar) || lexer.currentChar == '%' || lexer.currentChar == '/' {
		token.Literal = lexer.readIdentifier()
		token.Type = lexer.lookupKeyword(token.Literal)
		return token
	}

	token.Literal = string(lexer.currentChar)
	token.Type = models.IDENTIFIER
	lexer.readChar()
	return token
}

func (lexer *Lexer) lookupKeyword(ident string) models.TokenType {
	upper := strings.ToUpper(ident)

	keywords := map[string]models.TokenType{
		"SELECT":  models.SELECT,
		"FROM":    models.FROM,
		"WHERE":   models.WHERE,
		"UPDATE":  models.UPDATE,
		"SET":     models.SET,
		"DELETE":  models.DELETE,
		"COUNT":   models.COUNT,
		"LIKE":    models.LIKE,
		"NAME":    models.NAME,
		"CONTENT": models.CONTENT,
		"ORDER":   models.ORDER,
		"BY":      models.BY,
		"ASC":     models.ASC,
		"DESC":    models.DESC,
		"AND":     models.AND,
	}

	if token, ok := keywords[upper]; ok {
		return token
	}

	return models.IDENTIFIER
}

func (lexer *Lexer) isLetter(currentChar byte) bool {
	return unicode.IsLetter(rune(currentChar))
}

func (lexer *Lexer) isDigit(currentChar byte) bool {
	return unicode.IsDigit(rune(currentChar))
}
