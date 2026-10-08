package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type observation struct {
	stdout, stderr string
	code           int
}

func run(command string, args ...string) observation {
	cmd := exec.Command(command, args...)
	var out, errout bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errout
	err := cmd.Run()
	code := 0
	if err != nil {
		if exited, ok := err.(*exec.ExitError); ok {
			code = exited.ExitCode()
		} else {
			code = -1
			errout.WriteString(err.Error())
		}
	}
	return observation{out.String(), errout.String(), code}
}

// Each row is isolated so a preceding failed check cannot hide this site's check.
// Fixture text is materialized as .ts under its own project, as in the base worker.
func TestGeneratorWitnesses(t *testing.T) {
	t.Parallel()
	cli := filepath.Join(t.TempDir(), "adamic")
	if result := run("go", "build", "-o", cli, "../../cmd/adamic"); result.code != 0 {
		t.Fatalf("CLI build: %+v", result)
	}
	var sites []struct {
		ID string `json:"id"`
	}
	data, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	if len(sites) != 27 {
		t.Fatalf("want 27 ledger rows, got %d", len(sites))
	}
	for index, site := range sites {
		if site.ID != fmt.Sprintf("D%d", 195+index) {
			t.Fatalf("ledger identity/order at %d: %s", index, site.ID)
		}
		t.Run(site.ID, func(t *testing.T) {
			text, err := os.ReadFile(filepath.Join("fixtures", site.ID+".ts.txt"))
			if err != nil {
				t.Fatal(err)
			}
			for _, absent := range []bool{false, true} {
				t.Run(fmt.Sprintf("absent=%t", absent), func(t *testing.T) {
					directory := t.TempDir()
					path := filepath.Join(directory, "main.ts")
					source := strings.ReplaceAll(string(text), "ABSENT", fmt.Sprintf("%t", absent))
					if err := os.WriteFile(path, []byte(source), 0644); err != nil {
						t.Fatal(err)
					}
					config := `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`
					if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0644); err != nil {
						t.Fatal(err)
					}
					want := "present\n"
					if absent {
						want = "undefined\n"
					}
					node := run("node", "--disable-warning=ExperimentalWarning", path)
					if node.code != 0 || node.stdout != want || node.stderr != "" {
						t.Fatalf("source Node: %+v, want %q", node, want)
					}
					checked, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if err != nil {

						t.Fatalf("unexpected lowering failure for %s: %v", site.ID, err)
					}
					checks := ir.InsertedChecks(program)
					if len(checks) != 1 || checks[0].Kind != "indexed-presence" {
						t.Fatalf("explain actual guards: %+v", checks)
					}
					explain := run(cli, "--explain-checks", path)
					if explain.code != 0 || explain.stderr != "" || strings.Count(explain.stdout, ": checked indexed-presence\n") != 1 || !strings.Contains(explain.stdout, "checked: indexed-presence=1 catch-error=0 json-stringify-defined=0 optional-write=0\ntrusted: 0\n") {
						t.Fatalf("CLI explain: %+v", explain)
					}
					c := native.C(program)
					module, err := filepath.Abs("../../oracle/adamic.mjs")
					if err != nil {
						t.Fatal(err)
					}
					js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
					jsPath := filepath.Join(directory, "backend.mjs")
					if err := os.WriteFile(jsPath, []byte(js), 0644); err != nil {
						t.Fatal(err)
					}
					backend := run("node", jsPath)
					var expected observation
					for index, line := range strings.Split(source, "\n") {
						if strings.HasPrefix(line, "const value:") {
							where := fmt.Sprintf("%s:%d:%d", path, index+1, strings.Index(line, " = ")+4)
							if checks[0].Where != where {
								t.Fatalf("check at %q, want %q", checks[0].Where, where)
							}
							expected = observation{"", "adamic: panic: indexed read is absent: " + where + "\n", 70}
						}
					}
					for _, sanitize := range []bool{false, true} {
						binary := filepath.Join(directory, fmt.Sprintf("native-%t", sanitize))
						if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
							t.Fatal(err)
						}
						actual := run(binary)
						if absent {
							if actual != expected {
								t.Fatalf("native sanitize=%t: %+v, want %+v", sanitize, actual, expected)
							}
						} else if actual != node {
							t.Fatalf("native sanitize=%t: %+v, Node %+v", sanitize, actual, node)
						}
					}
					if absent {
						if backend != expected {
							t.Fatalf("JS %+v, native %+v", backend, expected)
						}
						// Erase this one emitted panic, retaining the lookup. Successful compilation
						// is required; a warning or compiler error is never counted as a killed mutant.
						panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
						if len(panicCall.FindAllString(c, -1)) != 1 {
							t.Fatal("want one panic to erase")
						}
						mutant := filepath.Join(directory, "mutant")
						if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), mutant, native.Options{Sanitize: true}); err != nil {
							t.Fatalf("mutant build: %v", err)
						}
						mutated := run(mutant)
						if mutated == expected {
							t.Fatal("erase-check mutant survived")
						}
						t.Logf("PROVEN %s: pinned stderr=%q; erase-check mutant exit=%d", site.ID, expected.stderr, mutated.code)
					} else if backend != node {
						t.Fatalf("JS %+v, Node %+v", backend, node)
					}
				})
			}
		})
	}
}
