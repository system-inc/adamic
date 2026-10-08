package indexed_d

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func nullableProgram(t *testing.T, source string, unchecked bool) (*ir.Program, string, result) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "nullable.ts")
	write(t, path, source)
	write(t, filepath.Join(directory, "tsconfig.json"), fmt.Sprintf(`{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":%t,"lib":["es2024"],"module":"esnext","noEmit":true},"files":["nullable.ts"]}`, unchecked))
	node := run("node", path)
	if node.code != 0 {
		t.Fatalf("source Node: %+v", node)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	return program, path, node
}

func nullableBackends(t *testing.T, program *ir.Program, want result) string {
	t.Helper()
	directory := t.TempDir()
	module, err := filepath.Abs("../../oracle/adamic.mjs")
	if err != nil {
		t.Fatal(err)
	}
	js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
	path := filepath.Join(directory, "backend.mjs")
	write(t, path, js)
	if got := run("node", path); got != want {
		t.Fatalf("JS: %+v, want %+v", got, want)
	}
	c := native.C(program)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, fmt.Sprintf("native-%t", sanitize))
		if err := native.Build(c, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := run(binary); got != want {
			t.Fatalf("native sanitize=%t: %+v, want %+v", sanitize, got, want)
		}
	}
	return c
}

// Source Node decides the three-state observations, including a runtime-built string,
// the real empty string and the real text "null", which must never be the sentinel.
func TestNullableStringRepresentation(t *testing.T) {
	const source = `function observe(value: string | null | undefined): void {
 console.log(String(value === null));
 console.log(String(value === undefined));
 console.log(String(value == null));
 console.log(typeof value);
 console.log(String(value));
 console.log(value);
 console.log(String(value === ""));
 console.log(String(value === "null"));
 console.log(String(value ?? "fallback"));
}
function pass(value: string | null): string | null { return value; }
const values: (string | null)[] = [null, "ma" + "de", "", "null"];
observe(pass(values[0] ?? null));
observe(pass(values[1] ?? null));
observe(values[4]);
observe(values[2]);
observe(values[3]);
`
	program, _, node := nullableProgram(t, source, true)
	c := nullableBackends(t, program, node)
	sentinel := "adamic_reference_null(adamic_kind_string)"
	if !strings.Contains(c, sentinel) {
		t.Fatal("sentinel mutation has no target")
	}
	mutant := strings.ReplaceAll(c, sentinel, "NULL")
	binary := filepath.Join(t.TempDir(), "sentinel-null")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatalf("mutant build is not a kill: %v", err)
	}
	got := run(binary)
	if got == node {
		t.Fatal("NULL sentinel mutant survived")
	}
	t.Logf("Node controls %q; release, sanitized and JS match; NULL sentinel mutant caught: exit=%d stdout=%q stderr=%q", node.stdout, got.code, got.stdout, got.stderr)
}
