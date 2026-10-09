package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/configuration"
)

// `--print-config` answers what runs on a file, rule by registered rule, in ESLint's shape, so a parity
// check can diff the two engines (#mnmx9s4). The cases are the ones a key listing gets wrong: an
// override turning a rule off for one path, a rule nobody configured, and a core rule configured only
// through its typescript-eslint twin's key.
func TestPrintConfigResolvesEveryRegisteredRuleForAFile(t *testing.T) {
	t.Parallel()
	config := &configuration.Config{
		Root: "/repository",
		Rules: map[string]configuration.RuleSetting{
			"max-classes-per-file":               {Severity: configuration.SeverityError, Options: []json.RawMessage{json.RawMessage(`{"ignoreExpressions":true}`)}},
			"@typescript-eslint/no-invalid-this": {Severity: configuration.SeverityError},
			"no-var":                             {Severity: configuration.SeverityWarn},
		},
		Overrides: []configuration.Override{{
			Files: []string{"**/*.test.ts"},
			Rules: map[string]configuration.RuleSetting{"no-var": {Severity: configuration.SeverityOff}},
		}},
	}

	printed := func(path string) map[string][]json.RawMessage {
		t.Helper()
		var out strings.Builder
		if err := writeResolvedConfig(&out, config.Resolve(path)); err != nil {
			t.Fatal(err)
		}
		var document struct {
			Rules map[string][]json.RawMessage `json:"rules"`
		}
		if err := json.Unmarshal([]byte(out.String()), &document); err != nil {
			t.Fatalf("not JSON in ESLint's shape: %v\n%s", err, out.String())
		}
		return document.Rules
	}

	source := printed("/repository/source/Thing.ts")
	if got := source["max-classes-per-file"]; len(got) != 2 || string(got[0]) != "2" || compacted(t, got[1]) != `{"ignoreExpressions":true}` {
		t.Errorf("max-classes-per-file printed as %s, want [2, {\"ignoreExpressions\":true}]", got)
	}
	if got := source["no-var"]; len(got) != 1 || string(got[0]) != "1" {
		t.Errorf("no-var printed as %s, want [1]", got)
	}
	// The core rule has no key of its own and runs through its twin's, as the resolver decides.
	if got := source["no-invalid-this"]; len(got) != 1 || string(got[0]) != "2" {
		t.Errorf("core no-invalid-this printed as %s, want [2] through its twin's key", got)
	}
	if _, printedAtAll := source["no-debugger"]; printedAtAll {
		t.Error("a rule nobody configured was printed")
	}

	test := printed("/repository/source/Thing.test.ts")
	if got := test["no-var"]; len(got) != 1 || string(got[0]) != "0" {
		t.Errorf("no-var on a test file printed as %s, want [0] from the override", got)
	}
}

// compacted renders a JSON element without the indentation the printer adds.
func compacted(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		t.Fatal(err)
	}
	return buffer.String()
}
