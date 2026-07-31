package tests

import (
	"testing"

	"github.com/overthinkinglabs/sqd/tests/mock"
)

func TestLikePatternPrefix(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content LIKE '%test'")

	if !command.Pattern.MatchString("mytest") {
		t.Error("should match 'mytest'")
	}

	if command.Pattern.MatchString("testing") {
		t.Error("should not match 'testing'")
	}
}

func TestLikePatternSuffix(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content LIKE 'test%'")

	if !command.Pattern.MatchString("testing") {
		t.Error("should match 'testing'")
	}

	if command.Pattern.MatchString("mytest") {
		t.Error("should not match 'mytest'")
	}
}

func TestLikePatternBoth(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content LIKE '%test%'")

	if !command.Pattern.MatchString("mytesting") {
		t.Error("should match 'mytesting'")
	}
}

func TestLikePatternExact(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content LIKE 'test'")

	if !command.Pattern.MatchString("mytesting") {
		t.Error("LIKE without % should match substring")
	}

	if !command.Pattern.MatchString("test") {
		t.Error("LIKE without % should match exact string too")
	}

	if command.Pattern.MatchString("other") {
		t.Error("LIKE without % should not match unrelated string")
	}
}
