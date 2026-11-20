package lower

import (
	"errors"
	"strings"
	"testing"
)

const classMergeWriteSource = `interface X { a: number; f?: string }
function put(x: X, value: string): void { x.f = value; }
function read(x: X): string { return String(x.f) + "/" + typeof x.f; }
class Merged { a = 1; }
interface Merged { f?: string; }
const m = new Merged();
put(m, "m");
console.log(read(m) + " " + Object.keys(m).join(","));
`

func TestClassMergeRefusal(t *testing.T) {
	t.Parallel()
	const inherited = "Adamic 0.1 refuses an interface declaration merged into class Merged, adding member "
	const optional = "Adamic 0.1 refuses an interface declaration merged into class Merged, adding member f; declare f on the class itself"
	for _, probe := range []struct {
		name, source, want string
	}{
		{"optional_write", classMergeWriteSource, "main.a:5:1: " + optional},
		{"optional_read", `class Merged { a = 1; }
interface Merged { f?: string; }
const m = new Merged();
console.log(String(m.f));
`, "main.a:2:1: " + optional},
		{"required_unused", `class Base { inherited = 2; }
class Merged extends Base { a = 1; }
interface Merged { a: number; inherited: number; g: string; later: number; }
`, "main.a:3:1: Adamic 0.1 refuses an interface declaration merged into class Merged, adding member g; declare g on the class itself"},
		{"namespace_unused", `class Merged { a = 1; static duplicate = 0; }
namespace Merged { export type duplicate = number; const local = 0; export const k = 1, later = 2; }
`, "main.a:2:1: Adamic 0.1 refuses a namespace declaration merged into class Merged, adding member k; declare k as a static member on the class itself or export it separately"},
		{"namespace_class_merge_twin", `class N {} namespace N {export const x=1;}`, "main.a:1:12: Adamic 0.1 refuses a namespace declaration merged into class N, adding member x; declare x as a static member on the class itself or export it separately"},
		{"m01", `interface Extra { f?: string; g: string }
class Merged { a = 1 }
interface Merged extends Extra {}
`, "main.a:3:1: " + inherited + "f; declare f on the class itself"},
		{"m02", `interface Extra { f?: string; g: string }
class Merged { a = 1 }
interface Merged extends Extra { a: number }
`, "main.a:3:1: " + inherited + "f; declare f on the class itself"},
		{"m09", `class Base { b = 1 }
class Merged { a = 1 }
interface Merged extends Base {}
console.log(String(new Merged().b + 1));
`, "main.a:3:1: " + inherited + "b; declare b on the class itself"},
		{"m10", `interface Extra<T> { g: T }
class Merged { a = 1 }
interface Merged extends Extra<string> {}
`, "main.a:3:1: " + inherited + "g; declare g on the class itself"},
		{"m12", `interface Extra { run(): number }
class Merged { a = 1 }
interface Merged extends Extra {}
`, "main.a:3:1: " + inherited + "run; declare run on the class itself"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want Refused %q, got %v", probe.want, err)
			}
			got := refused.Error()
			if slash := strings.LastIndex(refused.Where, "/"); slash >= 0 {
				got = got[slash+1:]
			}
			if got != probe.want {
				t.Fatalf("got %q, want %q", got, probe.want)
			}
			t.Log(got)
		})
	}
}

func TestClassMergeAllowed(t *testing.T) {
	t.Parallel()
	t.Run("class_field", func(t *testing.T) {
		t.Parallel()
		source := strings.Replace(classMergeWriteSource,
			"class Merged { a = 1; }\ninterface Merged { f?: string; }",
			"class Merged { a = 1; f?: string; }", 1)
		lowersAndAgreesWithNodeAs(t, source, ".a")
	})
	t.Run("duplicate_field", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, `class Present { promised = 7; }
interface Present { promised: number; }
console.log(String(new Present().promised));
`, ".a")
	})
	t.Run("inherited_field", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, `class Base { promised = 8; }
class Present extends Base {}
interface Present { promised: number; }
console.log(String(new Present().promised));
`, ".a")
	})
	t.Run("m11", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, `interface Extra { f?: string; g: string }
class Merged { a = 1; f?: string = "f"; g = "g"; }
interface Merged extends Extra {}
const m = new Merged();
console.log(String(m.f) + m.g);
`, ".a")
	})
	t.Run("typescript", func(t *testing.T) {
		t.Parallel()
		// The TypeScript merge reserves f on Merged, so a write through X
		// reaches a slot even before the instance has an own property for f.
		lowersAndAgreesWithNodeAs(t, classMergeWriteSource, ".ts")
	})
	for _, probe := range []struct {
		name, source string
	}{
		{"interfaces", `interface Both { a: number; }
interface Both { f: string; }
const x: Both = { a: 1, f: "m" };
console.log(x.f);
`},
		{"empty_interface", `class Merged { a = 1; }
interface Merged {}
console.log(String(new Merged().a));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
