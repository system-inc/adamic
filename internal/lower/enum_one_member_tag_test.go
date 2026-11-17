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

// A one-member numeric enum's member is its enum's whole type, and tsc lets any number into it: First's
// kind is One.A, yet `{ kind: n, a }` with n a number holds Two.B's value. A tag test over a field that can
// hold such a member proves no layout (tagHoldsAnyNumber). Before it, isSecond's comparison passed a First
// as a Second, and the read of b stopped at exit 70 where Node read undefined and threw (Language's
// J-pmbsp94 on #x2cjvy9, live on main 910f21af). Mutant: tagHoldsAnyNumber answering openNumericEnumType,
// which lowers the predicate rows again.
func oneMemberProgram(enums, predicate, read string) string {
	return enums + `interface First extends Base {
  readonly kind: One.A;
  readonly a: string;
}
const n: number = [1, 2].length - 1;
` + predicate + `function read(s: Base): string {
  try {
    return ` + read + `;
  } catch (error) {
    return "threw";
  }
}
const f: First = { kind: n, a: "x" };
console.log(read(honest) + " " + read(f));
`
}

const oneMemberTwo = `enum One {
  A,
}
enum Two {
  B = 1,
}
interface Base {
  readonly kind: One | Two;
}
interface Second extends Base {
  readonly kind: Two.B;
  readonly b: number;
}
const honest: Second = { kind: Two.B, b: 2 };
`

const oneMemberMany = `enum One {
  A,
}
enum Many {
  B = 1,
  C = 2,
}
interface Base {
  readonly kind: One | Many;
}
interface Second extends Base {
  readonly kind: Many.B;
  readonly b: number;
}
const honest: Second = { kind: Many.B, b: 2 };
`

// The predicate over Base, beside one-member enums alone (q02) and beside a multi-member enum (q05), is no
// proof of Second's layout, in .a and .ts. It stops at compile time, never at the read.
func TestOneMemberEnumTagProvesNoLayout(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"one-member enums", oneMemberProgram(oneMemberTwo, "function isSecond(s: Base): s is Second {\n  return s.kind === Two.B;\n}\n", `isSecond(s) ? s.b.toFixed(1) : "other"`)},
		{"beside a multi-member enum", oneMemberProgram(oneMemberMany, "function isSecond(s: Base): s is Second {\n  return s.kind === Many.B;\n}\n", `isSecond(s) ? s.b.toFixed(1) : "other"`)},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, probe.source, extension); err == nil {
					t.Fatal("lowered; want a compile-time stop, since a one-member enum's member holds any number")
				}
			})
		}
	}
}

// The write alone, a number into One.A, lowers and agrees with Node (q00): the slot rule is #snypn8c's.
func TestOneMemberEnumWriteAgrees(t *testing.T) {
	t.Parallel()
	source := oneMemberTwo + "interface First extends Base {\n  readonly kind: One.A;\n  readonly a: string;\n}\nconst n: number = [1, 2].length - 1;\nconst f: First = { kind: n, a: \"x\" };\nconsole.log(String(f.kind) + \" \" + f.a);\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "1 x\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// q01: `s as Second` in .ts over a First is a lie the checked view stops loudly at the read (rulings.md:18,
// Compiler's word on #x2cjvy9). It's held here either way: exit 70 at the read of b on both backends, where
// Node reads undefined and throws.
func TestOneMemberEnumCastStopsAtTheRead(t *testing.T) {
	t.Parallel()
	source := oneMemberTwo + `interface First extends Base {
  readonly kind: One.A;
  readonly a: string;
}
const n: number = [1, 2].length - 1;
function read(s: Base): string {
  if (s.kind === Two.B) {
    const t = s as Second;
    return String(t.b + 1);
  }
  return "other";
}
const f: First = { kind: n, a: "x" };
console.log(read(honest));
console.log(read(f));
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
	run := func(command *exec.Cmd, stdout, stderrPart string, code int) {
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
		if out.String() != stdout || got != code || !strings.Contains(errs.String(), stderrPart) || strings.Contains(errs.String(), "AddressSanitizer") {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want %d %q, stderr with %q", command.Path, got, out.String(), errs.String(), code, stdout, stderrPart)
		}
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), "3\nNaN\n", "", 0)
	program, err := lowerTypeScriptAssertionSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	run(command, "3\n", "field read failed", 70)
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), "3\n", "field read failed", 70)
}
