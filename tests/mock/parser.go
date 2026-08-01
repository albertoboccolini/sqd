package mock

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services/sql"
)

type MockParser struct {
	parser *sql.Parser
}

func NewParser() *MockParser {
	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	commandBuilder := sql.NewCommandBuilder()
	return &MockParser{
		parser: sql.NewParser(extractor, batchParser, commandBuilder),
	}
}

func (mockParser *MockParser) Parse(query string) models.Command {
	command, err := mockParser.parser.Parse(query)
	if err != nil {
		panic(err)
	}
	return command
}
