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
func lowerCountedImpliedTargets(test *testing.T, source string) (*lowering, map[*ast.Node]int) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	test.Cleanup(release)
	counts := map[*ast.Node]int{}
	lowering := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	lowering.impliedTargetQuery = func(node *ast.Node) *checker.Type {
		counts[node]++
		if counts[node] > 1 {
			test.Fatalf("implied target queried parent kind %v more than once", node.Kind)
		}
		return checked.GetTypeAtLocation(node)
	}
	if refuseError := lowering.refuse(file); refuseError != nil {
		test.Fatal(refuseError)
	}
	if declareModuleError := lowering.declareModule(file.Statements.Nodes); declareModuleError != nil {
		test.Fatal(declareModuleError)
	}
	var statementsError error
	lowering.result.Main, statementsError = lowering.statements(file.Statements.Nodes)
	if statementsError != nil {
		test.Fatal(statementsError)
	}
	if len(lowering.result.Main) == 0 {
		test.Fatal("no statements lowered")
	}
	return lowering, counts
}

func TestImpliedTargetLargeArrayQueriesOnce(test *testing.T) {
	test.Parallel()
	const elements = 3072
	var source strings.Builder
	source.WriteString("const values = [\n")
	for index := 0; index < elements; index++ {
		fmt.Fprintf(&source, "{value:%d,done:false},\n", index)
	}
	source.WriteString("]; const other = [{value:7}]; console.log(String(values.length + other.length));")
	lowering, counts := lowerCountedImpliedTargets(test, source.String())
	literals := 0
	for node, count := range counts {
		if node.Kind != ast.KindArrayLiteralExpression {
			continue
		}
		literals++
		if count != 1 {
			test.Fatalf("array queries = %d, want 1", count)
		}
		target := lowering.slots.implied[node]
		if target == nil || lowering.checker.GetPropertyOfType(target, "value") == nil {
			test.Fatal("array element target lost object fields")
		}
	}
	if literals != 2 {
		test.Fatalf("queried %d array literals, want 2", literals)
	}
}

func TestImpliedTargetConditionalQueriesOnce(test *testing.T) {
	test.Parallel()
	var source strings.Builder
	source.WriteString("function choose(n:number):boolean{return n>0;} const result = ")
	const branches = 128
	for index := 0; index < branches; index++ {
		fmt.Fprintf(&source, "choose(%d) ? {value:%d} : (", index, index)
	}
	source.WriteString("{value:999}")
	source.WriteString(strings.Repeat(")", branches))
	source.WriteString(";console.log(String(result.value));")
	lowering, counts := lowerCountedImpliedTargets(test, source.String())
	conditionals := 0
	for node, count := range counts {
		if node.Kind != ast.KindConditionalExpression {
			continue
		}
		conditionals++
		if count != 1 ||
			lowering.checker.TypeToString(
				lowering.slots.implied[node],
			) != lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)) {
			test.Fatal("conditional target changed or queried repeatedly")
		}
		// The condition is not a branch, even after its parent's target is cached.
		if lowering.impliedTarget(node.AsConditionalExpression().Condition) != nil {
			test.Fatal("condition acquired a branch target")
		}
	}
	if conditionals != branches {
		test.Fatalf("queried %d conditionals, want %d", conditionals, branches)
	}
}

func checkImpliedLogicalQueries(test *testing.T, operator string) {
	test.Helper()
	const operands = 256
	var source strings.Builder
	source.WriteString(
		"function box(n:number):{readonly value:number}|undefined{return n<0?undefined:{value:n};} const result = ",
	)
	for index := 0; index < operands; index++ {
		if index != 0 {
			source.WriteString(operator)
		}
		fmt.Fprintf(&source, "box(%d)", index)
	}
	source.WriteString("; console.log(String(result?.value ?? -1));")
	lowering, counts := lowerCountedImpliedTargets(test, source.String())
	logical := 0
	for node, count := range counts {
		if node.Kind != ast.KindBinaryExpression || node.AsBinaryExpression().OperatorToken.Kind != map[string]ast.Kind{
			"||": ast.KindBarBarToken,
			"&&": ast.KindAmpersandAmpersandToken,
			"??": ast.KindQuestionQuestionToken,
		}[operator] {
			continue
		}
		logical++
		if count != 1 ||
			lowering.checker.TypeToString(
				lowering.slots.implied[node],
			) != lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)) {
			test.Fatal("logical target changed or queried repeatedly")
		}
	}
	if logical < operands-1 {
		test.Fatalf("queried %d logical expressions, want at least %d", logical, operands-1)
	}
}
func TestImpliedTargetOrQueriesOnce(test *testing.T) {
	test.Parallel()
	checkImpliedLogicalQueries(test, "||")
}
func TestImpliedTargetAndQueriesOnce(test *testing.T) {
	test.Parallel()
	checkImpliedLogicalQueries(test, "&&")
}
func TestImpliedTargetNullishQueriesOnce(test *testing.T) {
	test.Parallel()
	checkImpliedLogicalQueries(test, "??")
}

func TestImpliedTargetTupleCachesNil(test *testing.T) {
	test.Parallel()
	lowering, counts := lowerCountedImpliedTargets(
		test,
		"const tuple: readonly [number,number] = [1,2]; console.log(String(tuple[0]));",
	)
	// Contextual tuple elements need not ask impliedTarget during the walk. Ask
	// both explicitly to prove a nil target is cached rather than queried twice.
	file := lowering.program.Files()[0]
	literal := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration().Initializer
	for _, element := range literal.AsArrayLiteralExpression().Elements.Nodes {
		if lowering.impliedTarget(element) != nil {
			test.Fatal("tuple unexpectedly acquired array element target")
		}
	}
	if counts[literal] != 1 {
		test.Fatalf("tuple queries = %d, want 1", counts[literal])
	}
}
