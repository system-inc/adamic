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
			if strings.Contains(fixture.typed, "[]") {
				lookup = "adamic_array_at("
			}
			if count := strings.Count(c, lookup); count != 1 {
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
