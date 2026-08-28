package tests

import (
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/tests/mock"
)

func TestParseSelect(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM test.txt WHERE content LIKE '%foo%'")

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
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM file.sql WHERE content = 'exact'")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if command.File != "file.sql" {
		t.Fatalf("expected file.sql, got %s", command.File)
	}
}

func TestParseUpdate(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("UPDATE file.txt SET content='new' WHERE content = 'old'")

	if command.Action != models.UPDATE {
		t.Fatalf("expected UPDATE, got %v", command.Action)
	}

	if command.Replace != "new" {
		t.Fatalf("expected 'new', got %s", command.Replace)
	}

	if !command.Pattern.MatchString("old") {
		t.Error("pattern should match 'old'")
	}
}

func TestParseDelete(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("DELETE FROM file.txt WHERE content = 'remove'")

	if command.Action != models.DELETE {
		t.Fatalf("expected DELETE, got %v", command.Action)
	}
}

func TestExactMatch(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content = 'exact'")

	if !command.Pattern.MatchString("exact") {
		t.Error("should match 'exact'")
	}

	if command.Pattern.MatchString("not exact") {
		t.Error("should not match 'not exact'")
	}
}

func TestUpdatePreservesSpace(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("UPDATE *.md SET content = '#### ' WHERE content LIKE '### %'")

	input := "### Title"
	result := command.Pattern.ReplaceAllString(input, command.Replace)

	if result != "#### Title" {
		t.Errorf("expected '#### Title', got '%s'", result)
	}
}

func TestParseWhereNameEquals(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE name = 'test.txt'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if command.WherePattern == nil {
		t.Fatal("WherePattern is nil")
	}

	if !command.WherePattern.MatchString("test.txt") {
		t.Error("pattern should match 'test.txt'")
	}

	if command.WherePattern.MatchString("other.txt") {
		t.Error("pattern should not match 'other.txt' (exact match)")
	}
}

func TestParseWhereNameLike(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.md WHERE name LIKE '%.tmp'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if command.WherePattern == nil {
		t.Fatal("WherePattern is nil")
	}

	if !command.WherePattern.MatchString("file.tmp") {
		t.Error("pattern should match 'file.tmp'")
	}

	if !command.WherePattern.MatchString("test.tmp") {
		t.Error("pattern should match 'test.tmp'")
	}

	if command.WherePattern.MatchString("file.md") {
		t.Error("pattern should not match 'file.md'")
	}
}

func TestParseCountWithWhereName(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM *.go WHERE name LIKE 'test%'")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if !command.WherePattern.MatchString("test_file.go") {
		t.Error("pattern should match 'test_file.go'")
	}

	if command.WherePattern.MatchString("main.go") {
		t.Error("pattern should not match 'main.go'")
	}
}

func TestParseOrderByName(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%foo%' ORDER BY name")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.TokenType(0) {
		t.Errorf("expected no explicit direction, got %v", command.OrderBy[0].Direction)
	}
}

func TestParseOrderByNameAsc(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%test%' ORDER BY name ASC")

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.ASC {
		t.Errorf("expected ASC direction, got %v", command.OrderBy[0].Direction)
	}
}

func TestParseOrderByContentDesc(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.md WHERE content LIKE '%%' ORDER BY content DESC")

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.CONTENT {
		t.Errorf("expected ORDER BY content, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.DESC {
		t.Errorf("expected DESC direction, got %v", command.OrderBy[0].Direction)
	}
}

func TestParseOrderByMultipleColumns(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%test%' ORDER BY name ASC, content DESC")

	if len(command.OrderBy) != 2 {
		t.Fatalf("expected 2 order by clauses, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected first ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.ASC {
		t.Errorf("expected first direction ASC, got %v", command.OrderBy[0].Direction)
	}

	if command.OrderBy[1].Column != models.CONTENT {
		t.Errorf("expected second ORDER BY content, got %v", command.OrderBy[1].Column)
	}

	if command.OrderBy[1].Direction != models.DESC {
		t.Errorf("expected second direction DESC, got %v", command.OrderBy[1].Direction)
	}
}

func TestParseCountWithOrderBy(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM *.go WHERE content LIKE '%func%' ORDER BY name ASC")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.ASC {
		t.Errorf("expected ASC direction, got %v", command.OrderBy[0].Direction)
	}
}

func TestParseCountNameWithOrderBy(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(name) FROM *.txt WHERE content LIKE '%test%' ORDER BY content DESC")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if len(command.SelectTargets) != 1 || command.SelectTargets[0] != models.NAME {
		t.Fatalf("expected SelectTargets [NAME], got %v", command.SelectTargets)
	}

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.CONTENT {
		t.Errorf("expected ORDER BY content, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.DESC {
		t.Errorf("expected DESC direction, got %v", command.OrderBy[0].Direction)
	}
}

func TestParseOrderByWithoutDirection(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%test%' ORDER BY name, content")

	if len(command.OrderBy) != 2 {
		t.Fatalf("expected 2 order by clauses, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected first ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[1].Column != models.CONTENT {
		t.Errorf("expected second ORDER BY content, got %v", command.OrderBy[1].Column)
	}
}

func TestParseWhereAnd(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%foo%' AND content = 'exact'")

	if command.Pattern == nil {
		t.Fatal("primary pattern is nil")
	}

	if !command.Pattern.MatchString("foo") {
		t.Error("primary pattern should match 'foo'")
	}

	if command.ExtraPattern == nil {
		t.Fatal("extra pattern is nil")
	}

	if !command.ExtraPattern.MatchString("exact") {
		t.Error("extra pattern should match 'exact'")
	}

	if command.ExtraPattern.MatchString("not exact") {
		t.Error("extra pattern should not match 'not exact'")
	}
}

func TestParseSelectLine(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT line FROM file.txt WHERE content = 'exact'")

	if len(command.SelectTargets) != 1 || command.SelectTargets[0] != models.LINE {
		t.Fatalf("expected SelectTargets [LINE], got %v", command.SelectTargets)
	}
}

func TestParseLimit(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%test%' LIMIT 5")

	if command.Limit != 5 {
		t.Fatalf("expected Limit 5, got %d", command.Limit)
	}
}

func TestParseOrderByWithLimit(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE content LIKE '%test%' ORDER BY name DESC LIMIT 3")

	if len(command.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(command.OrderBy))
	}

	if command.OrderBy[0].Column != models.NAME {
		t.Errorf("expected ORDER BY name, got %v", command.OrderBy[0].Column)
	}

	if command.OrderBy[0].Direction != models.DESC {
		t.Errorf("expected DESC direction, got %v", command.OrderBy[0].Direction)
	}

	if command.Limit != 3 {
		t.Fatalf("expected Limit 3, got %d", command.Limit)
	}
}
