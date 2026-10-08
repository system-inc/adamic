package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestGenericResolvedReturnContracts(t *testing.T) {
	source := `
function memoize<T>(callback: () => T): () => T { const value = callback(); return () => value; }
function forEach<T, U>(array: readonly T[], callback: (value: T) => U | undefined): U | undefined { return undefined; }
function filter<T>(array: readonly T[]): readonly T[] | undefined { return array; }
function setTextRange<T>(range: T): T { return range; }
function append<T>(array: T[]): T[] | undefined { return array; }
function unproven<T>(value: T): any { return value; }
memoize(() => 'value');
forEach([1], () => 'value');
filter([1]);
setTextRange('value');
append([1]);
unproven(1);
`
	path := filepath.Join(t.TempDir(), "contracts.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checker, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := &lowering{checker: checker, program: program}
	want := map[string]string{"memoize": "() => string", "forEach": "string | undefined", "filter": "readonly number[] | undefined", "setTextRange": "\"value\"", "append": "number[] | undefined"}
	seen := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier {
			name := node.AsCallExpression().Expression.Text()
			if _, known := want[name]; known || name == "unproven" {
				seen[name] = true
				returns, err := l.resolvedFunctionReturn(node, checker.GetResolvedSignature(node))
				if name == "unproven" {
					var gap *NotYet
					if !errors.As(err, &gap) || !strings.Contains(gap.What, "resolved return type is any") {
						t.Errorf("any return must stay refused: %v", err)
					}
				} else if err != nil {
					t.Errorf("%s: %v", name, err)
				} else if got := checker.TypeToString(returns); got != want[name] {
					t.Errorf("%s resolved return: got %s, want %s", name, got, want[name])
				}
			}
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	for name := range want {
		if !seen[name] {
			t.Errorf("missing call to %s", name)
		}
	}
}

func TestGenericReturnFixtureLowers(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/any_returns_concrete.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestGenericMapperKeepsCheckerIdentity(t *testing.T) {
	if reflect.TypeOf((*typeMapper)(nil)) != reflect.TypeOf((*checker.TypeMapper)(nil)) {
		t.Fatal("generic mapper must retain the checker type identity; an opaque local struct is unsafe to snapshot")
	}
}
