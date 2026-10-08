package indexedc

import (
	"context"
	"fmt"
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

// Not parallel: the two original compound-read shapes and their backend mutants
// run sequentially, using the same independently observed hole values as Node.
func TestHoleCompoundReads(t *testing.T) {
	cli := filepath.Join(t.TempDir(), "adamic")
	if built := run(t, "go", "build", "-o", cli, "../../cmd/adamic"); built.code != 0 {
		t.Fatalf("build CLI: %+v", built)
	}
	for _, flag := range []int{1, 2} {
		id := fmt.Sprintf("D%d", 154+flag)
		t.Run(id, func(t *testing.T) {
			for _, present := range []bool{true, false} {
				t.Run(fmt.Sprintf("present-%t", present), func(t *testing.T) {
					directory := t.TempDir()
					path := filepath.Join(directory, "main.ts")
					initialization := ""
					before := "undefined"
					if present {
						initialization = "enabledSyntaxKindFeatures[kind] = 0;"
						before = "0"
					}
					source := fmt.Sprintf("const enabledSyntaxKindFeatures = new Array<number>(2);\nconst kind = 0;\n%s\nconsole.log(\x60${enabledSyntaxKindFeatures[kind]}\x60);\nenabledSyntaxKindFeatures[kind] |= %d;\nconsole.log(\x60${enabledSyntaxKindFeatures[kind]}\x60);\n", initialization, flag)
					write(t, path, source)
					write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"lib":["es2024"],"module":"esnext","moduleDetection":"force","noEmit":true},"files":["main.ts"]}`)
					wantNode := observation{fmt.Sprintf("%s\n%d\n", before, flag), "", 0}
					if node := run(t, "node", path); node != wantNode {
						t.Fatalf("source Node: %+v, want %+v", node, wantNode)
					}
					checked, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					program, err := lower.Lower(context.Background(), checked)
					if err != nil {
						t.Fatal(err)
					}
					where := fmt.Sprintf("%s:5:1", path)
					checks := ir.InsertedChecks(program)
					if len(checks) != 1 || checks[0] != (ir.InsertedCheck{Kind: "indexed-presence", Where: where}) {
						t.Fatalf("compound read check: %+v", checks)
					}
					explain := run(t, cli, "--explain-checks", path)
					expectedExplain := where + ": checked indexed-presence\nchecked: indexed-presence=1 catch-error=0 json-stringify-defined=0 optional-write=0\ntrusted: 0\n"
					if explain.code != 0 || explain.stdout+explain.stderr != expectedExplain {
						t.Fatalf("compound explain: %+v", explain)
					}
					want := wantNode
					if !present {
						want = observation{"undefined\n", "adamic: panic: indexed read is absent: " + where + "\n", 70}
					}
					module, err := filepath.Abs("../../oracle/adamic.mjs")
					if err != nil {
						t.Fatal(err)
					}
					js := strings.Replace(javascript.JavaScript(program), "from 'adamic'", "from 'file://"+filepath.ToSlash(module)+"'", 1)
					jsPath := filepath.Join(directory, "backend.mjs")
					write(t, jsPath, js)
					if got := run(t, "node", jsPath); got != want {
						t.Fatalf("JS: %+v, want %+v", got, want)
					}
					c := native.C(program)
					for _, sanitized := range []bool{false, true} {
						binary := filepath.Join(directory, fmt.Sprintf("native-%t", sanitized))
						if err := native.Build(c, binary, native.Options{Sanitize: sanitized}); err != nil {
							t.Fatal(err)
						}
						if got := run(t, binary); got != want {
							t.Fatalf("native: %+v, want %+v", got, want)
						}
						if !present {
							panicCall := regexp.MustCompile(`adamic_panic\([^;\n]*->bytes[^;\n]*\);`)
							if len(panicCall.FindAllString(c, -1)) != 1 {
								t.Fatal("requires exactly one compound guard")
							}
							mutant := binary + "-mutant"
							if err := native.Build(panicCall.ReplaceAllString(c, "(void)0;"), mutant, native.Options{Sanitize: sanitized}); err != nil {
								t.Fatal(err)
							}
							got := run(t, mutant)
							if got != wantNode {
								t.Fatalf("erased compound guard must reproduce Node's unguarded bitwise coercion: %+v", got)
							}
							t.Logf("%s compound guard erased: mutant exits 0 printing %q; named exit-70 assertion catches it; sanitized=%t", id, got.stdout, sanitized)
						}
					}
				})
			}
		})
	}
}
