package indexedwitness

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

type witness struct {
	IDs        []string `json:"ids"`
	Read       string   `json:"read"`
	Reads      []string `json:"reads"`
	Receiver   string   `json:"receiver"`
	Source     string   `json:"source"`
	BlockStage string   `json:"block_stage"`
	Block      string   `json:"block"`
}

type observation struct {
	stdout, stderr string
	code           int
}

func run(command string, arguments ...string) observation {
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
	return observation{stdout.String(), stderr.String(), code}
}
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// Each manifest preserves the ledger receiver and index form. Project .ts files
// are generated inputs, rather than new Adamic programs claiming proven types.
func TestCheckerIndexedWitnesses(t *testing.T) {
	cli := filepath.Join(t.TempDir(), "adamic")
	if got := run("go", "build", "-o", cli, "../../cmd/adamic"); got.code != 0 {
		t.Fatalf("CLI build: %+v", got)
	}
	paths, err := filepath.Glob("D*.json")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "gaps/destructuring-third.json")
	for _, manifest := range paths {
		t.Run(strings.TrimSuffix(manifest, ".json"), func(t *testing.T) {
			data, err := os.ReadFile(manifest)
			if err != nil {
				t.Fatal(err)
			}
			var fixture witness
			if err := json.Unmarshal(data, &fixture); err != nil {
				t.Fatal(err)
			}
			for _, absent := range []bool{false, true} {
				t.Run(fmt.Sprintf("absent=%v", absent), func(t *testing.T) {
					directory := t.TempDir()
					path := filepath.Join(directory, "main.ts")
					source := strings.ReplaceAll(fixture.Source, "ABSENT", fmt.Sprint(absent))
					write(t, path, source)
					write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`)
					want := observation{"7\n", "", 0}
					if absent {
						want.stdout = "undefined\n"
					}
					node := run("node", path)
					if node != want {
						t.Fatalf("source Node: %+v, want %+v", node, want)
					}
					checked, err := load.Load([]string{path})
					if fixture.BlockStage == "load" {
						assertRefusal(t, cli, path, fixture, err)
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if fixture.BlockStage == "lower" {
						assertRefusal(t, cli, path, fixture, err)
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					checks := ir.InsertedChecks(program)
					reads := fixture.Reads
					if len(reads) == 0 {
						reads = []string{fixture.Read}
					}
					locations := make([]string, len(reads))
					where := ""
					for index, read := range reads {
						position := strings.Index(source, read)
						if position < 0 {
							t.Fatal("read missing")
						}
						prefix := source[:position]
						line := strings.Count(prefix, "\n") + 1
						column := position - strings.LastIndex(prefix, "\n")
						locations[index] = fmt.Sprintf("%s:%d:%d", path, line, column)
						if read == fixture.Read {
							where = locations[index]
						}
					}
					if where == "" || len(checks) != len(reads) {
						t.Fatalf("expected binding checks %v: %+v", locations, checks)
					}
					for index, location := range locations {
						if checks[index].Kind != "indexed-presence" || checks[index].Where != location {
							t.Fatalf("read %s: %+v", location, checks)
						}
					}
					explain := run(cli, "--explain-checks", path)
					if explain.code != 0 || explain.stderr != "" || !strings.Contains(explain.stdout, fmt.Sprintf("checked: indexed-presence=%d", len(reads))) || !strings.Contains(explain.stdout, "trusted: 0\n") {
						t.Fatalf("CLI explain: %+v", explain)
					}
					for _, location := range locations {
						if !strings.Contains(explain.stdout, location+": checked indexed-presence\n") {
							t.Fatalf("CLI missed %s: %+v", location, explain)
						}
					}
					if absent {
						want = observation{"", "adamic: panic: indexed read is absent: " + where + "\n", 70}
					}
					c := native.C(program)
					for _, sanitize := range []bool{false, true} {
						binary := filepath.Join(directory, fmt.Sprintf("native-%v", sanitize))
						if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
							t.Fatal(err)
						}
						got := run(binary)
						if got != want {
							t.Fatalf("native sanitize=%v: %+v, want %+v", sanitize, got, want)
						}
					}
					module, err := filepath.Abs("../../oracle/adamic.mjs")
					if err != nil {
						t.Fatal(err)
					}
					js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
					jsPath := filepath.Join(directory, "backend.mjs")
					write(t, jsPath, js)
					if got := run("node", jsPath); got != want {
						t.Fatalf("JS backend: %+v, want %+v", got, want)
					}
					if absent {
						panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
						if len(panicCall.FindAllString(c, -1)) != len(reads) {
							t.Fatal("mutant must erase exactly one guard")
						}
						binary := filepath.Join(directory, "mutant")
						if err := native.Build(eraseSitePanic(t, c, panicCall, locations, where), binary, native.Options{Sanitize: true}); err != nil {
							t.Fatalf("mutant must build: %v", err)
						}
						got := run(binary)
						if got == want {
							t.Fatal("erase guard mutant survived")
						}
						t.Logf("%v receiver=%s Node undefined; named stderr pinned in release/sanitized native and JS; erased guard caught: exit=%d stdout=%q stderr=%q", fixture.IDs, fixture.Receiver, got.code, got.stdout, got.stderr)
					}
				})
			}
		})
	}
}

// A refusal is evidence of a gap, never evidence of a runtime check. If the gap
// closes, this assertion fails so the witness must gain the full runtime proof.
func assertRefusal(t *testing.T, cli, path string, fixture witness, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), fixture.Block) {
		t.Fatalf("expected %s refusal %q, got %v", fixture.BlockStage, fixture.Block, err)
	}
	explain := run(cli, "--explain-checks", path)
	if explain.code != 1 || !strings.Contains(explain.stderr, fixture.Block) || strings.Contains(explain.stdout, "checked indexed-presence") {
		t.Fatalf("refused witness must not count as checked: %+v", explain)
	}
	t.Logf("BLOCKED %v receiver=%s: %v; source Node observations passed; CLI refuses and does not count a check", fixture.IDs, fixture.Receiver, err)
}

// Erase only the selected binding's panic, retaining every other binding guard.
func eraseSitePanic(t *testing.T, c string, pattern *regexp.Regexp, locations []string, where string) string {
	t.Helper()
	matches := pattern.FindAllStringIndex(c, -1)
	for index, location := range locations {
		if location == where {
			match := matches[index]
			return c[:match[0]] + "(void)0;" + c[match[1]:]
		}
	}
	t.Fatal("mutant site missing")
	return c
}
