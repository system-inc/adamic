package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func decoratorsOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/decorators_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/decorators_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_decorators_oracle.go")
	original, err := os.ReadFile(filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/expressions.go"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := "func (b *Builder[E]) expr(node *ast.Node)"
	if strings.Count(string(original), anchor) != 1 {
		t.Fatal("Go expr anchor changed")
	}
	changed := strings.Replace(string(original), anchor, "func (b *Builder[E]) adamicOriginalExpr(node *ast.Node)", 1)
	instrumented := filepath.Join(t.TempDir(), "expressions.go")
	if err := os.WriteFile(instrumented, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/expressions.go"): instrumented, filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestDecoratorsMatchesGo(t *testing.T) {
	oracle := decoratorsOracle(t)
	original, err := os.ReadFile("testdata/cfg_consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	type fixture struct{ File, Source, Rule string }
	var inputs []fixture
	if err := json.Unmarshal(original, &inputs); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{
		"@a class C {}", "@a @b @c class C {}", "@same @same class C {}",
		"@π @𐐀 class C { @first @second field: number; @method m(@param x: number) {} }",
		"@(a || b) @f(g(1)) class C {}", "class C { @f() [key()]() {} }",
		"class C { m(@a @b x: number, @c y: string) {} }", "@ class C {}",
	} {
		inputs = append(inputs, fixture{File: "controls.ts", Rule: "decorator-controls", Source: source})
	}
	fixtureData, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	fixtures := filepath.Join(t.TempDir(), "fixtures.json")
	if err := os.WriteFile(fixtures, fixtureData, 0644); err != nil {
		t.Fatal(err)
	}
	adapted := execute(t, "", oracle, "--adapt", fixtures)
	corpus := filepath.Join(t.TempDir(), "facts.json")
	if err := os.WriteFile(corpus, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("actual Go consumer coverage [AST nodes, delegated expressions]: %s", bytes.TrimSpace(execute(t, "", oracle, "--coverage", fixtures)))
	want := execute(t, "", oracle, "--want", fixtures)
	entry, _ := filepath.Abs("decorators_main.a")
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d consumer AST and decorator-order controls, %d identical Go/source Node/emitted JS/sanitized native bytes", bytes.Count(want, []byte{'\n'}), len(want))
	for _, change := range []struct{ name, old, replacement string }{{"drop-first-decorator", "visitExpression(decorator.expression);", "if(decorator !== node.decorators[0]) { visitExpression(decorator.expression); }"}, {"duplicate-expression", "visitExpression(decorator.expression);", "visitExpression(decorator.expression); visitExpression(decorator.expression);"}} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"decorators_main.a", "control_flow_decorators.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "decorators.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "decorators_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and exits cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}
