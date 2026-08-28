package tests

import (
	"testing"

	"github.com/overthinkinglabs/sqd/src/services/sql"
)

func TestCommandBuilderResolvesAliasAsFirstPathSegment(t *testing.T) {
	aliases := map[string]string{
		"obsidian": "/home/user/Documents/obsidian",
	}
	builder := sql.NewCommandBuilder(aliases)

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	parser := sql.NewParser(extractor, batchParser, builder)

	command, err := parser.Parse("SELECT content FROM obsidian/Università/*.md WHERE content LIKE \"### %\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/home/user/Documents/obsidian/Università/*.md"
	if command.File != expected {
		t.Errorf("expected %q, got %q", expected, command.File)
	}
}

func TestCommandBuilderResolvesFromAlias(t *testing.T) {
	aliases := map[string]string{"md": "*.md", "logs": "*.log"}
	builder := sql.NewCommandBuilder(aliases)

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	parser := sql.NewParser(extractor, batchParser, builder)

	command, err := parser.Parse("SELECT * FROM md WHERE content LIKE \"todo\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if command.File != "*.md" {
		t.Errorf("expected file pattern *.md, got %s", command.File)
	}
}

func TestCommandBuilderKeepsLiteralSourceWhenNoAlias(t *testing.T) {
	aliases := map[string]string{"md": "*.md"}
	builder := sql.NewCommandBuilder(aliases)

	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	parser := sql.NewParser(extractor, batchParser, builder)

	command, err := parser.Parse("SELECT * FROM notes.txt WHERE content LIKE \"todo\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if command.File != "notes.txt" {
		t.Errorf("expected file notes.txt, got %s", command.File)
	}
}
