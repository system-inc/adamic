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
		{"write", `process.env.X = 'value';`, "written"},
		{"compound", `process.env.X += 'value';`, "written"},
		{"element write", `const key = 'X'; process.env[key] = 'value';`, "written"},
		{"delete", `delete process.env.X;`, "deleted"},
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
