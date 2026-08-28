package sql

import (
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
)

type BatchParser struct {
	extractor *Extractor
}

func NewBatchParser(extractor *Extractor) *BatchParser {
	return &BatchParser{extractor: extractor}
}

func (batchParser *BatchParser) parseDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion
	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lexer := NewLexer(part)
		var deletion models.Deletion

		for {
			token := lexer.NextToken()
			if token.Type == models.EQUALS {
				token = lexer.NextToken()
				if token.Type == models.STRING {
					deletion.Pattern = batchParser.extractor.compileExact(token.Literal)
					deletion.Negate = false
					break
				}
			}

			if token.Type == models.NOT_EQUALS {
				token = lexer.NextToken()
				if token.Type == models.STRING {
					deletion.Pattern = batchParser.extractor.compileExact(token.Literal)
					deletion.Negate = true
					break
				}
			}

			if token.Type == models.LIKE {
				token = lexer.NextToken()
				if token.Type == models.STRING {
					deletion.Pattern = batchParser.extractor.compileLike(token.Literal)
					deletion.Negate = false
					break
				}
			}

			if token.Type == models.EOF {
				break
			}
		}

		deletions = append(deletions, deletion)
	}

	return deletions
}

func (batchParser *BatchParser) parseBatchReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement
	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var replacement models.Replacement
		lexer := NewLexer(part)

		for {
			token := lexer.NextToken()
			if token.Type == models.EOF {
				break
			}

			switch token.Type {
			case models.SET:
				if !setReplacement(&replacement, lexer) {
					continue
				}
			case models.WHERE:
				setWhereClause(batchParser, &replacement, lexer)
			}
		}

		replacements = append(replacements, replacement)
	}

	return replacements
}

func setReplacement(replacement *models.Replacement, lexer *Lexer) bool {
	token := lexer.NextToken()
	if token.Type != models.CONTENT {
		return false
	}

	token = lexer.NextToken()
	if token.Type != models.EQUALS {
		return false
	}

	token = lexer.NextToken()
	if token.Type != models.STRING {
		return false
	}

	replacement.Replace = token.Literal
	return true
}

func setWhereClause(batchParser *BatchParser, replacement *models.Replacement, lexer *Lexer) {
	token := lexer.NextToken()
	if token.Type != models.CONTENT {
		return
	}

	token = lexer.NextToken()
	switch token.Type {
	case models.EQUALS:
		token = lexer.NextToken()
		if token.Type == models.STRING {
			replacement.Pattern = batchParser.extractor.compileExact(token.Literal)
			replacement.Negate = false
		}
	case models.NOT_EQUALS:
		token = lexer.NextToken()
		if token.Type == models.STRING {
			replacement.Pattern = batchParser.extractor.compileExact(token.Literal)
			replacement.Negate = true
		}
	case models.LIKE:
		token = lexer.NextToken()
		if token.Type == models.STRING {
			replacement.Pattern = batchParser.extractor.compileLike(token.Literal)
			replacement.Negate = false
		}
	}
}
