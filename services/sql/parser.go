package sql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/models/ast"
)

type Parser struct {
	extractor      *Extractor
	batchParser    *BatchParser
	commandBuilder *CommandBuilder
	lexer          *Lexer
	currentToken   models.Token
	peekToken      models.Token
}

func NewParser(extractor *Extractor, batchParser *BatchParser, commandBuilder *CommandBuilder) *Parser {
	return &Parser{
		extractor:      extractor,
		batchParser:    batchParser,
		commandBuilder: commandBuilder,
	}
}

func (parser *Parser) initLexer(input string) {
	parser.lexer = NewLexer(input)
	parser.nextToken()
	parser.nextToken()
}

func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) currentTokenIs(tokenType models.TokenType) bool {
	return parser.currentToken.Type == tokenType
}

func (parser *Parser) peekTokenIs(tokenType models.TokenType) bool {
	return parser.peekToken.Type == tokenType
}

func (parser *Parser) parseComparison(pattern **regexp.Regexp, negate *bool, substrings *[]string) {
	parser.nextToken()

	if parser.currentTokenIs(models.EQUALS) {
		parser.nextToken()
		exactMatch := parser.currentToken.Literal
		*pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		*substrings = []string{exactMatch}
		*negate = false
		parser.nextToken()
		return
	}

	if parser.currentTokenIs(models.NOT_EQUALS) {
		parser.nextToken()
		exactMatch := parser.currentToken.Literal
		*pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		*negate = true
		parser.nextToken()
		return
	}

	if parser.currentTokenIs(models.LIKE) {
		parser.nextToken()
		likePattern := parser.currentToken.Literal
		*pattern = parser.extractor.likeToRegex(likePattern)
		*substrings = parser.extractor.LikeSubstrings(likePattern)
		*negate = false
		parser.nextToken()
	}
}

func (parser *Parser) parseWhereClause() *ast.Where {
	if !parser.currentTokenIs(models.WHERE) {
		return nil
	}

	parser.nextToken()
	whereClause := &ast.Where{Target: models.CONTENT}

	if parser.currentTokenIs(models.NAME) {
		whereClause.Target = models.NAME
		parser.parseComparison(&whereClause.Pattern, &whereClause.Negate, &whereClause.Substrings)
		return whereClause
	}

	if parser.currentTokenIs(models.CONTENT) {
		parser.parseComparison(&whereClause.Pattern, &whereClause.Negate, &whereClause.Substrings)
		whereClause.And = parser.parseAndClause()
		if whereClause.And != nil {
			whereClause.And.Target = models.CONTENT
		}
	}

	return whereClause
}

func (parser *Parser) parseAndClause() *ast.Where {
	if !parser.currentTokenIs(models.AND) {
		return nil
	}

	parser.nextToken()
	andClause := &ast.Where{Target: models.CONTENT}

	if parser.currentTokenIs(models.CONTENT) {
		parser.parseComparison(&andClause.Pattern, &andClause.Negate, &andClause.Substrings)
	}

	return andClause
}

func (parser *Parser) parseOrderItem() models.OrderBy {
	item := models.OrderBy{}

	switch {
	case parser.currentTokenIs(models.NAME):
		item.Column = models.NAME
	case parser.currentTokenIs(models.CONTENT):
		item.Column = models.CONTENT
	default:
		return item
	}

	parser.nextToken()

	switch {
	case parser.currentTokenIs(models.ASC):
		item.Direction = models.ASC
		parser.nextToken()
	case parser.currentTokenIs(models.DESC):
		item.Direction = models.DESC
		parser.nextToken()
	}

	return item
}

func (parser *Parser) parseOrderBy() []models.OrderBy {
	if !parser.currentTokenIs(models.ORDER) || !parser.peekTokenIs(models.BY) {
		return nil
	}

	parser.nextToken()
	parser.nextToken()

	orderBy := make([]models.OrderBy, 0)

	for {
		orderItem := parser.parseOrderItem()
		if orderItem.Column == models.TokenType(0) {
			break
		}

		orderBy = append(orderBy, orderItem)

		if !parser.currentTokenIs(models.COMMA) {
			break
		}
		parser.nextToken()
	}

	return orderBy
}

func (parser *Parser) parseSelectStatement() ast.Node {
	statement := &ast.Select{Target: models.ASTERISK}

	if parser.peekTokenIs(models.COUNT) {
		statement.IsCount = true
		parser.nextToken()

		if parser.peekTokenIs(models.LPAREN) {
			parser.nextToken()
		}
	}

	if parser.peekTokenIs(models.NAME) {
		statement.Target = models.NAME
		parser.nextToken()
	}

	if parser.peekTokenIs(models.CONTENT) {
		statement.Target = models.CONTENT
		parser.nextToken()
	}

	if parser.peekTokenIs(models.ASTERISK) {
		parser.nextToken()
	}

	for !parser.currentTokenIs(models.EOF) {
		if parser.currentTokenIs(models.FROM) {
			parser.nextToken()
			if parser.currentTokenIs(models.IDENTIFIER) || parser.currentTokenIs(models.STRING) {
				statement.Source = parser.currentToken.Literal
			}
		}

		if whereClause := parser.parseWhereClause(); whereClause != nil {
			statement.WhereClause = whereClause
		}

		if orderBy := parser.parseOrderBy(); orderBy != nil {
			statement.OrderBy = orderBy
		}

		parser.nextToken()
	}

	statement.Source = strings.TrimSpace(statement.Source)
	return statement
}

func (parser *Parser) parseUpdateStatement(sql string) ast.Node {
	statement := &ast.Update{}

	parser.nextToken()
	if parser.currentTokenIs(models.IDENTIFIER) || parser.currentTokenIs(models.STRING) {
		statement.Source = parser.currentToken.Literal
	}

	upperSql := strings.ToUpper(sql)
	setCount := strings.Count(upperSql, "SET")

	if setCount > 1 {
		statement.IsBatch = true
		statement.Source = parser.extractor.extractFilename(sql, "UPDATE", "SET")
		statement.Replacements = parser.batchParser.parseBatchReplacements(sql)
		return statement
	}

	for !parser.currentTokenIs(models.EOF) {
		if whereClause := parser.parseWhereClause(); whereClause != nil {
			statement.WhereClause = whereClause
		}

		if parser.currentTokenIs(models.SET) {
			parser.nextToken()

			if parser.currentTokenIs(models.CONTENT) {
				parser.nextToken()
				if parser.currentTokenIs(models.EQUALS) {
					parser.nextToken()
					replacement := models.Replacement{Replace: parser.currentToken.Literal}
					if statement.WhereClause != nil {
						replacement.Pattern = statement.WhereClause.Pattern
						replacement.Negate = statement.WhereClause.Negate
					}
					statement.Replacements = []models.Replacement{replacement}
				}
			}
		}

		parser.nextToken()
	}

	statement.Source = strings.TrimSpace(statement.Source)
	return statement
}

func (parser *Parser) parseDeleteStatement(sql string) ast.Node {
	statement := &ast.Delete{}

	upperSql := strings.ToUpper(sql)
	whereIdx := strings.Index(upperSql, "WHERE")

	if whereIdx != -1 {
		whereClause := sql[whereIdx+5:]
		commaCount := strings.Count(whereClause, ",")

		if commaCount > 0 {
			statement.IsBatch = true
			statement.Source = parser.extractor.extractFilename(sql, "DELETE FROM", "WHERE")
			statement.Deletions = parser.batchParser.parseDeletions(sql)
			return statement
		}
	}

	for !parser.currentTokenIs(models.EOF) {
		if parser.currentTokenIs(models.FROM) {
			parser.nextToken()
			if parser.currentTokenIs(models.IDENTIFIER) || parser.currentTokenIs(models.STRING) {
				statement.Source = parser.currentToken.Literal
			}
		}

		if whereClause := parser.parseWhereClause(); whereClause != nil {
			statement.WhereClause = whereClause
		}

		parser.nextToken()
	}

	statement.Source = strings.TrimSpace(statement.Source)
	return statement
}

func (parser *Parser) Parse(sql string) (models.Command, error) {
	parser.initLexer(sql)

	var node ast.Node

	if parser.currentTokenIs(models.SELECT) {
		node = parser.parseSelectStatement()
	}

	if parser.currentTokenIs(models.UPDATE) {
		node = parser.parseUpdateStatement(sql)
	}

	if parser.currentTokenIs(models.DELETE) {
		node = parser.parseDeleteStatement(sql)
	}

	// The validator should have already rejected unrecognized statements
	// but we check just in case to avoid nil pointer dereferences later on.
	// So we return a generic error here instead of a displayable error
	// since this is an unexpected internal state rather than a user input issue.
	if node == nil {
		return models.Command{}, fmt.Errorf("unrecognized statement: %s", sql)
	}

	command, err := node.Accept(parser.commandBuilder)
	if err != nil {
		return models.Command{}, fmt.Errorf("failed to build command: %w", err)
	}

	return command, nil
}
