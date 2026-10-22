package agreementproducts

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// programs are every program internal/lower's native agreement rows build, by the name of the row that runs it. A row
// reaches its program by name, and the lower agreement C (C) holds each one's source and the C the compiler lowers it
// to, planted as Plant plants it, so a runner that reads products and never builds them (ADAMIC_BUILD_CACHE=read) finds
// every one. A program not here has no product: lower's TestNativeAgreementRowsNameRegisteredPrograms holds every row
// to a name here and every name here to a row.
var programs = map[string]string{
	"ArrayNarrowingOptionalNumber":  "const items: (number | undefined)[] = [1, 2];\nif (items.every(value => value !== undefined)) console.log('total ' + items.reduce((a, b) => a + b, 0));\n",
	"ArrayNarrowingOptionalString":  "const items: (string | undefined)[] = ['a', 'b'];\nif (items.every(value => value !== undefined)) console.log(items.join(''));\n",
	"ArrayNarrowingSomeNegated":     "function isNumber(value: string | number): value is number { return typeof value === 'number'; }\nconst items: (string | number)[] = ['a', 'b', 1];\nitems.pop();\nif (!items.some(isNumber)) {\n let text = '';\n for (const value of items) if (typeof value === 'string') text += value;\n console.log(text);\n}\n",
	"ArrayNarrowingFindLastIndex":   "function isNumber(value: string | number): value is number { return typeof value === 'number'; }\nconst items: (string | number)[] = ['x', 2, 4];\nconsole.log('index ' + items.findLastIndex(isNumber));\n",
	"ArrayNarrowingPreservedUnion":  "function isScalar(value: string | number | boolean): value is string | number { return typeof value !== 'boolean'; }\nconst items: (string | number | boolean)[] = [1, 'b'];\nif (items.every(isScalar)) {\n for (const value of items) console.log(typeof value === 'number' ? 'number ' + value : value);\n}\n",
	"ArrayNarrowingBooleanCallback": "const items: (string | number)[] = [1, 2, 'discard'];\nitems.pop();\nconsole.log(String(items.every((value): boolean => typeof value === 'number')));\n",
	"FX5MixedLiteral":               "const items: (string | number)[] = [1, 'a'];\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\n",
	"RegExpUnicodeClassSource":      "const regex = new RegExp('[[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);",
	"RegExpSlashClassSource":        "const regex = new RegExp('[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);",
	"RegExpEscapedSlashSource":      "const regex = new RegExp('\\\\/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('/')}`);",
	"TypedArrayFromArray": `const values = new Uint8Array([1, 2]);
console.log(values.length.toString());
for (const value of values) { console.log(value.toString()); }
`,
	"TypedArrayFromEmptyArray": `const values = new Uint8Array([]);
console.log(values.length.toString());
for (const value of values) { console.log(value.toString()); }
`,
	"PlantedWrongOutput": "console.log('computed');",
	"PlantedWrongExit":   "console.log('computed');",
}

// Programs are the names of every program the lower agreement C holds, sorted.
func Programs() []string {
	return slices.Sorted(maps.Keys(programs))
}

// Plant is the program name's C as its row builds it: code, with the wrong native answer the row plants, for a row that
// plants one, which fails rather than plants when the C has no site for it; code as it is for every other row. Only
// native shows a planted answer: the row's JavaScript is lowered from the same source, unplanted.
func Plant(name, code string) (string, error) {
	switch name {
	case "PlantedWrongOutput":
		return strings.ReplaceAll(code, "computed", "wrong"), nil
	case "PlantedWrongExit":
		// main returns the host's exit status (process.exitCode, else 0); plant a fixed 7 there.
		const exit = "return adamic_host_exit_status();"
		index := strings.LastIndex(code, exit)
		if index < 0 {
			return "", errors.New("native main has no successful exit to mutate")
		}
		return code[:index] + strings.Replace(code[index:], exit, "return 7;", 1), nil
	}
	return code, nil
}

// C is the lower agreement C, built: the directory holding, for every program, <name>/main.a, its source, and
// <name>/main.c, the C stage 0's `adamic c` lowers it to, planted (Plant). Its owner is internal/lower's
// TestProduct_LowerAgreementC; every other test finds it (FindC).
//
// An agreement native's C was a file its test wrote before the native's window, a read the window can only refuse as
// temp not made by this run (#egjbcre, Builder's traced run). A handed input a test writes is a product, valued by its
// key (Node's ruling, the rule of #zb9268t's shards): so the C is this product, and each native is handed its
// <name>/main.c, a product read its window values by this product's key.
//
// It's a LeanProductOf build of what it's handed, the stage 0 compiler (buildcache.AdamicBuiltElsewhere, found first),
// whose window values that binary by its product's key: every file of the tree it compiles (lower, load and prelude.d.ts,
// ir, native's C emitter and the rest, cmd/adamic's GoInputs). The programs and their plants are this package's code,
// keyed by its recipe and its Files. Outside a trace it's keyed by Inputs alone, which declare the compiler's files and
// this package's, so a change to the lowering misses it on every machine.
func C(t testing.TB) string {
	t.Helper()
	inputs := cInputs(t)
	adamic := buildcache.AdamicBuiltElsewhere(t)
	return buildcache.LeanProductOf(t, buildcache.Handed{Inputs: inputs, Paths: []string{adamic}}, BuildAgreementC)
}

// FindC is the lower agreement C for a test that reads it and whose product test waits for owner, C's product test, in
// cloud/fast-gate/product-order.json: under the trace it's found by name, failing and naming owner when it isn't there,
// and never built, as runtimeproducts.Archive finds an archive; outside the trace it's built as C builds it, the stage 0
// compiler found only when it's missed, so a runner that finds the C never fetches the compiler.
func FindC(t testing.TB, owner string) string {
	t.Helper()
	inputs := cInputs(t)
	return buildcache.BuiltElsewhere(t, "the lower agreement C", inputs, owner, func(directory string) error {
		return BuildAgreementC(buildcache.Handed{Inputs: inputs, Paths: []string{buildcache.Adamic(t)}}, directory)
	})
}

// cInputs key the lower agreement C: its name, the platform, and as files every file of the tree the stage 0 compiler
// compiles (cmd/adamic's GoInputs, its go version among its tool reports) and this package, which holds the programs
// and their plants.
func cInputs(t testing.TB) buildcache.Inputs {
	t.Helper()
	compiler, err := buildcache.GoInputs("adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatalf("go list ./cmd/adamic: %v", err)
	}
	return buildcache.Inputs{
		// v1: lowered by stage 0's `adamic c` in a build of what it's handed (#egjbcre).
		Name:      "lower agreement C v1",
		Files:     append(slices.Clone(compiler.Files), "internal/native/agreementproducts"),
		Toolchain: append([]string{runtime.GOOS, runtime.GOARCH}, compiler.Toolchain...),
	}
}

// BuildAgreementC writes every program's source into directory, <name>/main.a, and runs the compiler it's handed, stage
// 0 by its absolute path, as `adamic c main.a` in <name>, keeping the C it prints, planted (Plant), as <name>/main.c.
// It's a top-level function of what it's handed alone (buildcache.LeanProductOf): no test, no package state but the
// programs' table, no environment (runtimeproducts' TestTheAgreementCBuildReachesOnlyItsDeclaredFiles). The source sits
// in the product's own directory, which reads the same every run, never in `TMPDIR`; the C is the same bytes wherever
// the source is, and the same as internal/lower's tests lower it in process (TestNativeAgreementLeanBuildIsBuildsBytes).
func BuildAgreementC(handed buildcache.Handed, directory string) error {
	if len(handed.Paths) != 1 {
		return fmt.Errorf("agreementproducts: handed %q, want the stage 0 compiler", handed.Paths)
	}
	for name, source := range programs {
		program := filepath.Join(directory, name)
		if err := os.Mkdir(program, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(program, "main.a"), []byte(source), 0o644); err != nil {
			return err
		}
		command := toolchain.CommandIn(program, handed.Paths[0], "c", "main.a")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("agreementproducts: adamic c %s: %v\n%s", name, err, stderr.Bytes())
		}
		code, err := Plant(name, stdout.String())
		if err != nil {
			return fmt.Errorf("agreementproducts: %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(program, "main.c"), []byte(code), 0o644); err != nil {
			return err
		}
	}
	return nil
}
