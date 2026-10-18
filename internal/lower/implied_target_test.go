package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Drive the refusal walk and statement lowerer with a checker-query observer.
// Fail before a repeated expensive query, so the uncached mutant stays bounded.
func lowerCountedImpliedTargets(t *testing.T, source string) (*lowering, map[*ast.Node]int) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	counts := map[*ast.Node]int{}
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	l.impliedTargetQuery = func(node *ast.Node) *checker.Type {
		counts[node]++
		if counts[node] > 1 {
			t.Fatalf("implied target queried parent kind %v more than once", node.Kind)
		}
		return checked.GetTypeAtLocation(node)
	}
	if err := l.refuse(file); err != nil {
		t.Fatal(err)
	}
	if err := l.declareModule(file.Statements.Nodes); err != nil {
		t.Fatal(err)
	}
	l.result.Main, err = l.statements(file.Statements.Nodes)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.result.Main) == 0 {
		t.Fatal("no statements lowered")
	}
	return l, counts
}

func TestImpliedTargetLargeArrayQueriesOnce(t *testing.T) {
	t.Parallel()
	const elements = 3072
	var source strings.Builder
	source.WriteString("const values = [\n")
	for i := 0; i < elements; i++ {
		fmt.Fprintf(&source, "{value:%d,done:false},\n", i)
	}
	source.WriteString("]; const other = [{value:7}]; console.log(String(values.length + other.length));")
	l, counts := lowerCountedImpliedTargets(t, source.String())
	literals := 0
	for node, count := range counts {
		if node.Kind != ast.KindArrayLiteralExpression {
			continue
		}
		literals++
		if count != 1 {
			t.Fatalf("array queries = %d, want 1", count)
		}
		target := l.impliedTargets[node]
		if target == nil || l.checker.GetPropertyOfType(target, "value") == nil {
			t.Fatal("array element target lost object fields")
		}
	}
	if literals != 2 {
		t.Fatalf("queried %d array literals, want 2", literals)
	}
}

func TestImpliedTargetConditionalQueriesOnce(t *testing.T) {
	t.Parallel()
	var source strings.Builder
	source.WriteString("function choose(n:number):boolean{return n>0;} const result = ")
	const branches = 128
	for i := 0; i < branches; i++ {
		fmt.Fprintf(&source, "choose(%d) ? {value:%d} : (", i, i)
	}
	source.WriteString("{value:999}")
	source.WriteString(strings.Repeat(")", branches))
	source.WriteString(";console.log(String(result.value));")
	l, counts := lowerCountedImpliedTargets(t, source.String())
	conditionals := 0
	for node, count := range counts {
		if node.Kind != ast.KindConditionalExpression {
			continue
		}
		conditionals++
		if count != 1 || l.checker.TypeToString(l.impliedTargets[node]) != l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) {
			t.Fatal("conditional target changed or queried repeatedly")
		}
		// The condition is not a branch, even after its parent's target is cached.
		if l.impliedTarget(node.AsConditionalExpression().Condition) != nil {
			t.Fatal("condition acquired a branch target")
		}
	}
	if conditionals != branches {
		t.Fatalf("queried %d conditionals, want %d", conditionals, branches)
	}
}

func checkImpliedLogicalQueries(t *testing.T, operator string) {
	t.Helper()
	const operands = 256
	var source strings.Builder
	source.WriteString("function box(n:number):{readonly value:number}|undefined{return n<0?undefined:{value:n};} const result = ")
	for i := 0; i < operands; i++ {
		if i != 0 {
			source.WriteString(operator)
		}
		fmt.Fprintf(&source, "box(%d)", i)
	}
	source.WriteString("; console.log(String(result?.value ?? -1));")
	l, counts := lowerCountedImpliedTargets(t, source.String())
	logical := 0
	for node, count := range counts {
		if node.Kind != ast.KindBinaryExpression || node.AsBinaryExpression().OperatorToken.Kind != map[string]ast.Kind{"||": ast.KindBarBarToken, "&&": ast.KindAmpersandAmpersandToken, "??": ast.KindQuestionQuestionToken}[operator] {
			continue
		}
		logical++
		if count != 1 || l.checker.TypeToString(l.impliedTargets[node]) != l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) {
			t.Fatal("logical target changed or queried repeatedly")
		}
	}
	if logical < operands-1 {
		t.Fatalf("queried %d logical expressions, want at least %d", logical, operands-1)
	}
}
func TestImpliedTargetOrQueriesOnce(t *testing.T)  { t.Parallel(); checkImpliedLogicalQueries(t, "||") }
func TestImpliedTargetAndQueriesOnce(t *testing.T) { t.Parallel(); checkImpliedLogicalQueries(t, "&&") }
func TestImpliedTargetNullishQueriesOnce(t *testing.T) {
	t.Parallel()
	checkImpliedLogicalQueries(t, "??")
}

func TestImpliedTargetTupleCachesNil(t *testing.T) {
	t.Parallel()
	l, counts := lowerCountedImpliedTargets(t, "const tuple: readonly [number,number] = [1,2]; console.log(String(tuple[0]));")
	// Contextual tuple elements need not ask impliedTarget during the walk. Ask
	// both explicitly to prove a nil target is cached rather than queried twice.
	file := l.program.Files()[0]
	literal := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration().Initializer
	for _, element := range literal.AsArrayLiteralExpression().Elements.Nodes {
		if l.impliedTarget(element) != nil {
			t.Fatal("tuple unexpectedly acquired array element target")
		}
	}
	if counts[literal] != 1 {
		t.Fatalf("tuple queries = %d, want 1", counts[literal])
	}
}
