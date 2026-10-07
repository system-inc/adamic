package wave28third

import (
	"bytes"
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Not parallel: native builds, sanitizers and timings share this machine.
func TestThirdBatch(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE28_THIRD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave28_third/suite.a")
	binary := h.build(stage0, "third", entry, archive, false)
	virtual := filepath.Join(repository, "cohere/adamic_wave28_third.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/wave28_third/oracle.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	build := exec.Command("go", "build", "-overlay", h.write("oracle-overlay.json", string(data)), "-o", oracle, virtual)
	build.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", build)
	config := h.write("tsconfig.json", fmt.Sprintf(`{"compilerOptions":{"target":"ES2022","module":"ESNext","strict":true,"alwaysStrict":false,"lib":["ES2022","DOM"]},"files":[%q]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts")))
	var paths []string
	for i, source := range thirdControls(t, repository) {
		suffix := "\nexport {};\n"
		if strings.Contains(source, "RegExp('\\1(a)')") {
			suffix = "\n"
		}
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+suffix))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid", exec.Command(oracle, config, manifest, "--valid-sources"))
	t.Logf("control sources %d, Go syntax-valid %d; all source inputs remain in comparison", len(paths), len(strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, rule := range []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"} {
		n := bytes.Count(truth.stdout, []byte("\t"+rule+"\t"))
		if n == 0 {
			t.Fatalf("no positive controls for %s", rule)
		}
		t.Logf("%s positive findings %d", rule, n)
	}
	if os.Getenv("ADAMIC_WAVE28_THIRD_FULL") == "1" {
		h.fullChecks(stage0, entry, archive, oracle, binary, config, manifest, truth.stdout)
		h.releasedChecks()
	}
}
func thirdControls(t *testing.T, repository string) []string {
	sources := []string{"throw 1;", "throw undefined;", "function f(undefined: unknown){throw undefined;}", "/\\1(a)/;", "/(a)\\1/;", "/(a\\1)/;", "/(?<=(a)\\1)b/;", "/\\1(?!(a))/;", "/(a|\\1b)/;", "new RegExp('\\\\1(a)');", "const r=RegExp;new r('\\\\1(a)');", "const {RegExp:r}=globalThis;r('\\\\1(a)');", "foo(function(){});", "foo(function bar(){bar();});", "foo(function(){this.x;}.bind(this));", "foo(function(arguments){arguments;});"}
	seen := map[string]bool{}
	for _, s := range sources {
		seen[s] = true
	}
	for _, name := range []string{"no_throw_literal_test.go", "no_useless_backreference_test.go", "no_useless_backreference_corpus_test.go", "prefer_arrow_callback_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok || len(literal.Elts) == 0 {
				return true
			}
			if array, ok := literal.Type.(*goast.ArrayType); ok {
				if element, ok := array.Elt.(*goast.Ident); ok && element.Name == "string" {
					for _, value := range literal.Elts {
						if str, ok := value.(*goast.BasicLit); ok && str.Kind == token.STRING {
							text, err := strconv.Unquote(str.Value)
							if err != nil {
								t.Fatal(err)
							}
							if !snapshot(text) && !seen[text] && (strings.Contains(text, "throw") || strings.Contains(text, "function") || strings.Contains(text, "RegExp") || strings.HasPrefix(text, "/")) {
								seen[text] = true
								sources = append(sources, text)
							}
						}
					}
				}
			}

			var source goast.Expr
			if first, ok := literal.Elts[0].(*goast.BasicLit); ok {
				source = first
			}
			for _, element := range literal.Elts {
				if pair, ok := element.(*goast.KeyValueExpr); ok {
					if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "source" || key.Name == "sourceText") {
						source = pair.Value
					}
				}
			}
			if literal, ok := source.(*goast.BasicLit); ok && literal.Kind == token.STRING {
				s, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				if !snapshot(s) && !seen[s] && (strings.Contains(s, "throw") || strings.Contains(s, "function") || strings.Contains(s, "RegExp") || strings.HasPrefix(s, "/")) {
					seen[s] = true
					sources = append(sources, s)
				}
			}
			return true
		})
	}
	return sources
}

func snapshot(text string) bool {
	for _, prefix := range []string{"forward ", "backward ", "nested ", "disjunctive ", "intoNegativeLookaround "} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}
