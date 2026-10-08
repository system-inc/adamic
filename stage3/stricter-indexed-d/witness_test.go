package indexed_d

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

type site struct {
	ID, File, Expression, Cause, Read, Receiver, Source, Present, Absent string
	Refusal, NullPresent, Hole                                           string
	Checks                                                               int
	WantPresent, WantAbsent, WantNull                                    string
	Line                                                                 int
	Blocked                                                              bool
}

type result struct {
	stdout, stderr string
	code           int
}

func run(command string, arguments ...string) result {
	cmd := exec.Command(command, arguments...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exited, ok := err.(*exec.ExitError); ok {
			code = exited.ExitCode()
		} else {
			code = -1
			stderr.WriteString(err.Error())
		}
	}
	return result{stdout.String(), stderr.String(), code}
}

func write(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerWitnesses(t *testing.T) {
	data, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	var sites []site
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "adamic")
	build := run("go", "build", "-o", cli, "../../cmd/adamic")
	if build.code != 0 {
		t.Fatalf("CLI build: %+v", build)
	}
	for _, s := range sites {
		t.Run(s.ID, func(t *testing.T) {
			variants := []string{"present", "absent", "hole"}
			if s.Checks == 2 {
				variants = append(variants, "outer-present", "outer-absent")
			}
			if s.NullPresent != "" {
				variants = append(variants, "null")
			}
			for _, name := range variants {
				absent := name == "absent" || name == "hole" || name == "outer-absent"
				current := s
				if strings.HasPrefix(name, "outer-") {
					current.Source = strings.Replace(s.Source, "const value: string = options.paths[key][i];", "const value: string[] = options.paths[key];", 1)
					current.Source = strings.Replace(current.Source, "console.log(`${value}`);", `console.log(value === undefined ? "undefined" : "7");`, 1)
					current.Read, current.Checks = "options.paths[key]", 1
					current.Absent = "{}"
				}
				s := current
				values, want := s.Present, "7\n"
				if absent {
					values, want = s.Absent, "undefined\n"
				}
				if name == "null" {
					values, want = s.NullPresent, "null\n"
				}
				if name == "present" && s.WantPresent != "" {
					want = s.WantPresent
				}
				if absent && s.WantAbsent != "" {
					want = s.WantAbsent
				}
				if name == "null" && s.WantNull != "" {
					want = s.WantNull
				}
				if name == "hole" {
					if s.Blocked {
						continue
					}
					element := "Item"
					if strings.Contains(s.Source, "value: string") {
						element = "string"
					}
					if strings.Contains(s.Source, "value: number") {
						element = "number"
					}
					values = "new Array<" + element + ">(1)"
					if s.Hole != "" {
						values = s.Hole
					}
				}
				t.Run(name, func(t *testing.T) {
					directory := t.TempDir()
					path := filepath.Join(directory, s.ID+".ts")
					source := strings.ReplaceAll(s.Source, "@VALUES@", values)
					write(t, path, source)
					write(t, filepath.Join(directory, "tsconfig.json"), fmt.Sprintf(`{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":[%q]}`, filepath.Base(path)))
					node := run("node", path)
					if node != (result{stdout: want}) {
						t.Fatalf("source Node: %+v, want %q", node, want)
					}
					checked, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if s.Blocked {
						refusal := s.Refusal
						if refusal == "" {
							refusal = "Adamic 0.1 refuses an index signature; use a Map"
						}
						if err == nil || !strings.Contains(err.Error(), refusal) {
							t.Fatalf("expected representation refusal %q: %v", refusal, err)
						}
						t.Logf("BLOCKED %s %s: %v; Node %q", s.ID, s.Receiver, err, node.stdout)
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					checks := ir.InsertedChecks(program)
					count := s.Checks
					if count == 0 {
						count = 1
					}
					if len(checks) != count {
						t.Fatalf("want individually observable indexed guards: %+v", checks)
					}
					// Compute the source position independently of the compiler's check inventory.
					offset := strings.Index(source, s.Read)
					if offset < 0 {
						t.Fatal("manifest read missing from source")
					}
					before := source[:offset]
					line := strings.Count(before, "\n") + 1
					column := len(before) - strings.LastIndex(before, "\n")
					where := fmt.Sprintf("%s:%d:%d", path, line, column)
					for _, check := range checks {
						if check.Where != where || check.Kind != "indexed-presence" {
							t.Fatalf("site: %+v, want %q indexed-presence", check, where)
						}
					}
					explain := run(cli, "--explain-checks", path)
					if explain.code != 0 || strings.Count(explain.stdout+explain.stderr, fmt.Sprintf("%s:%d:%d: checked indexed-presence\n", filepath.Base(path), line, column)) != count || !strings.Contains(explain.stdout+explain.stderr, fmt.Sprintf("checked: indexed-presence=%d", count)) || !strings.Contains(explain.stdout+explain.stderr, "trusted: 0") {
						t.Fatalf("explain: %+v", explain)
					}
					expected := node
					if absent {
						expected = result{stderr: "adamic: panic: indexed read is absent: " + where + "\n", code: 70}
					}
					js := javascript.JavaScript(program)
					module, err := filepath.Abs("../../oracle/adamic.mjs")
					if err != nil {
						t.Fatal(err)
					}
					js = strings.Replace(js, "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
					jsPath := filepath.Join(directory, "backend.mjs")
					write(t, jsPath, js)
					if got := run("node", jsPath); got != expected {
						t.Fatalf("JS: %+v, want %+v", got, expected)
					}
					c := native.C(program)
					for _, sanitize := range []bool{false, true} {
						binary := filepath.Join(directory, fmt.Sprintf("native-%t", sanitize))
						if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
							t.Fatal(err)
						}
						if got := run(binary); got != expected {
							t.Fatalf("native sanitize=%t: %+v, want %+v", sanitize, got, expected)
						}
					}
					if absent {
						// Erase only this site's panic, preserving its lookup and all other code.
						panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
						matches := panicCall.FindAllStringIndex(c, -1)
						if len(matches) != count {
							t.Fatal("mutant must erase exactly one guard")
						}
						binary := filepath.Join(directory, "mutant")
						// Chained reads evaluate the outer guard before the inner guard.
						selected := matches[len(matches)-1]
						mutant := c[:selected[0]] + "(void)0;" + c[selected[1]:]
						if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
							t.Fatalf("mutant build is not a kill: %v", err)
						}
						got := run(binary)
						if got == expected {
							t.Fatal("erased guard survived")
						}
						t.Logf("PROVEN %s read=%s; release/sanitized and JS exit 70 with exact stderr; erase-panic mutant caught: exit %d stdout=%q stderr=%q", s.ID, s.Read, got.code, got.stdout, got.stderr)
					} else {
						if name == "null" {
							binary := nullableSentinelMutant(t, c)
							got := run(binary)
							if got == expected {
								t.Fatal("sentinel conflation mutant survived")
							}
							t.Logf("PROVEN %s null sentinel mutant caught: exit %d stdout=%q stderr=%q", s.ID, got.code, got.stdout, got.stderr)
						}
						t.Logf("PRESENT %s matches source Node in release/sanitized and JS", s.ID)
					}
				})
			}
		})
	}
}
