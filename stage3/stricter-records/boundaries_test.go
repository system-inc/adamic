package records

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func checkedProgram(t *testing.T, source string) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	write(t, path, source)
	write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`)
	return path
}

func TestRecordBoundaries(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"record_cast":           `const record: { readonly [key: string]: number } = {}; const fixed = record as {}; console.log("done");`,
		"nested_cast":           `const outer: { readonly record: { readonly [key: string]: number } } = { record: {} }; const fixed = outer as { readonly record: {} }; console.log("done");`,
		"nested_mutable":        `const record: { readonly [key: string]: { [path: string]: string[] } } = {}; console.log("done");`,
		"nested_nullable":       `const record: { readonly [key: string]: { readonly [path: string]: string[] } | null } = {}; console.log("done");`,
		"mutual_recursive":      `interface A { readonly [key: string]: B } interface B { readonly [key: string]: A } const record: A = {}; console.log("done");`,
		"expanding_generic":     `type Growing<T> = { readonly [key: string]: Growing<T[]> }; const record: Growing<string> = {}; console.log("done");`,
		"nested_array_widening": `const record: { readonly [key: string]: { readonly [path: string]: ("a")[] } } = {}; const wide: { readonly [key: string]: { readonly [path: string]: string[] } } = record; console.log("done");`,
		"nested_record_view":    `const record: { readonly [key: string]: { readonly [path: string]: number } } = {}; const wide: { readonly [key: string]: {} } = record; console.log("done");`,
		"mutable":               `const record: { [key: string]: number } = {}; console.log("done");`,
		"nullable":              `const record: { readonly [key: string]: string | null } = {}; console.log("done");`,
		"undefined":             `const record: { readonly [key: string]: string | undefined } = {}; console.log("done");`,
		"recursive":             `interface RecordNode { readonly [key: string]: RecordNode } const record: RecordNode = {}; console.log("done");`,
		"mixed":                 `const record: { readonly [key: string]: number; readonly named: number } = { named: 7 }; console.log("done");`,
		"fixed_to_record":       `const fixed = { entry: 7 }; const record: { readonly [key: string]: number } = fixed; console.log("done");`,
		"record_to_fixed":       `const record: { readonly [key: string]: number } = {}; const fixed: {} = record; console.log("done");`,
		"nested_view":           `const outer: { readonly record: { readonly [key: string]: number } } = { record: {} }; const fixed: { readonly record: {} } = outer; console.log("done");`,
		"array_widening":        `const record: { readonly [key: string]: ("a")[] } = {}; const wide: { readonly [key: string]: string[] } = record; console.log("done");`,
		"reflection":            `const record: { readonly [key: string]: number } = {}; console.log(Object.keys(record).join(","));`,
		"named_read":            `const record: { readonly [key: string]: number } = {}; console.log(String(record.entry));`,
		"proto_literal":         `const record: { readonly [key: string]: string } = { __proto__: "7" }; console.log("done");`,
		"spread":                `const record: { readonly [key: string]: number } = {}; const copy: { readonly [key: string]: number } = { ...record }; console.log("done");`,
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			path := checkedProgram(t, source)
			node := run(t, "node", path)
			if node.code != 0 {
				t.Fatalf("source Node: %+v", node)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), checked)
			if err == nil {
				t.Fatal("unsupported record boundary compiled")
			}
			if !strings.Contains(err.Error(), "stage 0 can't lower") && !strings.Contains(err.Error(), "Adamic 0.1 refuses") {
				t.Fatal(err)
			}
			t.Logf("Node stdout=%q; explicit refusal: %v", node.stdout, err)
		})
	}
}

func TestRecordPrototypeMiss(t *testing.T) {
	t.Parallel()
	path := checkedProgram(t, `const record: { readonly [key: string]: string } = {}; const key = "toString"; console.log(typeof record[key]);`)
	if got := run(t, "node", path); got != (observation{"function\n", "", 0}) {
		t.Fatalf("Node: %+v", got)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	want := observation{"", "adamic: panic: record member 'toString' is missing; records hold own keys only\n", 70}
	c := native.C(program)
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), fmt.Sprintf("native-%t", sanitized))
		if err := native.Build(c, binary, native.Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		if got := run(t, binary); got != want {
			t.Fatalf("native: %+v", got)
		}
	}
	runtime, err := filepath.Abs("../../oracle/adamic.mjs")
	if err != nil {
		t.Fatal(err)
	}
	js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(runtime)+"'", 1)
	backend := filepath.Join(t.TempDir(), "backend.mjs")
	write(t, backend, js)
	if got := run(t, "node", backend); got != want {
		t.Fatalf("backend: %+v", got)
	}
	// The JS half of the existing runtime's own-only boundary needs its own mutant.
	mutant := strings.Replace(js, "if (Object.hasOwn(Object.prototype, key)) panic", "if (false) panic", 1)
	if mutant == js {
		t.Fatal("prototype mutant did not change code")
	}
	write(t, backend, mutant)
	if got := run(t, "node", backend); got != (observation{"undefined\n", "", 0}) {
		t.Fatalf("mutant must run cleanly with the wrong result: %+v", got)
	}
	t.Log("prototype guard erasure caught by exact named stderr and exit 70; mutant exits 0 printing undefined")
}

func TestRecordAliases(t *testing.T) {
	t.Parallel()
	path := checkedProgram(t, `type Paths = { readonly [key: string]: string[] };
 type Versions = { readonly [key: string]: Paths };
 function first(paths: Paths, key: string): string { const items = paths[key]; const item: string = items[0]; return item; }
 const text = "v";
 const paths: Paths = { entry: [text + text] };
 const versions: Versions = { version: paths };
 const alias: Versions = versions;
 function receiver(): Versions { console.log("receiver"); return alias; }
 function versionKey(): string { console.log("version-key"); return "ver" + "sion"; }
 function pathKey(): string { console.log("path-key"); return "en" + "try"; }
 function index(): number { console.log("index"); return 0; }
 console.log(receiver()[versionKey()][pathKey()][index()]);
 console.log(first(alias["version"], "en" + "try"));
 console.log(String(alias === versions));
 console.log(String(alias["version"] === paths));`)
	want := observation{"receiver\nversion-key\npath-key\nindex\nvv\nvv\ntrue\ntrue\n", "", 0}
	if got := run(t, "node", path); got != want {
		t.Fatalf("Node: %+v", got)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := filepath.Abs("../../oracle/adamic.mjs")
	if err != nil {
		t.Fatal(err)
	}
	js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(runtime)+"'", 1)
	backend := filepath.Join(t.TempDir(), "backend.mjs")
	write(t, backend, js)
	if got := run(t, "node", backend); got != want {
		t.Fatalf("backend: %+v", got)
	}
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "native")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		if got := run(t, binary); got != want {
			t.Fatalf("native: %+v", got)
		}
	}
}
