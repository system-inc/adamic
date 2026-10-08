package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Re-run every passing Error/NativeErrors/AggregateError test from a scoped
// runner report through all backends, preserving the runner's exact adaptation.
func TestErrorTest262Backends(t *testing.T) {
	checkout, report := os.Getenv("ADAMIC_ERROR_TEST262"), os.Getenv("ADAMIC_ERROR_TEST262_REPORT")
	if checkout == "" || report == "" {
		t.Skip("set ADAMIC_ERROR_TEST262 and ADAMIC_ERROR_TEST262_REPORT to the pinned checkout and scoped JSON report")
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Filters []struct {
			Path   string
			Pass   int
			Passes []string
		}
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, filter := range document.Filters {
		if filter.Path != "built-ins/Error" && filter.Path != "built-ins/NativeErrors" && filter.Path != "built-ins/AggregateError" {
			t.Fatalf("unscoped filter %s", filter.Path)
		}
		if len(filter.Passes) != filter.Pass {
			t.Fatal("report pass list/count disagree")
		}
		for _, path := range filter.Passes {
			total++
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join(checkout, "test", path))
				if err != nil {
					t.Fatal(err)
				}
				test := classify(path, string(source), true)
				if test.Skip != "" {
					t.Fatal(test.Skip)
				}
				directory := t.TempDir()
				input := filepath.Join(directory, "program.a")
				if err := os.WriteFile(input, []byte(test.Program), 0644); err != nil {
					t.Fatal(err)
				}
				loaded, err := load.Load([]string{input})
				if err != nil {
					t.Fatal(err)
				}
				program, err := lower.Lower(context.Background(), loaded)
				if err != nil {
					t.Fatal(err)
				}
				observe := func(name string, args ...string) (stdout, stderr []byte, exit int) {
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					command := exec.CommandContext(ctx, name, args...)
					var out, errors bytes.Buffer
					command.Stdout = &out
					command.Stderr = &errors
					if err := command.Run(); err != nil {
						if failure, ok := err.(*exec.ExitError); ok {
							exit = failure.ExitCode()
						} else {
							t.Fatal(err)
						}
					}
					if ctx.Err() != nil {
						t.Fatal(ctx.Err())
					}
					return out.Bytes(), errors.Bytes(), exit
				}
				module := filepath.Join(directory, "program.mts")
				if err := os.WriteFile(module, []byte(test.Program), 0644); err != nil {
					t.Fatal(err)
				}
				expectedOut, expectedErr, expectedExit := observe("node", "--disable-warning=ExperimentalWarning", "--experimental-strip-types", module)
				if expectedExit != 0 || len(expectedErr) != 0 {
					t.Fatalf("Node pass must be clean: %d %s", expectedExit, expectedErr)
				}
				js := filepath.Join(directory, "program.mjs")
				if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0644); err != nil {
					t.Fatal(err)
				}
				binary := filepath.Join(directory, "program")
				if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				wasm := filepath.Join(directory, "program.wasm")
				if err := native.Build(native.C(program), wasm, native.Options{Target: "wasm32-wasi"}); err != nil {
					t.Fatal(err)
				}
				for _, backend := range []struct {
					name    string
					command string
					args    []string
				}{{"native", binary, nil}, {"JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js}}, {"WASI", "node", []string{"--disable-warning=ExperimentalWarning", "../../oracle/wasi.mjs", wasm}}} {
					out, errors, exit := observe(backend.command, backend.args...)
					if !bytes.Equal(out, expectedOut) || !bytes.Equal(errors, expectedErr) || exit != expectedExit {
						t.Fatalf("%s full comparison differs: Node stdout %q stderr %q exit %d; backend stdout %q stderr %q exit %d", backend.name, expectedOut, expectedErr, expectedExit, out, errors, exit)
					}
					t.Log(backend.name + ": full stdout/stderr/exit agree with Node")
				}
			})
		}
	}
	if total == 0 {
		t.Fatal("empty pass set")
	}
	t.Logf("selected %d Error tests for all backend comparisons; %s", total, strings.TrimSpace(checkout))
}
