package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionalLiteralGuard(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_literal_guard.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(source), "number | undefined", "number")), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(`{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["ES2024"],"target":"ES2024","module":"ESNext","moduleResolution":"Bundler","types":[]},"files":["main.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	checks := program.ExplainedOptionalChecks()
	if len(checks) != 1 || !strings.Contains(checks[0], "checked TS2375") {
		t.Fatalf("want one emitted construction check, got %v", checks)
	}
	expected := onNode(t, fixture)
	c := native.C(lowered)
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(c, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(expected, execute(t, binary)); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, lowered)); difference != "" {
		t.Fatal(difference)
	}
	mutant := strings.ReplaceAll(c, "adamic_object_new(", "mutant_absent_literal(")
	if mutant == c {
		t.Fatal("mutant target absent")
	}
	helper := `#include "adamic.h"
static adamic_object *mutant_absent_literal(const adamic_shape *shape) {
 adamic_object *object = adamic_object_new(shape);
 for (size_t i=0;i<shape->count;i++) adamic_object_absent(object,i);
 return object;
}
`
	binary = filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(helper+mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "optional write lost own presence: 'slot'") {
		t.Fatalf("want guard exit 70, got %d %s", result.exitCode, result.stderr)
	}
	t.Logf("Node/native/JavaScript agree; absent-construction mutant caught by guard, exit 70; %s", checks[0])
}
