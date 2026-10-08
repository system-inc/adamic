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
	IDs      []string `json:"ids"`
	Read     string   `json:"read"`
	Receiver string   `json:"receiver"`
	Source   string   `json:"source"`
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
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if err != nil {
						t.Fatal(err)
					}
					checks := ir.InsertedChecks(program)
					position := strings.Index(source, fixture.Read)
					if position < 0 {
						t.Fatal("read missing")
					}
					prefix := source[:position]
					line := strings.Count(prefix, "\n") + 1
					column := position - strings.LastIndex(prefix, "\n")
					where := fmt.Sprintf("%s:%d:%d", path, line, column)
					if len(checks) != 1 || checks[0].Kind != "indexed-presence" || checks[0].Where != where {
						t.Fatalf("explain must list exactly this read %s: %+v", where, checks)
					}
					explain := run(cli, "--explain-checks", path)
					if explain.code != 0 || explain.stderr != "" || !strings.Contains(explain.stdout, where+": checked indexed-presence\n") || !strings.Contains(explain.stdout, "checked: indexed-presence=1") || !strings.Contains(explain.stdout, "trusted: 0\n") {
						t.Fatalf("CLI explain: %+v", explain)
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
						if len(panicCall.FindAllString(c, -1)) != 1 {
							t.Fatal("mutant must erase exactly one guard")
						}
						binary := filepath.Join(directory, "mutant")
						if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), binary, native.Options{Sanitize: true}); err != nil {
							t.Fatalf("mutant must build: %v", err)
						}
						got := run(binary)
						if got == want {
							t.Fatal("erase guard mutant survived")
						}
						t.Logf("%v receiver=%s Node undefined; named stderr pinned in release/sanitized native and JS; erased guard caught: exit=%d stdout=%q", fixture.IDs, fixture.Receiver, got.code, got.stdout)
					}
				})
			}
		})
	}
}
