package oracle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Fixtures keep their .a suffix. A temporary .ts copy exercises the distinct
// source-boundary policy without adding unchecked .ts sources to the repository.
func TestCheckedAny(t *testing.T) {
	for _, name := range checkedAnyFixtures {
		t.Run(name, func(t *testing.T) {
			program, path := checkedAnyProgram(t, name)
			truth := onNode(t, path)
			want := truth
			if strings.HasSuffix(name, "_misfit") {
				expression, required, tag := "config.value", "string", "number"
				if name == "parameter_misfit" {
					expression, required, tag = "value", "number", "string"
				}
				if name == "arithmetic_misfit" {
					expression, required, tag = "value", "number", "string"
				}
				if name == "stale_misfit" {
					expression, required, tag = "value", "string", "number"
				}
				if name == "assertion_misfit" {
					expression, required, tag = "'wrong' as any", "number", "string"
				}
				if name == "property_null_misfit" {
					expression, required, tag = "config", "non-null property receiver", "null"
				}
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: " + expression + " needs " + required + ", found " + tag + "\n")}
			}
			js := onJavaScriptBackend(t, program)
			if d := disagreement(want, js); d != "" {
				t.Fatalf("JavaScript %s: got %+v want %+v", d, js, want)
			}
			native, binary := nativelyUncached(t, program)
			if d := disagreement(want, native); d != "" {
				t.Fatalf("native %s: got %+v want %+v", d, native, want)
			}
			if strings.HasSuffix(name, "_misfit") && name != "property_null_misfit" {
				changed := 0
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != "checked_any_use" {
						continue
					}
					function.Body = function.Body[1:]
					changed++
				}
				if changed == 0 {
					t.Fatal("mutant removed no check")
				}
				mutant := onJavaScriptBackend(t, program)
				if d := disagreement(truth, mutant); d != "" {
					t.Fatalf("unchecked mutant must reproduce Node: %s: %+v", d, mutant)
				}
				if d := disagreement(want, mutant); d == "" {
					t.Fatal("pinned boundary observation did not catch missing check")
				}
				t.Logf("mutant removing %d guards caught by pinned exit/stderr: unchecked JavaScript matches Node %q", changed, mutant.stdout)
			}
			if name == "property_null_misfit" {
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name == "checked_any_property_receiver" {
						function.Body = function.Body[1:]
					}
				}
				mutant := onJavaScriptBackend(t, program)
				if disagreement(want, mutant) == "" {
					t.Fatal("null receiver mutant escaped the pinned check")
				}
				mutatedNative, _ := nativelyUncached(t, program)
				if disagreement(want, mutatedNative) == "" {
					t.Fatal("native null receiver mutant escaped the pinned check")
				}
				t.Logf("null receiver mutant caught: JavaScript exit=%d stderr=%q; native exit=%d stderr=%q", mutant.exitCode, mutant.stderr, mutatedNative.exitCode, mutatedNative.stderr)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("source Node stdout=%q stderr=%q exit=%d; backends stdout=%q stderr=%q exit=%d", truth.stdout, truth.stderr, truth.exitCode, want.stdout, want.stderr, want.exitCode)
			if _, err := lowered(t, path); err == nil || !strings.Contains(err.Error(), "explicit any in .a") {
				t.Fatalf(".a must refuse explicit any: %v", err)
			}
		})
	}
}

var checkedAnyFixtures = []string{"property", "property_misfit", "parameter", "parameter_misfit", "arithmetic", "arithmetic_misfit", "boolean", "transport", "validated", "config_validated", "property_null_misfit", "string_length", "assertion", "assertion_misfit", "stale_misfit", "config_diagnostics"}

func checkedAnyProgram(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path := filepath.Join(repository, "internal/load/testdata/0.1/refuse/checked_any", name+".a")
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ts := filepath.Join(t.TempDir(), "boundary.ts")
	if err := os.WriteFile(ts, source, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{ts})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func checkedAnyCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, name := range append(append([]string{}, checkedAnyFixtures...), checkedJSONFixtures...) {
		program, _ := checkedAnyProgram(t, name)
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		command, args := pinnedStack(binary)
		result := execute(t, command, args...)
		expected := 0
		if strings.HasSuffix(name, "_misfit") {
			expected = 70
		}
		if result.exitCode != expected {
			t.Fatalf("%s exit %d stderr %q", name, result.exitCode, result.stderr)
		}
		match := countsLine.FindSubmatch(result.stderr)
		if match == nil {
			t.Fatalf("%s has no counts: %q", name, result.stderr)
		}
		rows = append(rows, fmt.Sprintf("| checked .ts copy: checked_any/%s.a | %s | %s | %s | %s | %s | %s |", name, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	return rows
}

func TestCheckedAnyUnsupportedContracts(t *testing.T) {
	for _, name := range []string{"callable", "prototype", "array_view", "field_view", "json_alias_write", "json_alias_update", "json_iteration", "json_reflection", "json_array_method", "json_unprepared"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/load/testdata/0.1/refuse/checked_any", name+".a")
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			ts := filepath.Join(t.TempDir(), "boundary.ts")
			if err := os.WriteFile(ts, source, 0600); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{ts})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err == nil || program != nil {
				t.Fatalf("unsafe contract became accepted: %v", err)
			}
			t.Logf("%s", err)
		})
	}
}
