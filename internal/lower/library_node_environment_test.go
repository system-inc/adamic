package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestProcessEnvironmentRejectsNonReadUses(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, source, want string }{
		{"non-string value", `process.env.X = undefined;`, "written with a non-string value"},
		{"non-string write key", `process.env[0] = 'value';`, "read with a non-string key"},
		{"non-string delete key", `delete process.env[0];`, "read with a non-string key"},
		{"optional delete", `delete process.env?.X;`, "used as a value (optional chain)"},
		{"alias write", `const environment = process.env; environment.X = 'value';`, "used as a value"},
		{"alias delete", `const environment = process.env; delete environment.X;`, "used as a value"},
		{"nullish compound", `process.env.X ??= 'value';`, "written"},
		{"prototype write", `process.env['toString'] = 'value';`, "mutation with literal prototype key toString"},
		{"prototype delete", `delete process.env['toString'];`, "mutation with literal prototype key toString"},
		{"compound", `process.env.X += 'value';`, "written"},
		{"passed", `function accept(value: object): void {console.log('called');} accept(process.env);`, "used as a value"},
		{"stored", `const environment = process.env;`, "used as a value"},
		{"spread", `const environment = {...process.env};`, "used as a value"},
		{"keys", `console.log(Object.keys(process.env));`, "used as a value"},
		{"in", `console.log('X' in process.env);`, "used as a value"},
		{"in prototype", `const key = 'toString'; console.log(key in process.env);`, "used as a value"},
		{"forin", `for (const key in process.env) console.log(key);`, "used as a value"},
		{"optional", `console.log(process.env?.X);`, "used as a value"},
		{"number", `console.log(process.env[0]);`, "read with a non-string key"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `import process from 'node:process'; `+one.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "node:process.Process.env "+one.want) {
				t.Fatalf("want env NotYet %s, got %v", one.want, err)
			}
		})
	}
}

func TestProcessEnvironmentRefusesLiteralPrototypeNames(t *testing.T) {
	t.Parallel()
	for _, key := range ir.NodeEnvironmentPrototypeKeys {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `import process from 'node:process'; console.log(process.env.`+key+`);`)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "literal prototype key "+key) || !strings.Contains(notYet.What, "process.env['NAME']") || !strings.Contains(notYet.What, "plain object") {
				t.Fatalf("want located prototype NotYet with fix, got %v", err)
			}
		})
	}
}

// The paired oracle TestProcessEnvironmentAdmittedUsesMatchNode runs these uses
// natively, under ASan and counting, and through the JavaScript backend.
func TestProcessEnvironmentAdmitsStringWritesAndDeletes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`process.env.X = 'value';`,
		`const value = (process.env.X = 'a'); console.log(value);`,
		`const ok = delete process.env.X; console.log(String(ok));`,
		`function write(key: 'A' | 'B', value: 'x' | 'y'): void { process.env[key] = value; } write('A', 'x');`,
		`const key = 'X'; process.env[key] = 'value';`,
		`delete process.env.X;`,
		`const key = 'X'; delete process.env[key];`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, `import process from 'node:process'; `+source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
