package main

import (
	"context"
	"os"
	"os/exec"
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

func indexedRun(command string, arguments ...string) (string, int) {
	output, err := exec.Command(command, arguments...).CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	if exited, ok := err.(*exec.ExitError); ok {
		return string(output), exited.ExitCode()
	}
	return string(output) + err.Error(), -1
}

func TestIndexedPresenceRuntime(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, typed, node, want string
		absent                  bool
	}{
		{"typed-zero-Int8Array", `const values = new Int8Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Int8Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Uint8Array", `const values = new Uint8Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Uint8Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Uint8ClampedArray", `const values = new Uint8ClampedArray(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Uint8ClampedArray(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Int16Array", `const values = new Int16Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Int16Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Uint16Array", `const values = new Uint16Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Uint16Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Int32Array", `const values = new Int32Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Int32Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Uint32Array", `const values = new Uint32Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Uint32Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Float32Array", `const values = new Float32Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Float32Array(2); console.log(String(values[1]));`, "0\n", false},
		{"typed-zero-Float64Array", `const values = new Float64Array(2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Float64Array(2); console.log(String(values[1]));`, "0\n", false},
		{"observed-typed-metadata", `const values = new Uint8Array(2); console.log(` + "`${values[0]}:${values.length}:${typeof values}`" + `);`, `const values = new Uint8Array(2); console.log(String(values[0])+':'+values.length+':'+typeof values);`, "0:2:object\n", false},
		{"typed-return", `function make(): Uint8Array {return new Uint8Array(2);} const values = make(); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `function make() {return new Uint8Array(2);} const values = make(); console.log(String(values[1]));`, "0\n", false},
		{"typed-identity", `const values = new Uint8Array(2); const alias = values; const value: number = values[0]; console.log(` + "`${value}:${values === alias}`" + `);`, `const values = new Uint8Array(2); const alias = values; console.log(String(values[0])+':'+(values === alias));`, "0:true\n", false},
		{"typed-parameter", `function read(values: Uint8Array): number {return values[0];} const value = read(new Uint8Array(2)); console.log(` + "`${value}`" + `);`, `function read(values) {return values[0];} console.log(String(read(new Uint8Array(2))));`, "0\n", false},
		{"typed-enum-length", `enum Size {Count = 2} const values = new Uint8Array(Size.Count); const value: number = values[0]; console.log(` + "`${value}:${values.length}`" + `);`, `const values = new Uint8Array(2); console.log(String(values[0])+':'+values.length);`, "0:2\n", false},
		{"typed-empty", `const values = new Uint8Array(0); const value: number = values[0]; console.log(` + "`${value}`" + `);`, `const values = new Uint8Array(0); const value = values[0]; console.log(String(value));`, "undefined\n", true},
		{"typed-outside", `const values = new Float64Array(2); const value: number = values[2]; console.log(` + "`${value}`" + `);`, `const values = new Float64Array(2); const value = values[2]; console.log(String(value));`, "undefined\n", true},
		{"typed-fractional", `const values = new Int16Array(2); const value: number = values[0.5]; console.log(` + "`${value}`" + `);`, `const values = new Int16Array(2); const value = values[0.5]; console.log(String(value));`, "undefined\n", true},
		{"typed-present", `const values = new Uint32Array(2); const alias = values; const value: number = alias[1]; console.log(` + "`${value}:${values.length}`" + `);`, `const values = new Uint32Array(2); const alias = values; const value = alias[1]; console.log(String(value)+':'+values.length);`, "0:2\n", false},
		{"observed-typed", `const values = new Uint8Array(0); console.log(` + "`${values[0]}`" + `);`, `const values = new Uint8Array(0); console.log(String(values[0]));`, "undefined\n", false},
		{"hole-enum-effects", `enum Size { Count = 2 } function size(): typeof Size { console.log('size'); return Size; } const values: number[] = new Array<number>(size().Count); const value: number = values[0]; console.log(` + "`${value}`" + `);`, `function size() {console.log('size'); return {Count:2};} const values = new Array(size().Count); console.log(String(values[0]));`, "size\nundefined\n", true},
		{"hole-enum-length", `enum Size { Count = 2 } const values: number[] = new Array<number>(Size.Count); const value: number = values[0]; console.log(` + "`${value}`" + `);`, `const values = new Array(2); console.log(String(values[0]));`, "undefined\n", true},
		{"hole-compound", `const values: number[] = new Array<number>(2); values[0] |= 1; console.log('done');`, `const values = new Array(2); console.log(String(values[0]));`, "undefined\n", true},
		{"hole-integer-loop", `const values: number[] = new Array<number>(2); for (let index = 0; index < values.length; index++) { const value: number = values[index]; console.log(` + "`${value}`" + `); }`, `const values = new Array(2); for (let index = 0; index < values.length; index++) { console.log(String(values[index])); break; }`, "undefined\n", true},
		{"hole-reference-write", `const values: string[] = new Array<string>(2); const alias = values; alias[1] = 'a'.repeat(2); const value: string = values[1]; console.log(value);`, `const values = new Array(2); const alias = values; alias[1] = 'a'.repeat(2); const value = values[1]; console.log(value);`, "aa\n", false},
		{"hole-number", `const values: number[] = new Array<number>(2); const value: number = values[0]; console.log(` + "`${value}`" + `);`, `const values = new Array(2); const value = values[0]; console.log(String(value));`, "undefined\n", true},
		{"hole-string", `const values: string[] = new Array<string>(2); const value: string = values[1]; console.log(value);`, `const values = new Array(2); const value = values[1]; console.log(value);`, "undefined\n", true},
		{"hole-partial-fill", `const values: number[] = new Array<number>(3); const alias = values; alias.fill(7, 1, 2); const value: number = values[0]; console.log(` + "`${value}`" + `);`, `const values = new Array(3); const alias = values; alias.fill(7, 1, 2); const value = values[0]; console.log(String(value));`, "undefined\n", true},
		{"hole-alias-write", `const values: number[] = new Array<number>(2); const alias = values; alias[1] = 7; const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Array(2); const alias = values; alias[1] = 7; const value = values[1]; console.log(String(value));`, "7\n", false},
		{"hole-filled", `const values: number[] = new Array<number>(3); values.fill(7, 1, 2); const value: number = values[1]; console.log(` + "`${value}`" + `);`, `const values = new Array(3); values.fill(7, 1, 2); const value = values[1]; console.log(String(value));`, "7\n", false},
		{"observed-hole", `const values: number[] = new Array<number>(2); console.log(` + "`${values[0]}`" + `);`, `const values = new Array(2); console.log(String(values[0]));`, "undefined\n", false},
		{"number", `const values: number[] = []; const value: number = values[3]; console.log(` + "`${value}`" + `);`, `const values = []; const value = values[3]; console.log(String(value));`, "undefined\n", true},
		{"boolean", `const values: boolean[] = []; const value: boolean = values[0]; console.log(` + "`${value}`" + `);`, `const values = []; const value = values[0]; console.log(String(value));`, "undefined\n", true},
		{"string-array", `const values: string[] = []; const value: string = values[0]; console.log(value);`, `const values = []; const value = values[0]; console.log(value);`, "undefined\n", true},
		{"string", `const text = ''; const value: string = text[0]; console.log(value);`, `const text = ''; const value = text[0]; console.log(value);`, "undefined\n", true},
		{"negative", `const values: number[] = [1]; const value: number = values[-1]; console.log(` + "`${value}`" + `);`, `const values = [1]; const value = values[-1]; console.log(String(value));`, "undefined\n", true},
		{"fractional", `const values: number[] = [1]; const value: number = values[0.5]; console.log(` + "`${value}`" + `);`, `const values = [1]; const value = values[0.5]; console.log(String(value));`, "undefined\n", true},
		{"present-number", `const values: number[] = [7]; const value: number = values[0]; console.log(` + "`${value}`" + `);`, `const values = [7]; const value = values[0]; console.log(String(value));`, "7\n", false},
		{"present-string", `const text = 'a'; const value: string = text[0]; console.log(value);`, `const text = 'a'; const value = text[0]; console.log(value);`, "a\n", false},
		{"observed-undefined", `const values: number[] = []; console.log(` + "`${values[0]}`" + `);`, `const values = []; console.log(String(values[0]));`, "undefined\n", false},
		{"observed-string", `const text = ''; console.log(` + "`${text[0]}`" + `);`, `const text = ''; console.log(String(text[0]));`, "undefined\n", false},
		{"joint-optional", `const values: number[] = []; const point: { value?: number } = { value: values[0] }; console.log(` + "`${point.value}`" + `);`, `const values = []; const point = { value: values[0] }; console.log(String(point.value));`, "undefined\n", true},
		{"once", `function array(): number[] { console.log('array'); return []; } function index(): number { console.log('index'); return 0; } const value: number = array()[index()]; console.log(` + "`${value}`" + `);`, `function array() { console.log('array'); return []; } function index() { console.log('index'); return 0; } const value = array()[index()]; console.log(String(value));`, "array\nindex\nundefined\n", true},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "main.ts")
			for file, source := range map[string]string{
				"main.ts":       fixture.typed,
				"tsconfig.json": `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`,
			} {
				if err := os.WriteFile(filepath.Join(directory, file), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if output, code := indexedRun("node", "--eval", fixture.node); code != 0 || output != fixture.want {
				t.Fatalf("Node oracle: exit %d, output %q, want %q", code, output, fixture.want)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			checks := ir.InsertedChecks(program)
			wantChecks := 1
			if strings.HasPrefix(fixture.name, "observed-") {
				wantChecks = 0
			}
			if len(checks) != wantChecks || wantChecks == 1 && checks[0].Kind != "indexed-presence" {
				t.Fatalf("explain must count actual guards, never audit rows: %+v", checks)
			}
			c := native.C(program)
			lookup := "adamic_string_at("
			if strings.Contains(fixture.typed, "[]") || strings.Contains(fixture.name, "typed") {
				lookup = "adamic_array_at("
			}
			// The area represents three kinds as native typed arrays, each with one lookup.
			if strings.Contains(c, "adamic_typed_array_get(") {
				lookup = "adamic_typed_array_get("
			}
			count := strings.Count(c, lookup)
			if lookup == "adamic_array_at(" {
				count += strings.Count(c, "adamic_array_at_integer(")
			}
			if count != 1 {
				t.Fatalf("presence guard duplicated the lookup/bounds check: %s count %d", lookup, count)
			}
			binary := filepath.Join(directory, "native")
			if err := native.Build(c, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			nativeOutput, nativeCode := indexedRun(binary)
			js := javascript.JavaScript(program)
			module, err := filepath.Abs("../../oracle/adamic.mjs")
			if err != nil {
				t.Fatal(err)
			}
			js = strings.Replace(js, "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
			jsPath := filepath.Join(directory, "backend.mjs")
			if err := os.WriteFile(jsPath, []byte(js), 0o644); err != nil {
				t.Fatal(err)
			}
			jsOutput, jsCode := indexedRun("node", jsPath)
			if fixture.absent {
				if nativeCode != 70 || !strings.Contains(nativeOutput, "indexed read is absent: "+path+":1:") || jsCode != 70 || nativeOutput != jsOutput {
					t.Fatalf("presence guard: native exit %d %q, JS exit %d %q", nativeCode, nativeOutput, jsCode, jsOutput)
				}
				// Erase just this panic in emitted C. Keep the lookup and generated
				// program intact: the mutant must build and lose the named stop.
				panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
				if count := len(panicCall.FindAllString(c, -1)); count != 1 {
					t.Fatalf("want exactly one indexed guard to erase, got %d", count)
				}
				mutant := filepath.Join(directory, "mutant")
				if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), mutant, native.Options{Sanitize: true}); err != nil {
					t.Fatalf("mutant must build: %v", err)
				}
				mutantOutput, mutantCode := indexedRun(mutant)
				if mutantCode == 70 && strings.Contains(mutantOutput, "indexed read is absent:") {
					t.Fatal("erased guard mutant survived")
				}
				t.Logf("Node gives undefined; both backends stop at site with exit 70; erased guard caught (mutant exit %d)", mutantCode)
			} else if nativeCode != 0 || jsCode != 0 || nativeOutput != fixture.want || jsOutput != fixture.want {
				t.Fatalf("valid/observed value: native exit %d %q, JS exit %d %q, want %q", nativeCode, nativeOutput, jsCode, jsOutput, fixture.want)
			}
		})
	}
}
