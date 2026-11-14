package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// enumNeverBetween opens every #afs4xvb probe: a numeric enum, a holder of one, and flip, which puts E.A
// back through another name.
const enumNeverBetween = "enum E { A, B }\nfunction stray(value: unknown): string { return \"seen \" + String(value); }\ninterface Holder { e: E }\nfunction flip(holder: Holder): void { holder.e = E.A; }\n"

// enumNeverTraps is whether a lowered program holds enum_never.go's trap.
func enumNeverTraps(t *testing.T, source, extension string) bool {
	t.Helper()
	program, err := lowerSourceAs(t, source, extension)
	if err != nil {
		t.Fatalf("want it to lower, got %v", err)
	}
	for _, function := range program.Functions {
		if function.Name == "enum_never" {
			return true
		}
	}
	return false
}

// A numeric enum's never, which enum_never.go traps, is judged by never_read.go's test (#afs4xvb): with
// something that can write its place between the earliest narrowing and the read it stops as not yet,
// never a trap on a valid run, where Node prints seen 0 or after. holder.e and outer.holder.e checked
// against E.A, then flipped back, then checked against E.B; a let a closure flips (narrowedNeverReadGap
// stops it first); a switch's explicit and implicit default, each judged at its scrutinee
// (enumSwitchScrutineeGap), since the checked default traps the unmatched edge before an explicit
// default's own read of holder.e runs; and enums_open_never_index.a beside a getter, which any element
// read may run, so its three checks of values[0] are code between. Each pins where it stops and that it's
// a not yet. Without the judgment the first four trapped at exit 70 on both backends (J-xz537sq).
func TestEnumNeverAcrossAWriteIsNotYet(t *testing.T) {
	t.Parallel()
	index, err := os.ReadFile("../oracle/testdata/enums_open_never_index.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, source, at string }{
		{"path", enumNeverBetween + "function describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tif (holder.e === E.B) return 'b';\n\t{ const seen: unknown = holder.e; return stray(seen); }\n}\nconsole.log(describe({ e: E.B }));\n", "holder.e; return"},
		{"deep", enumNeverBetween + "interface Outer { holder: Holder }\nfunction flipOuter(outer: Outer): void { outer.holder.e = E.A; }\nfunction describe(outer: Outer): string {\n\tif (outer.holder.e === E.A) return 'a';\n\tflipOuter(outer);\n\tif (outer.holder.e === E.B) return 'b';\n\t{ const seen: unknown = outer.holder.e; return stray(seen); }\n}\nconsole.log(describe({ holder: { e: E.B } }));\n", "outer.holder.e; return"},
		{"let", enumNeverBetween + "function describe(start: E): string {\n\tlet e: E = start;\n\tconst flipLet = (): void => { e = E.A; };\n\tif (e === E.A) return 'a';\n\tflipLet();\n\tif (e === E.B) return 'b';\n\t{ const seen: unknown = e; return stray(seen); }\n}\nconsole.log(describe(E.B));\n", "e; return"},
		{"switch_default", enumNeverBetween + "function describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tswitch (holder.e) {\n\t\tcase E.B: return 'b';\n\t\tdefault: return stray(holder.e);\n\t}\n}\nconsole.log(describe({ e: E.B }));\n", "holder.e) {"},
		{"switch_implicit", enumNeverBetween + "function describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tswitch (holder.e) {\n\t\tcase E.B: return 'b';\n\t}\n\treturn 'after';\n}\nconsole.log(describe({ e: E.B }));\n", "holder.e) {"},
		{"index_getter", string(index) + "class Sized { get size(): number { return 1; } }\nconsole.log(`${new Sized().size}`);\n", "values[0];\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				if want := strings.Replace(neverReadWhere(t, probe.source, probe.at), "main.ts", "main"+extension, 1); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) {
					t.Fatalf("got %v, want a not yet at %s", err, want)
				}
			})
		}
	}
}

// What the judgment keeps trapping (#afs4xvb): enums_open_never_index.a, whose three checks of values[0]
// read only the array when the program has no getter (elementReadsBetween, never_read.mutants.md's
// fifth mutant); holder.e checked against both members with nothing between; and a switch on holder.e
// with no narrowing before it, its implicit default the trap for a value no member is.
func TestEnumNeverWithNothingBetweenTraps(t *testing.T) {
	t.Parallel()
	index, err := os.ReadFile("../oracle/testdata/enums_open_never_index.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, source string }{
		{"index", string(index)},
		{"path", enumNeverBetween + "function describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tif (holder.e === E.B) return 'b';\n\t{ const seen: unknown = holder.e; return stray(seen); }\n}\nconsole.log(describe({ e: E.B }));\n"},
		{"switch_implicit", enumNeverBetween + "function describe(holder: Holder): string {\n\tswitch (holder.e) {\n\t\tcase E.A: return 'a';\n\t\tcase E.B: return 'b';\n\t}\n}\nconsole.log(describe({ e: E.B }) + describe({ e: E.A }));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				if !enumNeverTraps(t, probe.source, extension) {
					t.Fatal("want enum_never's trap")
				}
			})
		}
	}
}
