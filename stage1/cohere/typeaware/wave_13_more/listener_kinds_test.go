package wave13more

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	ast "github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The independent production listener table decides which kinds a rule needs.
// Numeric constants come from the pinned parser, never a second handwritten enum.
var parserKinds = map[string]int{
	"KindSourceFile":          int(ast.KindSourceFile),
	"KindCallExpression":      int(ast.KindCallExpression),
	"KindNewExpression":       int(ast.KindNewExpression),
	"KindVariableDeclaration": int(ast.KindVariableDeclaration),
	"KindThrowStatement":      int(ast.KindThrowStatement),
	"KindArrowFunction":       int(ast.KindArrowFunction),
	"KindReturnStatement":     int(ast.KindReturnStatement),
}

type listenerCase struct{ native, production string }

var listenerCases = []listenerCase{
	{"no_unassigned_vars.a", "core/no_unassigned_vars.go"},
	{"preserve_caught_error.a", "core/preserve_caught_error.go"},
	{"consistent_type_exports.a", "typescript/consistent_type_exports.go"},
	{"wave_13_next/no_process_exit_after_output.a", "nexus/correctness_no_process_exit_after_output.go"},
	{"wave_13_next/no_uncleared_race_timeout.a", "nexus/correctness_no_uncleared_race_timeout.go"},
	{"wave_13_next/require_blocking_standard_streams.a", "nexus/correctness_require_blocking_standard_streams.go"},
	{"wave_13_more/no_obj_calls.a", "core/no_obj_calls.go"},
	{"wave_13_more/no_object_constructor.a", "core/no_object_constructor.go"},
	{"wave_13_more/no_promise_executor_return.a", "core/no_promise_executor_return.go"},
}

var listenerDeclaration = regexp.MustCompile(`readonly listenerKinds: number\[\] = \[([0-9, ]+)\];`)

func productionListenerKinds(t *testing.T, path string) []int {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var kinds []int
	goast.Inspect(file, func(node goast.Node) bool {
		literal, ok := node.(*goast.CompositeLit)
		if !ok {
			return true
		}
		typ, ok := literal.Type.(*goast.SelectorExpr)
		if !ok || typ.Sel.Name != "Listeners" {
			return true
		}
		pkg, ok := typ.X.(*goast.Ident)
		if !ok || pkg.Name != "rule" {
			return true
		}
		for _, item := range literal.Elts {
			entry, ok := item.(*goast.KeyValueExpr)
			if !ok {
				t.Fatalf("unexpected listener entry in %s", path)
			}
			key, ok := entry.Key.(*goast.SelectorExpr)
			if !ok {
				t.Fatalf("unexpected listener key in %s", path)
			}
			kind, ok := parserKinds[key.Sel.Name]
			if !ok {
				t.Fatalf("unsupported pinned parser kind %s", key.Sel.Name)
			}
			kinds = append(kinds, kind)
		}
		return false
	})
	if len(kinds) == 0 {
		t.Fatalf("no production listeners in %s", path)
	}
	slices.Sort(kinds)
	return kinds
}

func checkListenerKinds(source string, wanted []int) error {
	matches := listenerDeclaration.FindAllStringSubmatch(source, -1)
	if len(matches) != 1 {
		return fmt.Errorf("expected one numeric listener declaration, got %d", len(matches))
	}
	var got []int
	for _, part := range strings.Split(matches[0][1], ",") {
		kind, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return err
		}
		got = append(got, kind)
	}
	slices.Sort(got)
	if !slices.Equal(got, wanted) {
		return fmt.Errorf("listener kinds %v, production wants %v", got, wanted)
	}
	return nil
}

func TestWave13ListenerKinds(t *testing.T) {
	for _, item := range listenerCases {
		t.Run(item.native, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", item.native))
			if err != nil {
				t.Fatal(err)
			}
			wanted := productionListenerKinds(t, filepath.Join("..", "..", "..", "..", "cohere", "internal", "lint", "rules", item.production))
			if err := checkListenerKinds(string(source), wanted); err != nil {
				t.Fatal(err)
			}
			t.Logf("numeric parser listeners %v agree with production Go", wanted)
		})
	}
}

func TestWave13ListenerKindMutants(t *testing.T) {
	for _, item := range listenerCases {
		t.Run(item.native, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", item.native))
			if err != nil {
				t.Fatal(err)
			}
			text := string(source)
			wanted := productionListenerKinds(t, filepath.Join("..", "..", "..", "..", "cohere", "internal", "lint", "rules", item.production))
			if err := checkListenerKinds(text, wanted); err != nil {
				t.Fatal(err)
			}
			match := listenerDeclaration.FindStringSubmatch(text)
			first := strings.TrimSpace(strings.Split(match[1], ",")[0])
			value, err := strconv.Atoi(first)
			if err != nil {
				t.Fatal(err)
			}
			changed := strings.Replace(match[0], "["+first, "["+strconv.Itoa(value+1), 1)
			mutant := strings.Replace(text, match[0], changed, 1)
			if mutant == text {
				t.Fatal("mutant did not change the real declaration")
			}
			if err := checkListenerKinds(mutant, wanted); err == nil {
				t.Fatal("wrong numeric listener survived")
			} else {
				t.Logf("numeric-listener mutant caught: %v", err)
			}
		})
	}
}
