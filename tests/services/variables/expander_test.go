package variables

import (
	"testing"

	"github.com/overthinkinglabs/sqd/src/services/variables"
)

func TestExpanderReplacesDefinedVariables(t *testing.T) {
	expander := variables.NewExpander()

	result := expander.Expand("hello $name and ${place}", map[string]string{
		"name":  "world",
		"place": "moon",
	})

	expected := "hello world and moon"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpanderFallsBackToEnvironmentVariables(t *testing.T) {
	expander := variables.NewExpander()
	t.Setenv("SQD_TEST_ENV", "from_env")

	result := expander.Expand("value is $SQD_TEST_ENV", nil)

	expected := "value is from_env"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpanderPrefersFlagVariablesOverEnvironment(t *testing.T) {
	expander := variables.NewExpander()
	t.Setenv("SQD_TEST_PRIO", "env")

	result := expander.Expand("value is $SQD_TEST_PRIO", map[string]string{
		"SQD_TEST_PRIO": "flag",
	})

	expected := "value is flag"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpanderLeavesUndefinedPlaceholdersUnchanged(t *testing.T) {
	expander := variables.NewExpander()

	result := expander.Expand("keep $undefined and ${also_undefined}", nil)

	expected := "keep $undefined and ${also_undefined}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpanderReplacesWithEmptyValueFromFlags(t *testing.T) {
	expander := variables.NewExpander()

	result := expander.Expand("value is [$empty]", map[string]string{
		"empty": "",
	})

	expected := "value is []"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpanderReplacesWithEmptyEnvironmentValue(t *testing.T) {
	expander := variables.NewExpander()
	t.Setenv("SQD_TEST_EMPTY", "")

	result := expander.Expand("value is [$SQD_TEST_EMPTY]", nil)

	expected := "value is []"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
