package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func workersMemberName(symbol *ast.Symbol) string {
	if symbol == nil || symbol.Parent == nil || symbol.Parent.Name != "KvNamespace" {
		return ""
	}
	for _, d := range symbol.Declarations {
		if !load.IsPrelude(ast.GetSourceFileOfNode(d)) {
			continue
		}
		for p := d.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindModuleDeclaration && p.Name().Text() == "adamic/workers" {
				return "workers:KvNamespace." + symbol.Name
			}
		}
	}
	return ""
}

// Adamic #bxaz0p1: no calls, promises or return normalization until head 3b.
func (l *lowering) workersMemberRefusal(n *ast.Node) error {
	if n.Kind == ast.KindCallExpression {
		n = ast.SkipParentheses(n.AsCallExpression().Expression)
	}
	if n.Kind != ast.KindPropertyAccessExpression && n.Kind != ast.KindElementAccessExpression {
		return nil
	}
	if !load.WorkerKV(l.checker, l.checker.GetTypeAtLocation(n.Expression())) {
		return nil
	}
	name := "a computed member"
	if n.Kind == ast.KindPropertyAccessExpression {
		name = n.Name().Text()
	} else if key := ast.SkipParentheses(n.AsElementAccessExpression().ArgumentExpression); key.Kind == ast.KindStringLiteral {
		name = key.Text()
	}
	return l.notYet(n, "workers:KvNamespace."+name+"; a Workers binding's calls wait on head 3b")
}
func init() { refusalRules.register(149, "workersMemberRefusal", (*lowering).workersMemberRefusal) }
