package records

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
	ID      string `json:"id"`
	Typed   string `json:"typed"`
	Present string `json:"present"`
	Absent  string `json:"absent"`
	Read    string `json:"read"`
	Want    string `json:"want"`
	Blocked string `json:"blocked"`
	Checks  int    `json:"checks"`
	Arrays  int    `json:"arrays"`
}

type observation struct {
	stdout, stderr string
	code           int
}

func run(t *testing.T, command string, args ...string) observation {
	t.Helper()
	cmd := exec.Command(command, args...)
	var out, errout bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errout
	err := cmd.Run()
	code := 0
	if err != nil {
		exited, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exited.ExitCode()
	}
	return observation{out.String(), errout.String(), code}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

// Not parallel: builds one CLI and executes compiler/backend mutants sequentially.
func TestLedgerWitnesses(t *testing.T) {
	data, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	var sites []site
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	if len(sites) != 10 {
		t.Fatalf("want 10 record shapes, got %d", len(sites))
	}
	cli := os.Getenv("ADAMIC_WITNESS_CLI")
	if cli == "" {
		cli = filepath.Join(t.TempDir(), "adamic")
		if built := run(t, "go", "build", "-o", cli, "../../cmd/adamic"); built.code != 0 {
			t.Fatalf("build current CLI: %+v", built)
		}
	}
	for _, s := range sites {
		t.Run(s.ID, func(t *testing.T) {
			cases := []string{"present", "absent"}
			if s.Blocked != "" {
				cases = []string{"refused"}
			}
			for _, mode := range cases {
				t.Run(mode, func(t *testing.T) {
					directory := t.TempDir()
					input := s.Present
					wantNode := s.Want
					if mode != "present" && s.ID != "prototype-own-key" {
						input, wantNode = s.Absent, "undefined\n"
					}
					source := strings.ReplaceAll(s.Typed, "INPUT", input)
					path := filepath.Join(directory, "main.ts")
					write(t, path, source)
					// These are TypeScript interoperability probes, materialized as .ts under
					// their own project, not new Adamic programs with erased type promises.
					write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`)
					node := run(t, "node", path)
					if node != (observation{wantNode, "", 0}) {
						t.Fatalf("source Node: %+v", node)
					}
					checked, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if s.Blocked != "" {
						if err == nil || !strings.Contains(err.Error(), s.Blocked) {
							t.Fatalf("want refusal %q, got %v", s.Blocked, err)
						}
						t.Logf("Node gives undefined; observed refusal: %v", err)
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					offset := strings.Index(source, s.Read)
					if offset < 0 {
						t.Fatal("read missing from witness")
					}
					before := source[:offset]
					line := strings.Count(before, "\n") + 1
					column := len(before) - strings.LastIndex(before, "\n")
					where := fmt.Sprintf("%s:%d:%d", path, line, column)
					checks := ir.InsertedChecks(program)
					count := s.Checks
					if count == 0 {
						count = 1
					}
					if s.ID == "D129-D131" {
						count = 2
					}
					if len(checks) != count {
						t.Fatalf("site must be individually checked: %+v, want %s", checks, where)
					}
					for _, check := range checks {
						if check != (ir.InsertedCheck{Kind: "indexed-presence", Where: where}) {
							t.Fatalf("site: %+v, want %s", check, where)
						}
					}
					explain := run(t, cli, "--explain-checks", path)
					expectedExplain := ""
					for _, check := range checks {
						expectedExplain += check.Where + ": checked indexed-presence\n"
					}
					expectedExplain += fmt.Sprintf("checked: indexed-presence=%d catch-error=0 json-stringify-defined=0 optional-write=0\ntrusted: 0\n", count)
					if explain.code != 0 || explain.stdout+explain.stderr != expectedExplain {
						t.Fatalf("explain: %+v", explain)
					}
					want := observation{s.Want, "", 0}
					if mode == "absent" && s.ID != "prototype-own-key" {
						want = observation{"", "adamic: panic: indexed read is absent: " + where + "\n", 70}
					}
					c := native.C(program)
					arrays := s.Arrays
					if s.ID == "D129-D131" {
						arrays = 1
					}
					if strings.Count(c, "adamic_record_get(") != count-arrays || strings.Count(c, "adamic_array_at(") != arrays {
						t.Fatal("each guard must retain exactly one indexed lookup")
					}
					js := javascript.JavaScript(program)
					runtime, err := filepath.Abs("../../oracle/adamic.mjs")
					if err != nil {
						t.Fatal(err)
					}
					js = strings.Replace(js, "from 'adamic'", "from 'file://"+filepath.ToSlash(runtime)+"'", 1)
					backend := filepath.Join(directory, "backend.mjs")
					write(t, backend, js)
					if got := run(t, "node", backend); got != want {
						t.Fatalf("backend Node: %+v, want %+v", got, want)
					}
					for _, sanitized := range []bool{false, true} {
						binary := filepath.Join(directory, fmt.Sprintf("native-%t", sanitized))
						if err := native.Build(c, binary, native.Options{Sanitize: sanitized}); err != nil {
							t.Fatal(err)
						}
						if got := run(t, binary); got != want {
							t.Fatalf("native sanitized=%t: %+v, want %+v", sanitized, got, want)
						}

						if mode == "present" && sanitized {
							counted := binary + "-counted"
							if err := native.Build(c, counted, native.Options{Sanitize: true, Count: true}); err != nil {
								t.Fatal(err)
							}
							got := run(t, counted)
							counts := regexp.MustCompile(`^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`).FindStringSubmatch(got.stderr)
							if got.code != 0 || got.stdout != want.stdout || counts == nil || counts[1] != counts[2] || counts[6] != "0" {
								t.Fatalf("unbalanced counted record: %+v", got)
							}
							t.Log(strings.TrimSpace(got.stderr))
						}
						if mode == "absent" && s.ID != "prototype-own-key" {
							// Erase only the single emitted presence panic, retaining its lookup.
							panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
							if len(panicCall.FindAllString(c, -1)) != count {
								t.Fatal("mutant requires one guard per indexed read")
							}
							mutant := binary + "-mutant"
							if err := native.Build(eraseLastGuard(c, panicCall), mutant, native.Options{Sanitize: sanitized}); err != nil {
								t.Fatalf("mutant must build: %v", err)
							}
							got := run(t, mutant)
							if got == want {
								t.Fatal("erase-check mutant survived")
							}
							t.Logf("erase %s guard caught by pinned stdout/stderr/exit assertion; sanitized=%t, mutant exit=%d stdout=%q", s.ID, sanitized, got.code, got.stdout)
						}
					}
					t.Logf("%s: source Node matched; both backends, release and sanitized native, exact named stderr and explain site passed", mode)
				})
			}
		})
	}
}

// Each chained read has its own guard. The final prefix guard emits last;
// erasing it retains all earlier record guards and every indexed lookup.
func eraseLastGuard(c string, re *regexp.Regexp) string {
	matches := re.FindAllStringIndex(c, -1)
	last := matches[len(matches)-1]
	return c[:last[0]] + "(void)0;" + c[last[1]:]
}
