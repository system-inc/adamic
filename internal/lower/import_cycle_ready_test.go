package lower

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// undecidedCycleRule models no finding from a rule that cannot follow the load-time
// call. It is a test seam, not a replacement for cohere's public runner.
type undecidedCycleRule struct{}

func (undecidedCycleRule) RunRule(*compiler.Program, *checker.Checker, []*ast.SourceFile, string) ([]Finding, error) {
	return nil, nil
}

func TestUndecidedCycleReadsUseReadyChecks(t *testing.T) {
	// Not parallel: replace the process-wide rule seam, restoring it before parallel tests run.
	previous := loadTimeReadRunner
	loadTimeReadRunner = undecidedCycleRule{}
	defer func() { loadTimeReadRunner = previous }()
	for _, probe := range []struct {
		name, a, b, entry, output string
		fails                     bool
	}{
		{"early value", "import './b.a'; export const value = 1;", "import { value } from './a.a'; console.log(`${value}`);", "a.a", "adamic: panic: ReferenceError: Cannot access 'value' before initialization\n", true},
		{"initialized value", "import './b.a'; export const value = 1;", "import { value } from './a.a'; console.log(`${value}`);", "b.a", "1\n", false},
		{"indirect value", "import { value } from './b.a'; function read(): number { return value; } console.log(`${read()}`);", "import './a.a'; export const value = 2;", "b.a", "adamic: panic: ReferenceError: Cannot access 'value' before initialization\n", true},
		{"early plain class", "import './b.a'; export class Box { readonly value = 7; } export function make(): number { return new Box().value; }", "import { make } from './a.a'; console.log(`${make()}`);", "a.a", "adamic: panic: ReferenceError: Cannot access 'Box' before initialization\n", true},
		{"initialized plain class", "import './b.a'; export class Box { readonly value = 7; } export function make(): number { return new Box().value; }", "import { make } from './a.a'; console.log(`${make()}`);", "b.a", "7\n", false},
		{"early extends", "import './b.a'; export class Base {}", "import { Base } from './a.a'; class Child extends Base {} console.log('child');", "a.a", "adamic: panic: ReferenceError: Cannot access 'Base' before initialization\n", true},
		{"initialized extends", "import './b.a'; export class Base {}", "import { Base } from './a.a'; class Child extends Base {} console.log('child');", "b.a", "child\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			directory := t.TempDir()
			for name, source := range map[string]string{"a.a": probe.a, "b.a": probe.b} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, probe.entry)
			// Observe a computed answer before the cycle can panic, rather than
			// letting a failing cycle pass the oracle on silence.
			if err := os.WriteFile(filepath.Join(directory, "probe.a"), []byte("console.log(`${1 + 2}`);"), 0o644); err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(entry)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(entry, append([]byte("import './probe.a';\n"), original...), 0o644); err != nil {
				t.Fatal(err)
			}
			probe.output = "3\n" + probe.output
			if probe.fails {
				probe.output = "3\n"
			}
			loaded, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			if probe.name == "initialized value" && strings.Contains(native.C(program), "ReferenceError: Cannot access 'value'") {
				t.Error("proven imported read retained a TDZ check")
			}
			if probe.name == "initialized extends" && strings.Contains(native.C(program), "ReferenceError: Cannot access 'Base'") {
				t.Error("proven base class read retained a TDZ check")
			}
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			generated := filepath.Join(directory, "generated.mjs")
			if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				t.Fatal(err)
			}
			run := func(command *exec.Cmd) {
				t.Helper()
				var stderr bytes.Buffer
				command.Stderr = &stderr
				output, err := command.Output()
				if (command.Path == binary || !probe.fails) && stderr.Len() != 0 {
					t.Errorf("%s: unexpected stderr %s", command.Path, stderr.Bytes())
				}
				if string(output) != probe.output {
					t.Errorf("%s: got %q, want %q (error %v)", command.Path, output, probe.output, err)
				}
				if probe.fails {
					exit, ok := err.(*exec.ExitError)
					if !ok || exit.ExitCode() != 1 {
						t.Errorf("expected exit 1, got %v", err)
					}
				} else if err != nil {
					t.Error(err)
				}
			}
			// Node reads the original .a modules with native ESM evaluation; its runtime
			// has the same stdout and exit 1 contract as Adamic for uncaught errors.
			// Step 21 excludes engine-specific stderr rendering from that contract.
			runner := filepath.Join("..", "..", "oracle", "node.mjs")
			for _, path := range []string{entry, generated} {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				arguments := []string{"--disable-warning=ExperimentalWarning", runner}
				command := exec.CommandContext(ctx, "node", append(arguments, path)...)
				run(command)
				cancel()
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, binary)
			if probe.fails {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			}
			run(command)
		})
	}
}
