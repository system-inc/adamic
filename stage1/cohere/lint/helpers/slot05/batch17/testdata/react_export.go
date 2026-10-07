package react

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

var adamicTrace string

func adamicText(n *ast.Node) string   { adamicTrace += "text:1;"; return n.Text() }
func adamicHookName(name string) bool { adamicTrace += "hook:" + name + ";"; return IsHookName(name) }
func adamicNamespaced(n *ast.Node, p func(string) bool) bool {
	adamicTrace += "member:1;"
	return IsNamespacedMember(n, p)
}

type AdamicHookRow struct {
	HasCall, HasExpression, Identifier, Property, Hook, Namespaced bool
	Text                                                           string
}

func AdamicHookObserve(call *ast.CallExpression) (AdamicHookRow, string) {
	row := AdamicHookRow{HasCall: call != nil}
	if call != nil && call.Expression != nil {
		row.HasExpression = true
		e := call.Expression
		row.Identifier = e.Kind == ast.KindIdentifier
		row.Property = e.Kind == ast.KindPropertyAccessExpression
		if row.Identifier {
			row.Text = e.Text()
			row.Hook = IsHookName(row.Text)
		}
		if row.Property {
			row.Namespaced = IsNamespacedMember(e, IsHookName)
		}
	}
	adamicTrace = ""
	result := IsHookCall(call)
	return row, fmt.Sprintf("%t|%s\n", result, adamicTrace)
}
