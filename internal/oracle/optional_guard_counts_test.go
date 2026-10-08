package oracle

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Count the generated project input, including the inserted optional guards.
// The .a fixture itself keeps true types and is the independent Node witness.
var optionalGuardFixtures = []string{
	"optional_write_guard.a", "optional_implements_guard.a",
	"optional_literal_guard.a", "optional_spread_guard.a", "optional_conditional_guard.a",
	"optional_required_view_guard.a", "optional_defined_guard.a", "optional_nullable_view_guard.a",
	"optional_nested_guard.a", "optional_nested_nullable_guard.a", "optional_nested_assignment_guard.a",
	"optional_array_guard.a", "optional_callback_guard.a",
}

func countedOptionalGuard(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(strings.Replace(string(source), "slot?: number | undefined", "slot?: number", 1), "copied?: number | undefined", "copied?: number", 1)
	if name == "optional_write_guard.a" {
		text = strings.ReplaceAll(strings.ReplaceAll(string(source), "number | undefined", "number"), "string | undefined", "string")
	}
	if name == "optional_implements_guard.a" {
		text = strings.Replace(string(source), "moduleResolverHost?: number | undefined", "moduleResolverHost?: number", 1)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(`{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["ES2024"],"target":"ES2024","module":"ESNext","moduleResolution":"Bundler","types":[]},"files":["main.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.ExplainedOptionalChecks()) == 0 {
		t.Fatal("counts must include at least one compiled optional contract")
	}
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	command, args := pinnedStack(binary)
	result := execute(t, command, args...)
	if report := unbalanced(t, result); report != "" {
		t.Fatal(report)
	}
	match := countsLine.FindSubmatch(result.stderr)
	return fmt.Sprintf("| internal/oracle/testdata/%s (project options) | %s | %s | %s | %s | %s | %s |", name, match[1], match[2], match[3], match[4], match[5], match[6])
}
