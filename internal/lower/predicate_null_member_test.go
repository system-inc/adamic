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
)

// A checked predicate whose target holds null tests null as a member, as it tests undefined, against
// how the value holds it (#m33ax27): a union's null is its sentinel and its undefined no pointer, so
// the two never meet, and a bare reference's no pointer is null only where its source holds null and
// not undefined. Each claim below is a checked closure, whose body the prover can't read, so its true
// side runs the membership. Each program agrees with Node, on the JavaScript backend and on sanitized
// native. Mutant: null tested as undefined (m33-mutant.diff), which stops every row that passes a null
// through a union (the filter copy, null alone, null | undefined, an object and a number) at its
// check. The bare reference row stays green under it, since typeof reads its null pointer as
// undefined too; the lie below is the one that tells undefined from null.
func TestCheckedPredicateTestsNullAsAMember(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		// The task's program, the shape of #ev14ch9's filter copy at core.ts:302.
		{"filter copy", `function keep<T, U extends T>(array: readonly T[], f: (x: T) => x is U): U[] {
    const result: U[] = [];
    for (const value of array) if (f(value)) result.push(value);
    return result;
}
const xs: (string | null | undefined)[] = ["a", null, undefined];
console.log(String(keep(xs, (v): v is string | null => v !== undefined).length));
`, "2\n"},
		// null alone, out of a union that boxes its string.
		{"null alone", `function count(values: (string | null)[], test: (value: string | null) => value is null): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count(["a", null, "b", null], (value): value is null => ["object"].includes(typeof value))));
`, "2\n"},
		// Both nullish members, each held its own way in the union.
		{"null or undefined", `function count(values: (string | null | undefined)[], test: (value: string | null | undefined) => value is null | undefined): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count(["a", null, undefined, "b"], (value): value is null | undefined => ["object", "undefined"].includes(typeof value))));
`, "2\n"},
		// null beside a class instance and a number: null is tested before the instance, as undefined is.
		{"object and number", `class Box {
  readonly size: number;
  constructor(size: number) { this.size = size; }
}
function count(values: (Box | number | string | null)[], test: (value: Box | number | string | null) => value is Box | number | null): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count([new Box(1), 2, "c", null], (value): value is Box | number | null => ["number", "object"].includes(typeof value))));
`, "3\n"},
		// Box | null is a bare reference, whose no pointer is null since its source holds no undefined.
		{"bare reference", `class Box {
  readonly size: number;
  constructor(size: number) { this.size = size; }
}
function count(values: (Box | null)[], test: (value: Box | null) => value is null): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count([new Box(1), null, new Box(2)], (value): value is null => ["object"].includes(typeof value) && !(value instanceof Box))));
`, "1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// A lying claim of string | null that passes undefined stops at the check, exit 70 with the located
// message, on sanitized native and the JavaScript backend, where Node counts 2. undefined is a union's
// no pointer, never its null sentinel, so the null member doesn't admit it. Mutant: null tested as
// undefined (m33-mutant.diff), where both backends print 2 and exit 0. The array holds no null, so only the undefined test can stop it.
func TestCheckedPredicateNullMemberRefusesUndefined(t *testing.T) {
	t.Parallel()
	source := `function count(values: (string | null | undefined)[], test: (value: string | null | undefined) => value is string | null): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
const values: (string | null | undefined)[] = [];
values.push("a");
values.push(undefined);
console.log(String(count(values, (value): value is string | null => true)));
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
	run := func(command *exec.Cmd, stdout, stderrPrefix, stderrPart string, code int) {
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
		if out.String() != stdout || got != code || !strings.HasPrefix(errs.String(), stderrPrefix) || !strings.Contains(errs.String(), stderrPart) || stderrPrefix == "" && errs.Len() != 0 {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want %d %q, stderr %q...%q", command.Path, got, out.String(), errs.String(), code, stdout, stderrPrefix, stderrPart)
		}
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), "2\n", "", "", 0)
	program, err := lowerTypeScriptAssertionSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	prefix := "adamic: panic: predicate an arrow, declared at "
	part := ", true branch: source string | null | undefined, target string | null"
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	run(command, "", prefix, part, 70)
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), "", prefix, part, 70)
}
