package main

import (
	"context"
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

func TestJSONStringifyUseRuntime(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, typed, node, want string
		absent                  bool
	}{
		{"undefined", `const value: string = JSON.stringify(undefined); console.log(value);`, `const value = JSON.stringify(undefined); console.log(value);`, "undefined\n", true},
		{"function", `const value: string = JSON.stringify(() => 1); console.log(value);`, `const value = JSON.stringify(() => 1); console.log(value);`, "undefined\n", true},
		{"held-result", `const result = JSON.stringify(undefined); const value: string = result; console.log(value);`, `const result = JSON.stringify(undefined); const value = result; console.log(value);`, "undefined\n", true},
		{"observed-result", `const result = JSON.stringify(undefined); console.log(` + "`${result}`" + `);`, `const result = JSON.stringify(undefined); console.log(String(result));`, "undefined\n", false},
		{"present", `const value: string = JSON.stringify(7); console.log(value);`, `const value = JSON.stringify(7); console.log(value);`, "7\n", false},
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
			if len(checks) != wantChecks || wantChecks == 1 && checks[0].Kind != "json-stringify-defined" {
				t.Fatalf("explain must count actual guards, never audit rows: %+v", checks)
			}
			c := native.C(program)
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
				if nativeCode != 70 || !strings.Contains(nativeOutput, "JSON.stringify result is undefined: "+path+":1:") || jsCode != 70 || nativeOutput != jsOutput {
					t.Fatalf("presence guard: native exit %d %q, JS exit %d %q", nativeCode, nativeOutput, jsCode, jsOutput)
				}
				// Erase just this panic in emitted C. Keep the lookup and generated
				// program intact: the mutant must build and lose the named stop.
				panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
				if count := len(panicCall.FindAllString(c, -1)); count != 1 {
					t.Fatalf("want exactly one JSON guard to erase, got %d", count)
				}
				mutant := filepath.Join(directory, "mutant")
				if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), mutant, native.Options{Sanitize: true}); err != nil {
					t.Fatalf("mutant must build: %v", err)
				}
				mutantOutput, mutantCode := indexedRun(mutant)
				if mutantCode == 70 && strings.Contains(mutantOutput, "JSON.stringify result is undefined:") {
					t.Fatal("erased guard mutant survived")
				}
				t.Logf("Node gives undefined; both backends stop at site with exit 70; erased guard caught (mutant exit %d)", mutantCode)
			} else if nativeCode != 0 || jsCode != 0 || nativeOutput != fixture.want || jsOutput != fixture.want {
				t.Fatalf("valid/observed value: native exit %d %q, JS exit %d %q, want %q", nativeCode, nativeOutput, jsCode, jsOutput, fixture.want)
			}
		})
	}
}
