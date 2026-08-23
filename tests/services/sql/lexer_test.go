package tests

import (
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services/sql"
)

func TestLexerBasicTokens(t *testing.T) {
	input := "SELECT * FROM file.txt WHERE content LIKE 'pattern'"

	tests := []struct {
		expectedType    models.TokenType
		expectedLiteral string
	}{
		{models.SELECT, "SELECT"},
		{models.ASTERISK, "*"},
		{models.FROM, "FROM"},
		{models.IDENTIFIER, "file.txt"},
		{models.WHERE, "WHERE"},
		{models.CONTENT, "content"},
		{models.LIKE, "LIKE"},
		{models.STRING, "pattern"},
		{models.EOF, ""},
	}

	lexer := sql.NewLexer(input)

	for i, expected := range tests {
		token := lexer.NextToken()

		if token.Type != expected.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%d, got=%d",
				i, expected.expectedType, token.Type)
		}

		if token.Literal != expected.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, expected.expectedLiteral, token.Literal)
		}
	}
}

func TestLexerUpdateStatement(t *testing.T) {
	input := "UPDATE file.txt SET content='new' WHERE content = 'old'"

	lexer := sql.NewLexer(input)

	expectedTokens := []models.TokenType{
		models.UPDATE,
		models.IDENTIFIER,
		models.SET,
		models.CONTENT,
		models.EQUALS,
		models.STRING,
		models.WHERE,
		models.CONTENT,
		models.EQUALS,
		models.STRING,
		models.EOF,
	}

	for i, expected := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expected {
			t.Errorf("token[%d] wrong. expected=%d, got=%d", i, expected, token.Type)
		}
	}
}

func TestLexerCountStatement(t *testing.T) {
	input := "SELECT COUNT(content) FROM *.md WHERE content LIKE '%TODO%'"

	lexer := sql.NewLexer(input)

	token := lexer.NextToken()
	if token.Type != models.SELECT {
		t.Errorf("expected SELECT, got %d", token.Type)
	}

	token = lexer.NextToken()
	if token.Type != models.COUNT {
		t.Errorf("expected COUNT, got %d", token.Type)
	}
}

func TestLexerDeleteStatement(t *testing.T) {
	input := "DELETE FROM *.go WHERE name LIKE '%.tmp'"

	lexer := sql.NewLexer(input)

	expectedTokens := []models.TokenType{
		models.DELETE,
		models.FROM,
		models.IDENTIFIER,
		models.WHERE,
		models.NAME,
		models.LIKE,
		models.STRING,
		models.EOF,
	}

	for i, expected := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expected {
			t.Errorf("token[%d] wrong. expected=%d, got=%d (literal=%q)", i, expected, token.Type, token.Literal)
		}
	}
}

func TestLexerLimitAndLineKeywords(t *testing.T) {
	input := "SELECT line FROM file.txt WHERE content LIKE '%x%' LIMIT 10"

	lexer := sql.NewLexer(input)

	expectedTokens := []models.TokenType{
		models.SELECT,
		models.LINE,
		models.FROM,
		models.IDENTIFIER,
		models.WHERE,
		models.CONTENT,
		models.LIKE,
		models.STRING,
		models.LIMIT,
		models.IDENTIFIER,
		models.EOF,
	}

	for i, expected := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expected {
			t.Errorf("token[%d] wrong. expected=%d, got=%d (literal=%q)", i, expected, token.Type, token.Literal)
		}
	}
}

func TestLexerHandlesUnicodeInIdentifier(t *testing.T) {
	lexer := sql.NewLexer("obsidian/Università/*.md")
	tokens := []string{}

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token.Literal)
		if token.Type == models.EOF {
			break
		}
	}

	expected := "obsidian/Università/*.md"
	if tokens[0] != expected {
		t.Errorf("expected identifier %q, got %q", expected, tokens[0])
	}
}
