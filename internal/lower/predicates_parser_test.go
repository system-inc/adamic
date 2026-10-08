package lower

import (
	"os"
	"strings"
	"testing"
)

func TestParserCallbackArguments(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/generic_callback_parameter_read.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
	// Checker inference inherits the overload claim. Only the independent
	// argument-body proof can keep this callback from reaching the callee.
	lying := strings.Replace(string(source), "function isDefined(value: string | undefined): value is string {\n    return value !== undefined;\n}", "function isDefined(value: string | undefined): value is string;\nfunction isDefined(value: string | undefined): boolean { return true; }", 1)
	lying = strings.Replace(lying, `"word", isDefined`, `"word", (value: string | undefined) => isDefined(value)`, 1)
	if _, err := lowerSource(t, lying); err == nil || !strings.Contains(err.Error(), "unproven predicate argument for parameter test") {
		t.Fatalf("want independently refused generic callback, got %v", err)
	}
}

func TestParserCallbackLie(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/callback_parameter_lie.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err == nil || !strings.Contains(err.Error(), "unproven predicate argument for parameter test") {
		t.Fatalf("want callback body proof refusal, got %v", err)
	}
}
