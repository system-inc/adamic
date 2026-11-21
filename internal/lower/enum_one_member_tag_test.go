package lower

import (
	"strings"
	"testing"
)

// A one-member numeric enum's member is its enum's whole type to the checker, and tsc lets any number into
// it: First's kind is One.A, yet `{ kind: n, a }` with n a number held Two.B's value, so isSecond's tag test
// passed a First as a Second and the read of b stopped at exit 70 where Node read undefined and threw
// (Language's J-pmbsp94 on #x2cjvy9, live on main 910f21af). The slot is a literal slot (openEnumSlot), so
// the unproven number is refused where it's written, and no tag test is handed the lie. Mutant: openEnumSlot
// answering openNumericEnumType, which lowers every refused row again.
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

// The checker prints a one-member enum's member as its enum, One.
const oneMemberSlotRefusal = "an unproven value assigned to a numeric literal or enum member slot One;"

// Every road Language built stops at the write, in .a and .ts: the write alone (q00), the predicate beside
// one-member enums (q02) and beside a multi-member enum (q05), and the comparison then cast (q01).
func TestOneMemberEnumSlotRefusesANumber(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"write alone", oneMemberTwo + "interface First extends Base {\n  readonly kind: One.A;\n  readonly a: string;\n}\nconst n: number = [1, 2].length - 1;\nconst f: First = { kind: n, a: \"x\" };\nconsole.log(String(f.kind) + \" \" + f.a);\n"},
		{"predicate", oneMemberProgram(oneMemberTwo, "function isSecond(s: Base): s is Second {\n  return s.kind === Two.B;\n}\n", `isSecond(s) ? s.b.toFixed(1) : "other"`)},
		{"beside a multi-member enum", oneMemberProgram(oneMemberMany, "function isSecond(s: Base): s is Second {\n  return s.kind === Many.B;\n}\n", `isSecond(s) ? s.b.toFixed(1) : "other"`)},
		{"cast", oneMemberProgram(oneMemberTwo, "", `s.kind === Two.B ? (s as Second).b.toFixed(1) : "other"`)},
	} {
		for _, extension := range []string{".a", ".ts"} {
			// In .a the cast is refused first, as an unproven downcast, which main already did.
			if probe.name == "cast" && extension == ".a" {
				continue
			}
			// In .a isSecond is refused first, at the predicate: Base's kind holds any number, so Base is
			// never closed and its tag test proves no layout (rulings.md:142 as amended, #tvw1zkp). The write
			// alone holds the slot's refusal in .a.
			want := oneMemberSlotRefusal
			if extension == ".a" && probe.name != "write alone" {
				want = "over Base, whose kind can hold any number"
			}
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, probe.source, extension); err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("want %q, got %v", want, err)
				}
			})
		}
	}
}

// The honest program, every First built with One.A, agrees with Node on every backend, the predicate and
// its read included: refusing the number takes nothing from a program that keeps the slot's promise.
func TestOneMemberEnumHonestAgrees(t *testing.T) {
	t.Parallel()
	source := oneMemberTwo + `interface First extends Base {
  readonly kind: One.A;
  readonly a: string;
}
function isSecond(s: Base): s is Second {
  return s.kind === Two.B;
}
function read(s: Base): string {
  return isSecond(s) ? s.b.toFixed(1) : "other";
}
const f: First = { kind: One.A, a: "x" };
console.log(read(honest) + " " + read(f));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2.0 other\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
