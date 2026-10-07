package react

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
)

var adamicTrace string
var adamicIDs map[*ast.Node]int

func adamicID(n *ast.Node) int {
	if n == nil {
		return -1
	}
	if id, ok := adamicIDs[n]; ok {
		return id
	}
	id := len(adamicIDs)
	adamicIDs[n] = id
	return id
}
func adamicEs5(n *ast.Node) bool {
	adamicTrace += fmt.Sprintf("es5:%d;", adamicID(n))
	return utilsreact.IsEs5ComponentCall(n)
}
func adamicSkip(n *ast.Node) *ast.Node {
	adamicTrace += fmt.Sprintf("skip:%d;", adamicID(n))
	return ast.SkipParentheses(n)
}
func adamicText(n *ast.Node) string {
	adamicTrace += fmt.Sprintf("text:%d;", adamicID(n))
	return n.Text()
}

type AdamicStrictNode struct {
	ID                   int
	Identifier, Property bool
	Text                 string
}
type AdamicStrictRow struct {
	Root, Callee, Skipped, Name int
	Es5                         bool
	Nodes                       []AdamicStrictNode
}

func AdamicStrictObserve(n *ast.Node) (AdamicStrictRow, string) {
	adamicIDs = map[*ast.Node]int{nil: -1}
	row := AdamicStrictRow{Root: adamicID(n), Callee: -1, Skipped: -1, Name: -1, Es5: utilsreact.IsEs5ComponentCall(n), Nodes: []AdamicStrictNode{}}
	var raw, skipped, name *ast.Node
	if n != nil && n.Kind == ast.KindCallExpression {
		raw = n.AsCallExpression().Expression
		skipped = ast.SkipParentheses(raw)
		if skipped != nil && skipped.Kind == ast.KindPropertyAccessExpression {
			name = skipped.AsPropertyAccessExpression().Name()
		}
	}
	row.Callee = adamicID(raw)
	row.Skipped = adamicID(skipped)
	row.Name = adamicID(name)
	seen := map[*ast.Node]bool{}
	for _, node := range []*ast.Node{nil, n, raw, skipped, name} {
		if seen[node] {
			continue
		}
		seen[node] = true
		shape := AdamicStrictNode{ID: adamicID(node)}
		if node != nil {
			shape.Identifier = node.Kind == ast.KindIdentifier
			shape.Property = node.Kind == ast.KindPropertyAccessExpression
			if shape.Identifier {
				shape.Text = node.Text()
			}
		}
		row.Nodes = append(row.Nodes, shape)
	}
	adamicTrace = ""
	result := isEs5ComponentCallStrict(n)
	return row, fmt.Sprintf("%t|%s\n", result, adamicTrace)
}
