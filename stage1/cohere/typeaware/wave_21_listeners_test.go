package typeaware

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

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Not parallel: native and sanitizer builds use the private checker archives.
// The expected registrations come from production Go listener maps, and the
// numeric values come from the pinned compiler's exported SyntaxKind constants.
func TestWave21ListenerDeclarations(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if requested := os.Getenv("ADAMIC_WAVE21_LISTENER_ARTIFACTS"); requested != "" {
		directory = requested
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	values := map[string]ast.Kind{
		"KindSourceFile": ast.KindSourceFile, "KindEnumDeclaration": ast.KindEnumDeclaration,
		"KindExpressionStatement": ast.KindExpressionStatement, "KindBinaryExpression": ast.KindBinaryExpression,
		"KindElementAccessExpression": ast.KindElementAccessExpression, "KindCallExpression": ast.KindCallExpression,
		"KindNewExpression": ast.KindNewExpression, "KindArrowFunction": ast.KindArrowFunction, "KindReturnStatement": ast.KindReturnStatement,
	}
	names := []string{"no_mixed_enums", "correctness_no_collection_misuse", "correctness_no_discarded_outcome", "correctness_no_discarded_pure_result", "correctness_no_process_exit_after_output", "correctness_no_uncleared_race_timeout", "correctness_require_blocking_standard_streams", "no_obj_calls", "no_object_constructor", "no_promise_executor_return", "set_state_in_effect", "set_state_in_render", "static_components"}
	var imports, output, expected strings.Builder
	firstModule := ""
	for index, name := range names {
		family := "core"
		if index >= 10 {
			family = "react"
		} else if strings.HasPrefix(name, "correctness_") {
			family = "nexus"
		} else if index == 0 {
			family = "typescript"
		}
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", family, name+".go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		kinds := []string{}
		maps := 0
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			typ, ok := literal.Type.(*goast.SelectorExpr)
			if !ok || typ.Sel.Name != "Listeners" {
				return true
			}
			maps++
			for _, element := range literal.Elts {
				pair, ok := element.(*goast.KeyValueExpr)
				if !ok {
					t.Fatal("non-keyed listener", name)
				}
				key, ok := pair.Key.(*goast.SelectorExpr)
				if !ok {
					t.Fatal("non-kind listener", name)
				}
				kind, ok := values[key.Sel.Name]
				if !ok {
					t.Fatal("unknown pinned kind", name, key.Sel.Name)
				}
				kinds = append(kinds, strconv.Itoa(int(kind)))
			}
			return false
		})
		if maps != 1 || len(kinds) == 0 {
			t.Fatal("production listener maps", name, maps)
		}
		manifestPath := filepath.Join(repository, "stage1/cohere/typeaware/wave21_rules", name, "rule.json")
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var manifest struct {
			Name  string   `json:"name"`
			Kinds []string `json:"kinds"`
		}
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			t.Fatal(err)
		}
		if manifest.Name == "" || len(manifest.Kinds) != len(kinds) {
			t.Fatal("manifest shape", name)
		}
		for i, kindName := range manifest.Kinds {
			kind, ok := values["Kind"+kindName]
			if !ok || strconv.Itoa(int(kind)) != kinds[i] {
				t.Fatal("manifest listener differs from Go", name, kindName)
			}
		}

		module := filepath.Join(repository, "stage1/cohere/typeaware", name+".a")
		if family == "react" {
			module = filepath.Join(repository, "stage1/cohere/typeaware/wave21_react", name+".a")
		}
		if index == 0 {
			firstModule = module
		}
		fmt.Fprintf(&imports, "import { syntaxKinds as kinds%d } from %s;\n", index, strconv.Quote(module))
		fmt.Fprintf(&output, "console.log(%s + kinds%d.join(','));\n", strconv.Quote(name+":"), index)
		fmt.Fprintf(&expected, "%s:%s\n", name, strings.Join(kinds, ","))
	}
	entry := h.write("listeners.a", imports.String()+output.String())
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	binary := h.build(stage0, "listeners", entry, archive, false)
	want := []byte(expected.String())
	got := h.must("listeners-run", exec.Command(binary))
	if !bytes.Equal(got.stdout, want) || len(got.stderr) != 0 {
		t.Fatalf("listener contract: got %s want %s stderr %s", got.stdout, want, got.stderr)
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "listeners-asan", entry, sanitized, true)
	got = h.must("listeners-asan-run", exec.Command(asan))
	if !bytes.Equal(got.stdout, want) || len(got.stderr) != 0 {
		t.Fatalf("sanitized listener contract: %s %s", got.stdout, got.stderr)
	}
	data, err := os.ReadFile(firstModule)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	from := "export const syntaxKinds: readonly number[] = [267];"
	if strings.Count(source, from) != 1 {
		t.Fatal("listener mutant anchor")
	}
	source = strings.Replace(source, from, "export const syntaxKinds: readonly number[] = [307];", 1)
	source = strings.ReplaceAll(source, "'./", "'"+filepath.Dir(firstModule)+"/")
	mutantModule := h.write("enum_listener_mutant.a", source)
	mutantEntry := strings.Replace(imports.String(), strconv.Quote(firstModule), strconv.Quote(mutantModule), 1) + output.String()
	mutant := h.build(stage0, "listener-mutant", h.write("listener_mutant_main.a", mutantEntry), archive, false)
	wrong := h.must("listener-mutant-run", exec.Command(mutant))
	if bytes.Equal(wrong.stdout, want) || len(wrong.stderr) != 0 {
		t.Fatal("listener mutant survived or failed outside contract")
	}
	t.Logf("13 numeric listener declarations match production Go, normal and ASAN; wrong-listener mutant exits 0 with empty stderr, byte comparison catches byte %d", firstDifference(wrong.stdout, want))
	t.Logf("listener contract:\n%s", want)
}
