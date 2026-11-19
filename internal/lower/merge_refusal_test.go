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
		{"required_unused", `class Merged { a = 1; }
interface Merged { g: string; }
`, "main.a:2:1: Adamic 0.1 refuses an interface declaration merged into class Merged, adding member g; declare g on the class itself"},
		{"namespace_unused", `class Merged { a = 1; }
namespace Merged { const local = 0; export const k = 1, later = 2; }
`, "main.a:2:1: Adamic 0.1 refuses a namespace declaration merged into class Merged, adding member k; declare k as a static member on the class itself or export it separately"},
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
	for _, probe := range []struct {
		name, source, extension string
	}{
		{"interfaces", `interface Both { a: number; }
interface Both { f: string; }
const x: Both = { a: 1, f: "m" };
console.log(x.f);
`, ".a"},
		{"empty_interface", `class Merged { a = 1; }
interface Merged {}
console.log(String(new Merged().a));
`, ".a"},
		{"typescript", classMergeWriteSource, ".ts"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, probe.extension); err != nil {
				t.Fatal(err)
			}
		})
	}
}
