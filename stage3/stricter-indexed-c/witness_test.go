package indexedc

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
	Lookups int    `json:"lookups"`
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

func TestLedgerWitnesses(t *testing.T) {
	data, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	var sites []site
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	if len(sites) != 18 {
		t.Fatalf("want 18 ledger rows, got %d", len(sites))
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
					if mode != "present" {
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
					if len(checks) != 1 || checks[0] != (ir.InsertedCheck{Kind: "indexed-presence", Where: where}) {
						t.Fatalf("site must be individually checked: %+v, want %s", checks, where)
					}
					explain := run(t, cli, "--explain-checks", path)
					expectedExplain := where + ": checked indexed-presence\n" + "checked: indexed-presence=1 catch-error=0 json-stringify-defined=0 optional-write=0\ntrusted: 0\n"
					if explain.code != 0 || explain.stdout+explain.stderr != expectedExplain {
						t.Fatalf("explain: %+v", explain)
					}
					want := observation{s.Want, "", 0}
					if mode == "absent" {
						want = observation{"", "adamic: panic: indexed read is absent: " + where + "\n", 70}
					}
					c := native.C(program)
					lookups := s.Lookups
					if lookups == 0 {
						lookups = 1
					}
					if strings.Count(c, "adamic_array_at(") != lookups {
						t.Fatal("guard must retain exactly one indexed lookup")
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
						if mode == "absent" {
							// Erase only the single emitted presence panic, retaining its lookup.
							panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
							if len(panicCall.FindAllString(c, -1)) != 1 {
								t.Fatal("mutant requires exactly one guard")
							}
							mutant := binary + "-mutant"
							if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), mutant, native.Options{Sanitize: sanitized}); err != nil {
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
