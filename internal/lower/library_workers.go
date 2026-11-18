package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
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

// Adamic #bxaz0p1: no calls, promises or return normalization until head 3b, whose first part (3b-1)
// admits one road, workersAwaitedGet's. The call and its callee each meet this rule, so both let it by.
func (l *lowering) workersMemberRefusal(n *ast.Node) error {
	if l.workersAwaitedGet(n) || (n.Parent != nil && n.Parent.Kind == ast.KindCallExpression && n.Parent.AsCallExpression().Expression == n && l.workersAwaitedGet(n.Parent)) {
		return nil
	}
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

// workersAwaitedGet reports whether a call is head 3b-1's one admitted road (Adamic #bxaz0p1): get
// called directly on a KV binding read as an identifier or a property access, which
// CheckWorkerEnvironment's slot rules prove came from the environment, never through ?., a
// parenthesized callee or a conditional or comma receiver, and the call the direct operand of await,
// parentheses allowed. A stored, chained or unawaited get stays rule 149's NotYet.
func (l *lowering) workersAwaitedGet(n *ast.Node) bool {
	if n.Kind != ast.KindCallExpression || n.Flags&ast.NodeFlagsOptionalChain != 0 {
		return false
	}
	access := n.AsCallExpression().Expression
	if access.Kind != ast.KindPropertyAccessExpression || access.AsPropertyAccessExpression().QuestionDotToken != nil || access.Name().Text() != "get" {
		return false
	}
	receiver := access.Expression()
	if (receiver.Kind != ast.KindIdentifier && receiver.Kind != ast.KindPropertyAccessExpression) || !load.WorkerKV(l.checker, l.checker.GetTypeAtLocation(receiver)) {
		return false
	}
	site := n
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	return site.Parent != nil && site.Parent.Kind == ast.KindAwaitExpression && site.Parent.Expression() == site
}

// workersCall lowers workersAwaitedGet's road to the binding's own get, as a host call that settles to
// text or null. Its Promise is the await's operand alone, and the value check rides the call
// (javascript's workersHostCall), so the await that settles it lowers as any other.
func (l *lowering) workersCall(node *ast.Node) (ir.Expression, bool, error) {
	if !l.workersAwaitedGet(node) {
		return nil, false, nil
	}
	const name = "workers:KvNamespace.get"
	call := node.AsCallExpression()
	if len(call.Arguments.Nodes) != 1 {
		return nil, true, l.notYet(node, name+" with other than one string key")
	}
	argument := call.Arguments.Nodes[0]
	if argument.Kind == ast.KindSpreadElement {
		return nil, true, l.notYet(argument, name+" with a spread argument")
	}
	if of, known := l.representation(l.checker.GetTypeAtLocation(argument)); !known || of != ir.String {
		return nil, true, l.notYet(argument, name+" with a key that isn't a string")
	}
	returns, known := l.representation(l.checker.GetTypeAtLocation(node))
	if !known || returns != ir.Promise {
		return nil, true, l.notYet(node, name+" whose result isn't the library Promise")
	}
	receiver, err := l.expression(call.Expression.Expression())
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(argument, name+" with a key that isn't a string")
	}
	return ir.NodeHostCall{Module: "workers:KvNamespace", Member: "get", Arguments: []ir.Expression{receiver, key}, Returns: returns, Throws: true}, true, nil
}
