package lower

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A cycle read the module order can't prove keeps its runtime dead-zone check, and one it proves loses it: lowering
// decides from the order alone, with no lint rule's findings to consult.
func TestUndecidedCycleReadsUseReadyChecks(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, a, b, entry, output string
		fails                     bool
	}{
		{
			"early value",
			"import './b.a'; export const value = 1;",
			"import { value } from './a.a'; console.log(`${value}`);",
			"a.a",
			"adamic: panic: ReferenceError: Cannot access 'value' before initialization\n",
			true,
		},
		{
			"initialized value",
			"import './b.a'; export const value = 1;",
			"import { value } from './a.a'; console.log(`${value}`);",
			"b.a",
			"1\n",
			false,
		},
		{
			"indirect value",
			"import { value } from './b.a'; function read(): number { return value; } console.log(`${read()}`);",
			"import './a.a'; export const value = 2;",
			"b.a",
			"adamic: panic: ReferenceError: Cannot access 'value' before initialization\n",
			true,
		},
		{
			"early plain class",
			"import './b.a'; export class Box { readonly value = 7; } export function make(): number { return new Box().value; }",
			"import { make } from './a.a'; console.log(`${make()}`);",
			"a.a",
			"adamic: panic: ReferenceError: Cannot access 'Box' before initialization\n",
			true,
		},
		{
			"initialized plain class",
			"import './b.a'; export class Box { readonly value = 7; } export function make(): number { return new Box().value; }",
			"import { make } from './a.a'; console.log(`${make()}`);",
			"b.a",
			"7\n",
			false,
		},
		{
			"early extends",
			"import './b.a'; export class Base {}",
			"import { Base } from './a.a'; class Child extends Base {} console.log('child');",
			"a.a",
			"adamic: panic: ReferenceError: Cannot access 'Base' before initialization\n",
			true,
		},
		{
			"initialized extends",
			"import './b.a'; export class Base {}",
			"import { Base } from './a.a'; class Child extends Base {} console.log('child');",
			"b.a",
			"child\n",
			false,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			directory := test.TempDir()
			for name, source := range map[string]string{"a.a": probe.a, "b.a": probe.b} {
				if writeError := os.WriteFile(
					filepath.Join(directory, name),
					[]byte(source),
					0o644,
				); writeError != nil {
					test.Fatal(writeError)
				}
			}
			// Observe a computed answer before the cycle can panic, rather than
			// letting a failing cycle pass the oracle on silence.
			if writeError := os.WriteFile(
				filepath.Join(directory, "probe.a"),
				[]byte("console.log(`${1 + 2}`);"),
				0o644,
			); writeError != nil {
				test.Fatal(writeError)
			}
			entry := filepath.Join(directory, probe.entry)
			original, readError := os.ReadFile(entry)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(
				entry,
				append([]byte("import './probe.a';\n"), original...),
				0o644,
			); writeError != nil {
				test.Fatal(writeError)
			}
			exit := 0
			if probe.fails {
				exit = 70
			}
			program := agreeEntry(test, entry, nil, exit, nil, nil)
			probe.output = "3\n" + probe.output
			// Behavior cannot see whether a redundant readiness check was eliminated.
			if probe.name == "initialized value" &&
				strings.Contains(native.C(program), "ReferenceError: Cannot access 'value'") {
				test.Error("proven imported read retained a TDZ check")
			}
			if probe.name == "initialized extends" &&
				strings.Contains(native.C(program), "ReferenceError: Cannot access 'Base'") {
				test.Error("proven base class read retained a TDZ check")
			}
			binary := filepath.Join(directory, "native")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			run := func(command *exec.Cmd) {
				test.Helper()
				output, combinedOutputError := command.CombinedOutput()
				if string(output) != probe.output {
					test.Errorf(
						"%s: got %q, want %q (error %v)",
						command.Path,
						output,
						probe.output,
						combinedOutputError,
					)
				}
				if probe.fails {
					exit, ok := combinedOutputError.(*exec.ExitError)
					if !ok || exit.ExitCode() != 70 {
						test.Errorf("expected exit 70, got %v", combinedOutputError)
					}
				} else if combinedOutputError != nil {
					test.Error(combinedOutputError)
				}
			}
			command := exec.Command(binary)
			if probe.fails {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			}
			run(command)
		})
	}
}
