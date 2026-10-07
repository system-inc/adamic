package wave12

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func typeParametersOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/type_parameters_oracle.go.txt")
	exports, _ := filepath.Abs("testdata/type_parameters_exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_type_parameters_oracle.go")
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
func TestTypeParametersMatchesGo(t *testing.T) {
	oracle := typeParametersOracle(t)
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
		"function f<T>() {}", "function f<T extends A>() {}", "function f<T = B>() {}",
		"function f<T extends A = B, U extends C = D>() {}",
		"class C<T extends typeof x = typeof y, U> {}",
		"interface I<T extends A, U = B> {}", "type A<T extends B = C> = T;",
		"const f = <T extends A = B, U extends C = D>(x: T): U => x;",
		"class C { m<T extends A = B>() {} }", "function f<π extends A = 𐐀>() {}",
		"function f<T extends>() {}", "function f<T = >() {}",
	} {
		inputs = append(inputs, fixture{File: "controls.ts", Rule: "type-parameter-controls", Source: source})
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
	entry, _ := filepath.Abs("type_parameters_main.a")
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d consumer AST and type-parameter-order controls, %d identical Go/source Node/emitted JS/sanitized native bytes", bytes.Count(want, []byte{'\n'}), len(want))
	for _, change := range []struct{ name, old, replacement string }{{"drop-constraint", "visitExpression(parameter.constraint);", ""}, {"swap-constraint-default", "visitExpression(parameter.constraint);\n        visitExpression(parameter.defaultType);", "visitExpression(parameter.defaultType);\n        visitExpression(parameter.constraint);"}} {
		t.Run(change.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"type_parameters_main.a", "control_flow_type_parameters.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "type_parameters.a") {
					if strings.Count(string(data), change.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), change.old, change.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "type_parameters_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiles and exits cleanly; Go comparison catches mutant at row %d", i, difference(got, want))
			}
		})
	}
}
