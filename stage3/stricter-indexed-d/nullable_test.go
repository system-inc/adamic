package indexed_d

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	t.Parallel()
	const source = `function observe(value: string | null | undefined): void {
 console.log(String(value === null));
 console.log(String(value === undefined));
 console.log(String(value == null));
 console.log(String(null == value));
 console.log(String(value != null));
 console.log(String(value !== null));
 console.log(String(value !== undefined));
 console.log(typeof value);
 console.log(String(value));
 console.log(String(value));
 console.log(String(value === ""));
 console.log(String(value === "null"));
 console.log(String(value ?? "fallback"));
 console.log(String(value?.length));
}
console.log(String(null == null));
console.log(String(undefined == null));
console.log(String(null != null));
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
	binary := nullableSentinelMutant(t, c)
	got := run(binary)
	if got == node {
		t.Fatal("NULL sentinel mutant survived")
	}
	t.Logf("Node controls %q; release, sanitized and JS match; NULL sentinel mutant caught: exit=%d stdout=%q stderr=%q", node.stdout, got.code, got.stdout, got.stderr)
}

func TestNullableStringPresenceGuard(t *testing.T) {
	t.Parallel()
	for _, values := range []string{`[null]`, `["ma" + "de"]`, `[]`} {
		t.Run(values, func(t *testing.T) {
			source := `const values: (string | null)[] = ` + values + `; const index = 0;
const value: string | null = values[index];
console.log(String(value === null));
console.log(String(value === undefined));
console.log(String(value == null));
console.log(typeof value);
console.log(String(value));
console.log(String(value));
`
			program, path, node := nullableProgram(t, source, false)
			checks := ir.InsertedChecks(program)
			where := path + ":2:30"
			if len(checks) != 1 || checks[0].Kind != "indexed-presence" || checks[0].Where != where {
				t.Fatalf("indexed guard: %+v", checks)
			}
			want := node
			if values == "[]" {
				want = result{stderr: "adamic: panic: indexed read is absent: " + where + "\n", code: 70}
			}
			c := nullableBackends(t, program, want)
			if values == "[]" {
				panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
				if len(panicCall.FindAllString(c, -1)) != 1 {
					t.Fatal("must erase exactly one presence guard")
				}
				for _, sanitize := range []bool{false, true} {
					binary := filepath.Join(t.TempDir(), fmt.Sprintf("no-guard-%t", sanitize))
					if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), binary, native.Options{Sanitize: sanitize}); err != nil {
						t.Fatalf("mutant build is not a kill: %v", err)
					}
					got := run(binary)
					if got != node {
						t.Fatalf("guard mutant must expose Node's absent value: %+v, want %+v", got, node)
					}
					if got == want {
						t.Fatal("presence guard mutant survived")
					}
					t.Logf("erase-guard mutant caught sanitize=%t: exit=%d stdout=%q stderr=%q", sanitize, got.code, got.stdout, got.stderr)
				}
			}
			t.Logf("Node %q; JS, release and sanitized agree with expected exit=%d stderr=%q", node.stdout, want.code, want.stderr)
		})
	}
}

// Recompile the actual runtime identity helper with null returning NULL. The archive
// copy and replacement object live only in this test's temporary directory.
func nullableSentinelMutant(t *testing.T, c string) string {
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
	if err := os.WriteFile(archive, original, 0o644); err != nil {
		t.Fatal(err)
	}
	members := run("ar", "t", archive)
	if members.code != 0 || !strings.Contains("\n"+members.stdout, "\nnullable.o\n") {
		t.Fatalf("nullable runtime member: %+v", members)
	}
	if got := run("ar", "d", archive, "nullable.o"); got.code != 0 {
		t.Fatalf("remove original runtime object: %+v", got)
	}
	runtimeSource, err := os.ReadFile("../../internal/native/runtime/nullable.c")
	if err != nil {
		t.Fatal(err)
	}
	target := "case adamic_kind_string: return &adamic_null_string;"
	if strings.Count(string(runtimeSource), target) != 1 {
		t.Fatal("runtime sentinel mutation must have one target")
	}
	mutant := filepath.Join(directory, "nullable.c")
	write(t, mutant, strings.Replace(string(runtimeSource), target, "case adamic_kind_string: return NULL;", 1))
	main := filepath.Join(directory, "main.c")
	write(t, main, c)
	binary := filepath.Join(directory, "sentinel-null")
	arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, main, mutant)
	arguments = append(arguments, native.RuntimeLinkFlags(archive)...)
	arguments = append(arguments, "-lm")
	if got := run("clang", arguments...); got.code != 0 {
		t.Fatalf("runtime sentinel mutant build is not a kill: %+v", got)
	}
	return binary
}
