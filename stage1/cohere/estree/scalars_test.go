package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scalarCases() []string {
	return []string{
		"1e-7; 1e-6; 1e20; 1e21; 1.2345678901234567; 5e-324;",
		"0xffffffffffffffff; 0x10000000000000000; 0x00000000000000000;",
		"0b1111111111111111111111111111111111111111111111111111111111111111; 0b10000000000000000000000000000000000000000000000000000000000000000;",
		"0o1777777777777777777777; 0o2000000000000000000000;",
		"1e999; 1e-999; 0; 0o77;",
		"1234567890123456789012345678901234567890n; 0xffffffffffffffffffffffffffffffffffffn;",
		"0b1111111111111111111111111111111111111111111111111111111111111111111111111111111111n; 0o777777777777777777777777777777777777n;",
		"0n; 0x000000000000000000n; 1_000_000_000_000_000_000_000n;",
	}
}
func TestScalarEdges(t *testing.T) {
	cases := scalarCases()
	list := manifest(t, cases)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", list)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"source Node": onNode(t, main, "--manifest", list), "sanitized native": execute(t, "", binary, "--manifest", list), "emitted JS": onNode(t, script, "--manifest", list)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatalf("%s: %s", name, diff)
		}
	}
	t.Logf("%d scalar edge files, %d bytes identical on Go, source Node, sanitized native and emitted JS", len(cases), len(want))
}

func TestScalarOriginalLibraries(t *testing.T) {
	cases := scalarCases()
	checkOriginalLibraries(t, []string{cases[0], cases[5], cases[6], cases[7], "1e-999; 0; 0o77;"}, 0)
}
func TestPinnedNumericGaps(t *testing.T) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	list := manifest(t, []string{"1e999; 0x10000000000000000;"})
	goAnswer := string(execute(t, "", goOracle(t), "--manifest", list))
	if strings.Count(goAnswer, ".value number NaN\n") != 2 {
		t.Fatal("Go numeric gap changed")
	}
	code := `import { pathToFileURL } from 'node:url';
const library=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);
const ast=library.parse('1e999; 0x10000000000000000;', {warnOnUnsupportedTypeScriptVersion:false});
for(const statement of ast.body) console.log(String(statement.expression.value));`
	answer := string(execute(t, "", "node", "--input-type=module", "-e", code, library))
	if answer != "Infinity\n18446744073709552000\n" {
		t.Fatalf("pinned library numeric gap changed: %q", answer)
	}
	t.Logf("Go values NaN/NaN; pinned original converter: %q", answer)
}
