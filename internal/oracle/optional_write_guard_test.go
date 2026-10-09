package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestOptionalWriteGuard(t *testing.T) {
	t.Parallel()
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_write_guard.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(strings.ReplaceAll(string(source), "number | undefined", "number"), "string | undefined", "string")), 0600); err != nil {
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
	if len(checks) != 3 || !strings.Contains(checks[0], "checked TS2412") {
		t.Fatalf("want three emitted checks, got %v", checks)
	}
	expected := onNode(t, fixture)
	c := native.C(lowered)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "native")
		if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if difference := disagreement(expected, got); difference != "" {
			t.Fatal(difference)
		}
	}
	js := filepath.Join(t.TempDir(), "main.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(lowered)), 0600); err != nil {
		t.Fatal(err)
	}
	jsResult := onNode(t, js)
	if difference := disagreement(expected, jsResult); difference != "" {
		t.Fatalf("%s: Node %d %s %s; JS %d %s %s", difference, expected.exitCode, expected.stdout, expected.stderr, jsResult.exitCode, jsResult.stdout, jsResult.stderr)
	}
	mutant := strings.ReplaceAll(c, "adamic_object_write_field(", "mutant_absent_write(")
	if mutant == c {
		t.Fatal("mutant target absent")
	}
	for _, name := range []string{"slot", "text", "classSlot"} {
		t.Run("absent-"+name, func(t *testing.T) {
			helper := `#include "adamic.h"
#include <string.h>
static adamic_value *mutant_absent_write(adamic_object *object, const char *name, adamic_slot_cache *cache) {
 adamic_value *slot = adamic_object_write_field(object, name, cache);
 if (strcmp(name, "` + name + `") == 0) adamic_object_absent(object, adamic_slot_index(object, slot));
 return slot;
}
`
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(helper+mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 70 || !strings.Contains(string(result.stderr), "optional write lost own presence: '"+name+"'") {
				t.Fatalf("want presence guard exit 70, got %d %s", result.exitCode, result.stderr)
			}
			t.Logf("absent-store mutant %s caught by presence guard, exit 70", name)
		})
	}
	t.Logf("native and JavaScript match Node; emitted checks: %v", checks)
}

func TestOptionalImplementsRepresentation(t *testing.T) {
	t.Parallel()
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_implements_guard.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	source = []byte(strings.Replace(string(source), "moduleResolverHost?: number | undefined", "moduleResolverHost?: number", 1))
	if err := os.WriteFile(path, source, 0600); err != nil {
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
	if len(checks) != 1 || !strings.Contains(checks[0], "checked TS2420") {
		t.Fatalf("want one checked implements relation, got %v", checks)
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
	mutant := strings.ReplaceAll(c, "adamic_object_write_field(", "mutant_absent_write(")
	if mutant == c {
		t.Fatal("mutant target absent")
	}
	helper := `#include "adamic.h"
static adamic_value *mutant_absent_write(adamic_object *object, const char *name, adamic_slot_cache *cache) {
 adamic_value *slot = adamic_object_write_field(object, name, cache);
 adamic_object_absent(object, adamic_slot_index(object, slot));
 return slot;
}
`
	binary = filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(helper+mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if disagreement(expected, result) == "" {
		t.Fatal("Node did not catch required-field absence mutant")
	}
	t.Logf("Node caught required-field absence mutant: exit %d stdout %q stderr %q; %s", result.exitCode, result.stdout, result.stderr, checks[0])
}
