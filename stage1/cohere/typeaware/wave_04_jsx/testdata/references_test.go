package react

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"strings"
	"testing"
)

func TestWave04JsxReferences(t *testing.T) {
	// Not parallel: the isolated oracle writes one caller-supplied artifact.
	f := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	type entry struct {
		Kind           int
		Text           string
		Receiver, Name int
		Arguments      []int
	}
	nodes := []*ast.Node{}
	entries := []entry{}
	add := func(n *ast.Node, text string, receiver, name int) int {
		nodes = append(nodes, n)
		entries = append(entries, entry{Kind: int(n.Kind), Text: text, Receiver: receiver, Name: name})
		return len(nodes) - 1
	}
	for _, name := range []string{"div", "App", "app", "_foo", "$foo", "테스트", "Foo-bar", "this", "React", "Fragment", "Provider", "createContext", "useMemo", "useCallback", "Other"} {
		add(f.NewIdentifier(name), name, -1, -1)
	}
	add(f.NewToken(ast.KindThisKeyword), "this", -1, -1)
	pairs := [][2]int{{2, 1}, {2, 0}, {8, 9}, {8, 10}, {8, 11}, {14, 11}, {15, 1}, {16, 9}, {0, 10}, {8, 12}, {8, 13}}
	for _, pair := range pairs {
		add(f.NewPropertyAccessExpression(nodes[pair[0]], nil, nodes[pair[1]], 0), "", pair[0], pair[1])
	}
	add(f.NewJsxNamespacedName(nodes[1], nodes[14]), "", 1, 14)
	add(f.NewParenthesizedExpression(nodes[8]), "", 8, -1)
	add(f.NewPropertyAccessExpression(nodes[28], nil, nodes[11], 0), "", 28, 11)
	add(f.NewParenthesizedExpression(nodes[29]), "", 29, -1)
	add(f.NewParenthesizedExpression(nodes[28]), "", 28, -1)
	require := add(f.NewIdentifier("require"), "require", -1, -1)
	react := add(f.NewStringLiteral("react", 0), "react", -1, -1)
	preact := add(f.NewStringLiteral("preact", 0), "preact", -1, -1)
	template := add(f.NewNoSubstitutionTemplateLiteral("react", 0), "react", -1, -1)
	for _, arguments := range [][]int{{react}, {preact}, {}, {react, preact}, {template}} {
		list := []*ast.Node{}
		for _, index := range arguments {
			list = append(list, nodes[index])
		}
		id := add(f.NewCallExpression(nodes[require], nil, nil, f.NewNodeList(list), 0), "", require, -1)
		entries[id].Arguments = arguments
	}
	id := add(f.NewCallExpression(nodes[14], nil, nil, f.NewNodeList([]*ast.Node{nodes[react]}), 0), "", 14, -1)
	entries[id].Arguments = []int{react}

	valueName := add(f.NewIdentifier("value"), "value", -1, -1)
	otherName := add(f.NewIdentifier("Value"), "Value", -1, -1)
	expression := add(f.NewJsxExpression(nil, nodes[react]), "", react, -1)
	emptyExpression := add(f.NewJsxExpression(nil, nil), "", -1, -1)
	value := add(f.NewJsxAttribute(nodes[valueName], nodes[expression]), "", expression, valueName)
	other := add(f.NewJsxAttribute(nodes[otherName], nodes[expression]), "", expression, otherName)
	boolean := add(f.NewJsxAttribute(nodes[valueName], nil), "", -1, valueName)
	literal := add(f.NewJsxAttribute(nodes[valueName], nodes[react]), "", react, valueName)
	empty := add(f.NewJsxAttribute(nodes[valueName], nodes[emptyExpression]), "", emptyExpression, valueName)
	spread := add(f.NewJsxSpreadAttribute(nodes[react]), "", react, -1)
	for _, properties := range [][]int{{}, {value}, {other}, {boolean}, {literal}, {empty}, {spread, value}, {other, value}, {boolean, value}, {literal, value}, {value, boolean}, {empty, value}} {
		list := []*ast.Node{}
		for _, index := range properties {
			list = append(list, nodes[index])
		}
		id := add(f.NewJsxAttributes(f.NewNodeList(list)), "", -1, -1)
		entries[id].Arguments = properties
	}

	var expected strings.Builder
	for _, n := range nodes {
		ref := resolvableJsxReference(n)
		index := -1
		for i, candidate := range nodes {
			if candidate == ref {
				index = i
				break
			}
		}
		fmt.Fprintf(&expected, "reference %d\n", index)
		// Structural routes need no checker; identifier routes are deliberately not exercised here.
		if n.Kind != ast.KindIdentifier {
			fmt.Fprintf(&expected, "fragment %t\n", jsxFragmentsNameIsFragment(rule.Context{}, n))
		}
		fmt.Fprintf(&expected, "factory %t\n", jsxNoConstructedContextValuesIsCreateContextCallee(n))
		fmt.Fprintf(&expected, "source %t\n", jsxFragmentsInitializerIsFragmentSource(n))
		value := jsxNoConstructedContextValuesValueExpression(n)
		valueIndex := -1
		for i, candidate := range nodes {
			if candidate == value {
				valueIndex = i
				break
			}
		}
		fmt.Fprintf(&expected, "value %d\n", valueIndex)
	}
	data := struct {
		Nodes    []entry
		Expected string
	}{entries, expected.String()}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_JSX_REFERENCES"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
}
