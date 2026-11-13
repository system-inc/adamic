package lower

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
)

func TestNeverConstructsAreRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, probe := range []struct{ name, source, what string }{
			{"any", `const x: any = 1; console.log(x);`, "a value of type any: any has no proven type"},
			{"eval", `eval("1 + 1");`, "reading the global eval: a native binary has no compiler at runtime"},
			{"new_function", `const f = new Function("return 1");`, "constructing a function with the global Function: a native binary has no compiler at runtime"},
			{"new_function_statement", `new Function("return 1");`, "constructing a function with the global Function: a native binary has no compiler at runtime"},
			{"function_call", `Function("return 1");`, "constructing a function with the global Function: a native binary has no compiler at runtime"},
			{"array_length", `const xs: number[] = [1, 2]; xs.length = 1;`, "assigning a non-tuple array's length: array writes must preserve its existing indices"},
		} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || refused.What != probe.what || refused.Fix == "" {
					t.Fatalf("Never construct must be Refused with its fix: %v", err)
				}
			})
		}
	}
}

func TestNeverArrayLengthWriteForms(t *testing.T) {
	t.Parallel()
	for _, receiver := range []string{"number[]", "Array<number>", "number[] | string[]"} {
		for _, write := range []string{"xs.length = 1", "xs.length += 1", "xs.length -= 1", "xs.length *= 2", "xs.length ||= 1", "xs.length++", "++xs.length", "xs.length--", "--xs.length"} {
			t.Run(receiver+write, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, fmt.Sprintf("function resize(xs: %s): void { %s; }", receiver, write), ".ts")
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "assigning a non-tuple array's length") {
					t.Fatalf("array length write must be Refused: %v", err)
				}
			})
		}
	}
}

func TestNeverConstructControlsStillLower(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const user = { eval(x: number): number { return x; } }; console.log(String(user.eval(1)));`,
		`export {}; class Function { value: number = 1; } console.log(String(new Function().value));`,
		`const x = { length: 2 }; x.length = 1; console.log(String(x.length));`,
		`class Sized { length: number = 2; } const x = new Sized(); x.length++; console.log(String(x.length));`,
		`const xs: [number, number] = [1, 2]; xs.length = 2; console.log(String(xs.length));`,
		`const x: unknown = 1; if (typeof x === "number") console.log(String(x));`,
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(source+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, source, extension); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestNeverControlsKeepCheckerDiagnostics(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, code string }{
		{`export {}; function eval(x: number): number { return x; }`, "TS1215"},
		{`const xs: [number,number] = [1,2]; xs.length = 1;`, "TS2322"},
		{`const xs: readonly number[] = [1,2]; xs.length = 1;`, "TS2540"},
		{`const xs = new Uint8Array(2); xs.length = 1;`, "TS2540"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.code+extension+probe.source, func(t *testing.T) {
				t.Parallel()
				_, err := load.Load([]string{writeAgreementSource(t, "main"+extension, probe.source)})
				var checked *load.CheckError
				if !errors.As(err, &checked) || !strings.Contains(err.Error(), probe.code) {
					t.Fatalf("keep checker diagnostic %s: %v", probe.code, err)
				}
			})
		}
	}
}

type neverRefusalMutant struct{ name, what, filter string }

var neverRefusalMutants = []neverRefusalMutant{
	{"any", "a value of type any: any has no proven type", "any"},
	{"eval", "reading the global eval: a native binary has no compiler at runtime", "eval"},
	{"function", "constructing a function with the global Function: a native binary has no compiler at runtime", "new_function|function_call"},
	{"array", "assigning a non-tuple array's length: array writes must preserve its existing indices", "array_length"},
}

func neverRefusalMutantBinary(t testing.TB, probe neverRefusalMutant) string {
	t.Helper()
	before := `return &Refused{Where: l.program.Where(node), What: "` + probe.what + `", Fix:`
	fixes := map[string]string{"any": "type it unknown and narrow it where it's read, or give it its real type", "eval": "parse the input with decodeJson<T>, or call the function directly", "function": "write a function or closure directly", "array": "build a new array with slice, or use push and pop"}
	where := "node"
	if probe.name == "array" {
		where = "target"
		before = strings.Replace(before, "Where(node)", "Where(target)", 1)
	}
	before += ` "` + fixes[probe.name] + `"}`
	return buildcache.GoTestMutated(t, "", "never-"+probe.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/never_refused.go", Before: before, After: `return l.notYet(` + where + `, "` + probe.what + `")`}})
}

func TestProduct_NeverRefusalMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range neverRefusalMutants {
		t.Run(probe.name, func(t *testing.T) { t.Parallel(); neverRefusalMutantBinary(t, probe) })
	}
}

func TestNeverRefusalMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, probe := range neverRefusalMutants {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			binary := neverRefusalMutantBinary(t, probe)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
			defer cancel()
			stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, binary, "-test.run", "^TestNeverConstructsAreRefused$/^("+probe.filter+")", "-test.count=1", "-test.v"))
			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), "Never construct must be Refused with its fix") {
				t.Fatalf("refusal mutant must fail its pin: err=%v stdout=%s stderr=%s", err, stdout, stderr)
			}
			t.Logf("%s mutant caught: %s", probe.name, stdout)
		})
	}
}

// These spellings stop before the Never property-write rule. Keep their existing
// representation gaps distinct from a refused non-tuple array .length write.
func TestNeverArrayLengthUnsupportedTargetsStayNotYet(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, probe := range []struct{ name, source, what string }{
			{"element_length", `const xs: number[] = [1, 2]; xs["length"] = 1;`, "an array index that isn't a number"},
			{"destructure_length", `const xs: number[] = [1, 2]; [xs.length] = [1];`, "assigning to a PropertyAccessExpression in a destructuring assignment"},
		} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				if !errors.As(err, &notYet) || notYet.What != probe.what {
					t.Fatalf("unsupported length target must keep NotYet %q: %v", probe.what, err)
				}
			})
		}
	}
}
