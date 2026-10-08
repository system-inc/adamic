package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestOptionalLiteralGuard(t *testing.T) {
	testOptionalConstruction(t, "optional_literal_guard.a", false)
}
func TestOptionalConditionalGuard(t *testing.T) {
	testOptionalConstruction(t, "optional_conditional_guard.a", false)
}
func TestOptionalDefinedGuard(t *testing.T) {
	testOptionalConstruction(t, "optional_defined_guard.a", false)
}
func TestOptionalRequiredViewGuard(t *testing.T) {
	testOptionalConstruction(t, "optional_required_view_guard.a", false)
}
func TestOptionalSpreadGuard(t *testing.T) {
	testOptionalConstruction(t, "optional_spread_guard.a", true)
}

func testOptionalConstruction(t *testing.T, fixtureName string, spread bool) {
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixtureName))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(strings.Replace(strings.Replace(string(source), "slot?: number | undefined", "slot?: number", 1), "copied?: number | undefined", "copied?: number", 1)), 0600); err != nil {
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
	wantChecks := 1
	if fixtureName == "optional_defined_guard.a" {
		wantChecks = 2
	}
	if len(checks) != wantChecks || (!strings.Contains(checks[0], "checked TS2375") && !strings.Contains(checks[0], "checked TS2322") && !strings.Contains(checks[0], "checked TS2379")) {
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
	if report := leaksUncached(t, lowered, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, lowered)); difference != "" {
		t.Fatal(difference)
	}
	if fixtureName == "optional_conditional_guard.a" {
		counted := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(c, counted, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		counts := execute(t, counted)
		if report := unbalanced(t, counts); report != "" {
			t.Fatal(report)
		}
		t.Logf("conditional counts: %s", counts.stderr)
		cleanup := regexp.MustCompile(`adamic_release\((adamic_local_[0-9]+_optionalConstruction)\);`)
		outer := cleanup.ReplaceAllString(c, "mutant_outer_cleanup($1);")
		if outer == c {
			t.Fatal("conditional cleanup mutant target absent")
		}
		helper := `#include "adamic.h"
#include <stdlib.h>
static void *mutant_outer_held = NULL;
static bool mutant_outer_registered = false;
static void mutant_release_outer(void) {
 adamic_release(mutant_outer_held);
 mutant_outer_held = NULL;
}
static void mutant_outer_cleanup(void *value) {
 if (!mutant_outer_registered) { (void)atexit(mutant_release_outer); mutant_outer_registered = true; }
 mutant_outer_held = value;
}
`
		outerBinary := filepath.Join(t.TempDir(), "outer-cleanup-mutant")
		if err := native.Build(helper+outer, outerBinary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		if difference := disagreement(expected, executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, outerBinary)); difference != "" {
			t.Fatal("outer cleanup changed observations before leak check: " + difference)
		}
		report := leakSanitizer(t, outerBinary)
		if !strings.Contains(report, "LeakSanitizer") {
			t.Fatalf("outer cleanup mutant not caught by leak check: %s", report)
		}
		if err := native.Build(helper+outer, counted, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		if report := unbalanced(t, execute(t, counted)); !strings.Contains(report, "heap values leaked") {
			t.Fatalf("outer cleanup mutant not caught by counts: %s", report)
		}
		t.Log("outer-scope cleanup mutant compiles and matches Node output; LeakSanitizer and allocation/free counts catch leaked earlier guard values")
	}
	if fixtureName == "optional_defined_guard.a" {
		declaration := regexp.MustCompile(`(?m)adamic_object \*\s*(adamic_local_[0-9]+_optionalView) = [^;]+;`)
		matches := declaration.FindAllStringSubmatchIndex(c, -1)
		if len(matches) != 2 {
			t.Fatalf("want two view declarations, got %d", len(matches))
		}
		last := matches[len(matches)-1]
		name := c[last[2]:last[3]]
		missing := c[:last[1]] + "\nadamic_release(" + name + ");\n" + name + " = NULL;\n" + c[last[1]:]
		mutantBinary := filepath.Join(t.TempDir(), "undefined-result-mutant")
		if err := native.Build(missing, mutantBinary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, mutantBinary)
		if result.exitCode != 70 || !strings.Contains(string(result.stderr), "optional contract produced undefined") {
			t.Fatalf("defined guard did not catch missing producer result: %d %s", result.exitCode, result.stderr)
		}
		t.Logf("undefined producer-result mutant caught by required-result guard, exit 70; checks: %v", checks)
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
	if spread {
		mutant = strings.ReplaceAll(c, "adamic_object_copy_reserving_checked(", "mutant_absent_spread(")
		if mutant == c {
			t.Fatal("spread mutant target absent")
		}
		helper = `#include "adamic.h"
static adamic_object *mutant_absent_spread(const adamic_object *source, const adamic_shape *reserved, const char *expression) {
 adamic_object *object = adamic_object_copy_reserving_checked(source, reserved, expression);
 for (size_t i=0;i<object->shape->count;i++) adamic_object_absent(object,i);
 return object;
}
`
	}
	binary = filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(helper+mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	field := "slot"
	if spread {
		field = "copied"
	}
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "optional write lost own presence: '"+field+"'") {
		t.Fatalf("want guard exit 70, got %d %s", result.exitCode, result.stderr)
	}
	t.Logf("Node/native/JavaScript agree; absent-construction mutant caught by guard, exit 70; %s", checks[0])
}
