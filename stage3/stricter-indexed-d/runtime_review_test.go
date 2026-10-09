package indexed_d

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNullableRuntimeReview(t *testing.T) {
	t.Parallel()
	fixtures := []struct{ name, source string }{
		{"map-set", `function keys(): void {
 const map = new Map<string | null | undefined, string>();
 map.set("", "empty"); map.set(null, "nil"); map.set("nu" + "ll", "text"); map.set(undefined, "missing");
 console.log(String(map.size)); console.log(String(map.get(undefined))); console.log(String(map.get(""))); console.log(String(map.get(null))); console.log(String(map.get("null")));
 map.set(null, "updated"); console.log(String(map.get(""))); console.log(String(map.get(null)));
 console.log(String(map.delete(null))); console.log(String(map.has(""))); console.log(String(map.has(null))); console.log(String(map.has("null")));
 const set = new Set<string | null | undefined>();
 set.add(""); set.add(null); set.add("nu" + "ll"); set.add(null); set.add(undefined);
 console.log(String(set.size)); console.log(String(set.has(undefined))); console.log(String(set.has(""))); console.log(String(set.has(null))); console.log(String(set.has("null")));
 console.log(String(set.delete(null))); console.log(String(set.has(""))); console.log(String(set.has(null))); console.log(String(set.has("null")));
}
keys();
`},
		{"narrow-concat", `function observe(x: string | null): void { if (x !== null) { console.log(x + "!"); } }
observe("a" + "bc"); observe(""); observe("null"); observe(null);
`},
		{"narrow-length", `function observe(x: string | null): void { if (x !== null) { console.log(String(x.length)); } }
observe("a" + "bc"); observe(""); observe("null"); observe(null);
`},
		{"narrow-json", `function observe(x: string | null): void { if (x !== null) { console.log(JSON.stringify(x)); } }
observe("a" + "bc"); observe(""); observe("null"); observe(null);
`},
		{"narrow-slice", `function observe(x: string | null): void { if (x !== null) { console.log(x.slice(1, 3)); } }
observe("a" + "bc"); observe(""); observe("null"); observe(null);
`},
		{"narrow-comparison", `function observe(x: string | null): void { if (x !== null) { console.log(String(x === "null")); console.log(String(x < "null")); } }
observe("a" + "bc"); observe(""); observe("null"); observe(null);
`},
		{"null-template", "function observe(x: string | null): void { console.log(`value:${x}`); }\nobserve(null); observe(\"null\"); console.log(`literal:${null}`);\n"},
		{"null-concat", `function observe(x: string | null): void { console.log("value:" + x); console.log(x + ":value"); }
observe(null); observe("null"); console.log("literal:" + null); console.log(null + ":literal");
`},
		{"null-json", `function observe(x: string | null): void { console.log(JSON.stringify(x)); }
observe(null); observe("null"); console.log(JSON.stringify(null));
`},
		{"null-string", `function observe(x: string | null): void { console.log(String(x)); }
observe(null); observe("null"); console.log(String(null));
`},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			program, _, node := nullableProgram(t, fixture.source, true)
			c := nullableBackends(t, program, node)
			t.Logf("Node %q; JS, release and ASan/UBSan match", node.stdout)
			if fixture.name == "map-set" {
				binary := nullableMapMutant(t, c)
				got := run(binary)
				if got == node {
					t.Fatal("map sentinel address-check mutant survived")
				}
				t.Logf("map address-check mutant caught: exit=%d stdout=%q stderr=%q", got.code, got.stdout, got.stderr)
			}
			if fixture.name == "null-json" {
				binary := nullableRuntimeMutant(t, c, "json_stringify", []string{
					"\tif (kind == adamic_json_string && adamic_reference_is_sentinel(value.reference)) {\n\t\tkind = adamic_json_null;\n\t}\n",
				})
				got := run(binary)
				if got == node {
					t.Fatal("JSON sentinel classification mutant survived")
				}
				t.Logf("JSON sentinel mutant caught: exit=%d stdout=%q stderr=%q", got.code, got.stdout, got.stderr)
			}
		})
	}
}

func TestNullableSentinelHeader(t *testing.T) {
	t.Parallel()
	_, _, node := nullableProgram(t, "console.log(String(null));\n", true)
	const c = `#include "adamic.h"
int main(void) {
 if (adamic_null_string.heap.references != 0 || adamic_null_string.heap.kind != adamic_kind_string) { return 1; }
 if (adamic_reference_null(adamic_kind_string) != &adamic_null_string) { return 2; }
 adamic_write_line(adamic_stdout, &adamic_null_string);
 return 0;
}
`
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "sentinel-header")
		if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := run(binary); got != node {
			t.Fatalf("raw sentinel output sanitize=%t: %+v, Node %+v", sanitize, got, node)
		}
	}
	t.Logf("one immortal string-kind sentinel directly prints Node String(null): %q", node.stdout)
}

func nullableMapMutant(t *testing.T, c string) string {
	t.Helper()
	return nullableRuntimeMutant(t, c, "map", []string{
		"\t\tif (adamic_reference_is_sentinel(string)) {\n\t\t\treturn 0x9e3779b97f4a7c15ull;\n\t\t}\n",
		"\t\tif (adamic_reference_is_sentinel(a) || adamic_reference_is_sentinel(b)) {\n\t\t\treturn a == b;\n\t\t}\n",
	})
}

func nullableRuntimeMutant(t *testing.T, c, member string, targets []string) string {
	t.Helper()
	options := native.Options{Sanitize: true}
	library, err := native.RuntimeLibrary("", options)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	archive := filepath.Join(directory, "libmutant.a")
	original, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	write(t, archive, string(original))
	if got := run("ar", "d", archive, member+".o"); got.code != 0 {
		t.Fatalf("remove map runtime: %+v", got)
	}
	source, err := os.ReadFile("../../internal/native/runtime/" + member + ".c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, target := range targets {
		if strings.Count(text, target) != 1 {
			t.Fatalf("map mutant target missing: %q", target)
		}
		text = strings.Replace(text, target, "", 1)
	}
	mutant := filepath.Join(directory, member+".c")
	write(t, mutant, text)
	main := filepath.Join(directory, "main.c")
	write(t, main, c)
	binary := filepath.Join(directory, "map-mutant")
	arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, main, mutant)
	arguments = append(arguments, native.RuntimeLinkFlags(archive)...)
	arguments = append(arguments, "-lm")
	if got := run("clang", arguments...); got.code != 0 {
		t.Fatalf("map mutant build is not a kill: %s %+v", fmt.Sprint(arguments), got)
	}
	return binary
}
