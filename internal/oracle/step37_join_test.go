package oracle

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestStep37JoinRefusesUnprovenGraph(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step37/cyclic.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || len(want.stderr) != 0 || !bytes.Contains(want.stdout, []byte("alpha:11")) {
		t.Fatalf("bad mutable cyclic Node witness: %+v", want)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"private graph", "wrapped graph", "worker keeps reference"} {
		t.Run(variant, func(t *testing.T) {
			input := path
			if variant == "wrapped graph" {
				text := strings.Replace(string(source), "parallelMap(files, parseSourceElements)", "parallelMap(files, (text: string, index: number) => ({ file: parseSourceElements(text, index) }))", 1)
				text = strings.Replace(text, "for (const file of parsed) {", "for (const wrapper of parsed) { const file = wrapper.file;", 1)
				input = filepath.Join(t.TempDir(), "wrapped.a")
				if err := os.WriteFile(input, []byte(text), 0o644); err != nil {
					t.Fatal(err)
				}
				if diff := disagreement(want, onNode(t, input)); diff != "" {
					t.Fatal(diff)
				}
			}
			if variant == "worker keeps reference" {
				text := strings.Replace(string(source), "function identifierPart", "let kept: Declaration | undefined = undefined;\nfunction identifierPart", 1)
				text = strings.Replace(text, "node.parent = file;", "node.parent = file; kept = node;", 1)
				input = filepath.Join(t.TempDir(), "kept.a")
				if err := os.WriteFile(input, []byte(text), 0o644); err != nil {
					t.Fatal(err)
				}
				if diff := disagreement(want, onNode(t, input)); diff != "" {
					t.Fatal(diff)
				}
			}
			for _, query := range []bool{false, true} {
				checked, err := load.Load([]string{input})
				if err != nil {
					t.Fatal(err)
				}
				program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: query})
				var refused *lower.Refused
				if program != nil || !errors.As(err, &refused) {
					t.Fatalf("graph handoff reached emission: %v", err)
				}
				if variant != "worker keeps reference" {
					if refused.What != "cannot move work.result: whole graph ownership at join is not proven" {
						t.Fatalf("wrong boundary: %v", err)
					}
				} else if !strings.Contains(refused.What, "work writes state another task can reach") {
					t.Fatalf("worker survivor was not caught by the task effect proof: %v", err)
				}
				t.Logf("query=%t: %v", query, err)
			}
		})
	}
}

// No cyclic transfer is admitted here. These controls exercise the supported
// mutable acyclic result and parent-side pool merge at every requested pool size.
func TestStep37JoinControlsAndMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, before, after string }{
		{"step37_bind.a", "node.value += 10;", "node.value += 0;"},
		{"step37_intern_merge.a", "const merged = new Map<string, number>();", "pools.reverse();\nconst merged = new Map<string, number>();"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted", test.name))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), test.before) != 1 {
				t.Fatal("mutant lost its source anchor")
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stderr) != 0 || len(want.stdout) == 0 {
				t.Fatalf("bad Node witness: %+v", want)
			}
			mutant := filepath.Join(t.TempDir(), test.name)
			if err := os.WriteFile(mutant, []byte(strings.Replace(string(source), test.before, test.after, 1)), 0o644); err != nil {
				t.Fatal(err)
			}
			wrong := onNode(t, mutant)
			if wrong.exitCode != 0 || len(wrong.stderr) != 0 || bytes.Equal(wrong.stdout, want.stdout) {
				t.Fatal("mutant did not finish and change Node stdout")
			}
			for _, side := range []struct {
				name, path string
				want       run
			}{{"control", path, want}, {"mutant", mutant, wrong}} {
				t.Run(side.name, func(t *testing.T) {
					program, err := lowered(t, side.path)
					if err != nil {
						t.Fatal(err)
					}
					if !usesParallelMap(program) || len(program.GraphTypes) != 0 {
						t.Fatal("control must use acyclic parallel results")
					}
					if diff := disagreement(side.want, onJavaScriptBackend(t, program)); diff != "" {
						t.Fatal(diff)
					}
					code := native.C(program)
					binary := filepath.Join(t.TempDir(), "sanitized")
					if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
						t.Fatal(err)
					}
					for _, threads := range []string{"1", "4", "16"} {
						got := executeParallel(t, threads, false, binary)
						if diff := disagreement(side.want, got); diff != "" {
							t.Fatalf("threads=%s: %s: %s", threads, diff, got.stderr)
						}
						report, err := leakcheck.Check(leakcheck.Program{
							C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
							Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
								observed := executeWith(t, append([]string{"ADAMIC_THREADS=" + threads}, environment...), name, arguments...)
								return leakcheck.Run{Stdout: observed.stdout, Stderr: observed.stderr, ExitCode: observed.exitCode}
							},
						})
						if err != nil || report != "" {
							t.Fatalf("threads=%s leak check: %v %s", threads, err, report)
						}
						if side.name == "mutant" && disagreement(want, got) != "stdout differs" {
							t.Fatal("mutant survived pristine Node comparison")
						}
					}
					if side.name == "control" && runtime.GOOS == "linux" {
						race := filepath.Join(t.TempDir(), "tsan")
						if err := native.Build(code, race, native.Options{ThreadSanitize: true}); err != nil {
							t.Fatal(err)
						}
						for _, threads := range []string{"1", "4", "16"} {
							for attempt := 0; attempt < 3; attempt++ {
								got := executeParallel(t, threads, false, race)
								if diff := disagreement(want, got); diff != "" {
									t.Fatalf("TSan threads=%s: %s: %s", threads, diff, got.stderr)
								}
							}
						}
					}
					t.Log("Node and JavaScript agree; native leak-clean at 1, 4, 16 threads; pristine comparison rejects semantic mutants")
				})
			}
		})
	}
}
