package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A union's null is the runtime's static adamic_null, a bare header of kind object with nothing after
// it. adamic_instanceof trusted the kind and read ->class past its end (#hy1sanv, J-vyd1ex8: ASan's
// global-buffer-overflow at class_inheritance.c:8). Here a checked claim of Box | number returns true
// for null, so the check at the call tests the Box member on the sentinel: it must stop with exit 70 at
// that check, on sanitized native and on the JavaScript backend, where Node counts 3. Mutant: the
// sentinel test dropped from adamic_instanceof, which brings the overflow back.
func TestNullSentinelIsNoClassInstance(t *testing.T) {
	t.Parallel()
	source := `class Box {
  readonly size: number;
  constructor(size: number) { this.size = size; }
}
function count(values: (Box | number | string | null)[], test: (value: Box | number | string | null) => value is Box | number): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count([new Box(1), 2, "s", null], (value): value is Box | number => typeof value !== "string")));
`
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	run := func(command *exec.Cmd, stdout, stderrPrefix string, code int) {
		t.Helper()
		var out, errs bytes.Buffer
		command.Stdout, command.Stderr = &out, &errs
		got := 0
		if err := command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			got = exit.ExitCode()
		}
		if out.String() != stdout || got != code || !strings.HasPrefix(errs.String(), stderrPrefix) || strings.Contains(errs.String(), "AddressSanitizer") || stderrPrefix == "" && errs.Len() != 0 {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want %d %q, stderr %q...", command.Path, got, out.String(), errs.String(), code, stdout, stderrPrefix)
		}
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), "3\n", "", 0)
	program, err := lowerTypeScriptAssertionSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	prefix := "adamic: panic: predicate an arrow, declared at "
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	run(command, "", prefix, 70)
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), "", prefix, 70)
}

// instanceof on a union holding null answers false, as Node does, and reads nothing past the sentinel.
func TestNullSentinelInstanceofIsFalse(t *testing.T) {
	t.Parallel()
	source := `class Box {
  readonly size: number;
  constructor(size: number) { this.size = size; }
}
function show(value: Box | number | null): string { return value instanceof Box ? "box" : "other"; }
const values: (Box | number | null)[] = [];
values.push(new Box(1));
values.push(null);
values.push(2);
for (const value of values) { console.log(show(value)); }
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "box\nother\nother\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
