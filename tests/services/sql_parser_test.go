package tests

import (
	"testing"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

func TestParseSelect(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM test.txt WHERE content LIKE '%foo%'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.File != "test.txt" {
		t.Fatalf("expected test.txt, got %s", command.File)
	}

	if command.Pattern == nil {
		t.Fatal("pattern is nil")
	}

	if !command.Pattern.MatchString("foo") {
		t.Error("pattern should match 'foo'")
	}
}

func TestParseCount(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT COUNT(*) FROM file.sql WHERE content = 'exact'")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if command.File != "file.sql" {
		t.Fatalf("expected file.sql, got %s", command.File)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}
}

func TestParseUpdate(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("UPDATE file.txt SET content='new' WHERE content = 'old'")

	if command.Action != models.UPDATE {
		t.Fatalf("expected UPDATE, got %v", command.Action)
	}

	if command.Replace != "new" {
		t.Fatalf("expected 'new', got %s", command.Replace)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}

	if !command.Pattern.MatchString("old") {
		t.Error("pattern should match 'old'")
	}
}

func TestParseDelete(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("DELETE FROM file.txt WHERE content = 'remove'")

	if command.Action != models.DELETE {
		t.Fatalf("expected DELETE, got %v", command.Action)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}
}

func TestParseBatchUpdate(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("UPDATE file.txt SET content='a' WHERE content = 'x', SET content='b' WHERE content = 'y'")

	if !command.IsBatch {
		t.Fatal("expected batch mode")
	}

	if len(command.Replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %d", len(command.Replacements))
	}

	if command.Replacements[0].Replace != "a" {
		t.Errorf("first replacement: expected 'a', got %s", command.Replacements[0].Replace)
	}

	if command.Replacements[1].Replace != "b" {
		t.Errorf("second replacement: expected 'b', got %s", command.Replacements[1].Replace)
	}
}

func TestParseBatchDelete(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("DELETE FROM file.txt WHERE content = 'x', WHERE content = 'y'")

	if !command.IsBatch {
		t.Fatal("expected batch mode")
	}

	if len(command.Deletions) != 2 {
		t.Fatalf("expected 2 deletions, got %d", len(command.Deletions))
	}
}

func TestLikePatternPrefix(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM f WHERE content LIKE '%test'")

	if !command.Pattern.MatchString("mytest") {
		t.Error("should match 'mytest'")
	}

	if command.Pattern.MatchString("testing") {
		t.Error("should not match 'testing'")
	}
}

func TestLikePatternSuffix(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM f WHERE content LIKE 'test%'")

	if !command.Pattern.MatchString("testing") {
		t.Error("should match 'testing'")
	}

	if command.Pattern.MatchString("mytest") {
		t.Error("should not match 'mytest'")
	}
}

func TestLikePatternBoth(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM f WHERE content LIKE '%test%'")

	if !command.Pattern.MatchString("mytesting") {
		t.Error("should match 'mytesting'")
	}
}

func TestLikePatternExact(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM f WHERE content LIKE 'test'")

	if !command.Pattern.MatchString("test") {
		t.Error("should match 'test'")
	}

	if !command.Pattern.MatchString("testing") {
		t.Error("should match 'testing' (LIKE without % is 'contains')")
	}
}

func TestExactMatch(t *testing.T) {
	sqlParser := services.NewSQLParser()
	command := sqlParser.Parse("SELECT * FROM f WHERE content = 'exact'")

	if !command.Pattern.MatchString("exact") {
		t.Error("should match 'exact'")
	}

	if command.Pattern.MatchString("not exact") {
		t.Error("should not match 'not exact'")
	}
}
