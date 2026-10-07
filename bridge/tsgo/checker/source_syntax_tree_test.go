package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestWave03SourceSyntaxTree(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.tsx")
	config := filepath.Join(directory, "tsconfig.json")
	source := `interface Props {enabled:boolean}; const C: React.FC<Props> = (p: Props, ref) => { class K extends React.Component {} let x = p.enabled ? 1 : 2; useMemo(async function*(a,...rest) { x += a; return (x!); }, [p.enabled,p[0]]); return <div/> };`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := program.Compiler.GetSourceFile(file)
	wire, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", "source-syntax-tree")
	if err != nil {
		t.Fatal(err)
	}
	values := decodedFields(t, wire)
	cursor := 2
	next := func() string {
		if cursor >= len(values) {
			t.Fatal("truncated syntax tree")
		}
		value := values[cursor]
		cursor++
		return value
	}
	integer := func() int {
		number, err := strconv.Atoi(next())
		if err != nil {
			t.Fatal(err)
		}
		return number
	}
	count := integer()
	var nodes []*ast.Node
	indices := map[*ast.Node]int{}
	parents := map[*ast.Node]*ast.Node{}
	var walk func(*ast.Node, *ast.Node)
	walk = func(node, parent *ast.Node) {
		nodes = append(nodes, node)
		indices[node] = len(nodes)
		parents[node] = parent
		node.ForEachChild(func(child *ast.Node) bool { walk(child, node); return false })
	}
	walk(sf.AsNode(), nil)
	if count != len(nodes) {
		t.Fatalf("count %d want %d", count, len(nodes))
	}
	for index, node := range nodes {
		kind, pos, end, start, text, parent := next(), integer(), integer(), integer(), next(), integer()
		if kind != strings.TrimPrefix(node.Kind.String(), "Kind") || pos != max(0, node.Pos()) || end != max(0, node.End()) || start != max(0, scanner.GetRangeOfTokenAtPosition(sf, max(0, node.Pos())).Pos()) || parent != indices[parents[node]] {
			t.Fatalf("node %d %s raw spans or ancestry differ", index, node.Kind)
		}
		if node.Kind == ast.KindIdentifier && text != node.Text() {
			t.Fatal("identifier text differs")
		}
		childCount := integer()
		var children []*ast.Node
		node.ForEachChild(func(child *ast.Node) bool { children = append(children, child); return false })
		if childCount != len(children) {
			t.Fatal("children count differs")
		}
		for _, child := range children {
			if integer() != indices[child] {
				t.Fatal("child identity differs")
			}
		}
		if integer() != indices[node.Name()] || integer() != indices[node.Type()] {
			t.Fatal("name/type identity differs")
		}
		initializer := integer()
		if node.Kind == ast.KindVariableDeclaration && initializer != indices[node.AsVariableDeclaration().Initializer] {
			t.Fatal("initializer identity differs")
		}
		if integer() != indices[node.Body()] || integer() != int(node.Flags) {
			t.Fatal("body/flags differ")
		}
		flags := integer()
		if ast.IsFunctionLike(node) && flags != int(ast.GetFunctionFlags(node)) {
			t.Fatal("function flags differ")
		}
		expression, left, right, yes, no, operator, rest := integer(), integer(), integer(), integer(), integer(), next(), next()
		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			if left != indices[binary.Left] || right != indices[binary.Right] || operator != strings.TrimPrefix(binary.OperatorToken.Kind.String(), "Kind") {
				t.Fatal("binary facts differ")
			}
		}
		if node.Kind == ast.KindConditionalExpression {
			conditional := node.AsConditionalExpression()
			if expression != indices[conditional.Condition] || yes != indices[conditional.WhenTrue] || no != indices[conditional.WhenFalse] {
				t.Fatal("conditional facts differ")
			}
		}
		if node.Kind == ast.KindParameter && (rest == "1") != (node.AsParameterDeclaration().DotDotDotToken != nil) {
			t.Fatal("rest differs")
		}
		arguments := integer()
		expectedArguments := 0
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Arguments != nil {
			expectedArguments = len(node.AsCallExpression().Arguments.Nodes)
		}
		if arguments != expectedArguments {
			t.Fatal("arguments count differs")
		}
		for n := 0; n < arguments; n++ {
			value := integer()
			if node.Kind != ast.KindCallExpression || value != indices[node.AsCallExpression().Arguments.Nodes[n]] {
				t.Fatal("call arguments differ")
			}
		}
		parameters := integer()
		expectedParameters := 0
		if ast.IsFunctionLike(node) && node.ParameterList() != nil {
			expectedParameters = len(node.ParameterList().Nodes)
		}
		if parameters != expectedParameters {
			t.Fatal("parameters count differs")
		}
		for n := 0; n < parameters; n++ {
			value := integer()
			if !ast.IsFunctionLike(node) || value != indices[node.ParameterList().Nodes[n]] {
				t.Fatal("parameters differ")
			}
		}
	}
	if cursor != len(values) {
		t.Fatal("trailing fields")
	}
	for _, question := range []string{"source-syntax-tree\nextra", "source-syntax-tree:unknown"} {
		if _, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatal("accepted malformed tree question")
		}
	}
	identifier := nodes[2]
	var output fields
	if _, err := program.sourceSyntaxTree(&output, identifier, "source-syntax-tree"); err == nil {
		t.Fatal("accepted non-source node")
	}
	t.Logf("%d raw syntax nodes checked", len(nodes))
}
