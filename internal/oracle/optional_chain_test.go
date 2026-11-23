package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// An optional chain's work after its nullish receiver runs under one guard (#cdpxjmb, docs/rulings.md row
// 207, lower/optional_chain.go), and m?.size on a Map or a Set is number | undefined (#8zcm7g4): each
// program in testdata/optional_chain agrees with Node natively, sanitized and on the JavaScript backend,
// and leaks nothing. chain.a holds the ruling's shapes, a?.b(), a?.b.c.d, a?.b.c(), a?.b?.c and a?.b().c,
// on an undefined, a null and a present receiver, the receiver read once and no argument evaluated where
// the chain stops, a call as a statement and a void one, and a function replaced while its arguments run,
// which isn't the one called; union_method.a is tsc's moduleNameResolver.ts shape, an optional method of a
// union of two host types; size.a prints each size itself, so 0 and undefined stay apart; fields.a reads a
// field declared optional as the last step, still number | undefined under the guard (Lower's must); and
// grouped_call.a is (greeter?.greet)(), which parentheses end, called with its object as Node calls it. Node's own
// output is pinned too, so a fixture that stopped exercising its rows shows here before it shows as an
// agreement. Mutants (scratch-oc): the steps run on the missing receiver, the receiver as the missing
// chain's value, the receiver evaluated again by the steps, the steps evaluated before the guard, the
// callee read after the arguments on the JavaScript backend, only undefined tested, ?.size held as a
// number, and every step's undefined stripped, a field's own included (fields.a natively); each goes red here.
func TestOptionalChainsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, program := range []struct{ name, want string }{
		{"chain.a", "false: undefined undefined undefined undefined undefined undefined undefined reads=9 argued=0\nnote void\ntrue: hello 4 L4! undefined leaf3 leaf4 2 reads=9 argued=6\nundefined first 1 second 1\nundefined talk 2 retalk 2\n"},
		{"union_method.a", "undefined undefined /cache undefined calls=1 reads=4\n"},
		{"size.a", "undefined 0 1\nundefined 0\ntrue true false\nundefined undefined 1 2\n"},
		{"fields.a", "undefined undefined\nundefined undefined\n0 false\n3 true\n"},
		{"grouped_call.a", "ada\n"},
	} {
		t.Run(program.name, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/optional_chain/" + program.name
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); string(node.stdout) != program.want || node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, want exit 0 and %q", node.exitCode, node.stdout, program.want)
			}
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(t, relative)
		})
	}
}

// A ! inside a guarded chain keeps its check, and parentheses end a chain (#cdpxjmb, Lower's musts):
// chained.ts's outer?.b!.c is undefined where outer is missing, and where b is, the assertion fails;
// grouped.ts's (outer?.b.c)!.length asserts the chain's value outside its guard, so a missing outer fails
// the assertion rather than giving undefined. Node throws reading a property of undefined at both, so
// the legs are held to the assertion's words, as asserted_method_test.go holds them, and Node to the
// stdout before it.
func TestOptionalChainAssertionsKeepTheirChecks(t *testing.T) {
	t.Parallel()
	for _, program := range []struct{ name, stdout, line, text string }{
		{"chained.ts", "undefined\nhere\n", ":11:", "outer?.b!"},
		{"grouped.ts", "4\n", ":11:", "(outer?.b.c)!"},
	} {
		t.Run(program.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_chain/asserted/"+program.name))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); node.exitCode == 0 || string(node.stdout) != program.stdout || !strings.Contains(string(node.stderr), "TypeError: Cannot read properties of undefined") {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q, want %q and a TypeError reading undefined", node.exitCode, node.stdout, node.stderr, program.stdout)
			}
			compiled, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, _ := natively(t, compiled)
			for name, got := range map[string]run{"native": native, "release": released(t, compiled), "JavaScript": onJavaScriptBackend(t, compiled)} {
				stderr := string(got.stderr)
				if got.exitCode != 70 || string(got.stdout) != program.stdout || !strings.HasPrefix(stderr, "adamic: panic: non-null assertion failed at "+path+program.line) || !strings.HasSuffix(stderr, ": "+program.text+" is null or undefined\n") {
					t.Fatalf("%s: exit %d, stdout %q, stderr %q, want 70, %q and the assertion on %s", name, got.exitCode, got.stdout, got.stderr, program.stdout, program.text)
				}
			}
		})
	}
}
