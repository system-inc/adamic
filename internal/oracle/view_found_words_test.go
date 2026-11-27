package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// viewFoundPredicate is a lying type predicate's program: Src's field, held as what, is read through Dst, which
// declares it as want, after isDst answers true for any kind "d". value fills the field and print is what the
// true branch prints.
func viewFoundPredicate(what, want, value, print string) string {
	return "interface Inner { a: number }\ninterface Src { kind: string; field: " + what + " }\ninterface Dst { kind: \"d\"; field: " + want + " }\nfunction isDst(x: Src): x is Dst {\n\treturn x.kind === \"d\";\n}\nconst v: Src = { kind: \"d\", field: " + value + " };\nconsole.log(isDst(v) ? \"dst \" + " + print + " : \"no\");\n"
}

// A checked view read's stop names what it found in the JavaScript backend's words (readiness.go: undefined and null
// are nullish, anything else its typeof, an array array and a Map Map), so both backends print the same line
// (#yf3k0aj). Natively an undefined or null slot, or a boolean | undefined one holding undefined, said "found
// unsupported representation". And a boolean | undefined slot holding a boolean, read as boolean, stopped natively
// where Node and the JavaScript backend ran on. The programs are written to the test's own directory, since a lying
// predicate's stop would count as a disagreement in the admission delta, as predicate_relative_lies does. Mutants in
// view_found_words.mutants.md.
func TestViewFoundWordsAgree(t *testing.T) {
	t.Parallel()
	const stop = "adamic: panic: field read failed: v.field is not a "
	for _, probe := range []struct {
		name, source, node, stderr string
	}{
		{"undefined_string.ts", viewFoundPredicate("string | undefined", "string", "undefined", "typeof v.field"), "dst undefined\n", stop + "string; expected string, found nullish\n"},
		{"undefined_computed.ts", "function maybe(n: number): string | undefined {\n\treturn n > 5 ? \"big\" : undefined;\n}\n" + viewFoundPredicate("string | undefined", "string", "maybe(1)", "typeof v.field"), "dst undefined\n", stop + "string; expected string, found nullish\n"},
		{"null_string.ts", viewFoundPredicate("string | null", "string", "null", "typeof v.field"), "dst object\n", stop + "string; expected string, found nullish\n"},
		{"null_object.ts", viewFoundPredicate("Inner | null", "Inner", "null", "typeof v.field"), "dst object\n", stop + "Inner; expected Inner, found nullish\n"},
		{"undefined_boolean.ts", viewFoundPredicate("boolean | undefined", "boolean", "undefined", "typeof v.field"), "dst undefined\n", stop + "boolean; expected boolean, found nullish\n"},
		{"present_boolean.ts", viewFoundPredicate("boolean | undefined", "boolean", "true", "String(v.field)"), "dst true\n", ""},
		{"present_false.ts", viewFoundPredicate("boolean | undefined", "boolean", "1 > 2", "String(v.field)"), "dst false\n", ""},
		{"present_string.ts", viewFoundPredicate("string | undefined", "string", "\"x\" + String(1)", "v.field"), "dst x1\n", ""},
		{"boolean_as_string.ts", viewFoundPredicate("boolean | undefined | string", "string", "true", "typeof v.field"), "dst boolean\n", stop + "string; expected string, found boolean\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				backend := []string{"sanitized native", "release native", "the JavaScript backend"}[index]
				if probe.stderr == "" {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, backend, got.exitCode, got.stdout, got.stderr)
					}
					continue
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != probe.stderr {
					t.Fatalf("%s: exit %d, stdout %q, stderr %q; want exit 70 and stderr %q", backend, got.exitCode, got.stdout, got.stderr, probe.stderr)
				}
			}
			if probe.stderr == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
