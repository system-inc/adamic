package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Refuse the witnessed divergent shapes before either backend inserts a stop.
// Follow literal keys through const aliases and direct call arguments, but do not
// reject unrelated index operations or the lane's own-key observation paths.
func (l *lowering) recordDivergenceRefusal(node *ast.Node) error {
	if node.Kind == ast.KindBinaryExpression {
		b := node.AsBinaryExpression()
		if b.OperatorToken.Kind == ast.KindInKeyword && l.recordElement(l.checker.GetTypeAtLocation(b.Right)) != nil {
			if key := l.recordPrototypeKey(b.Left, 0); key != "" && ast.IsIdentifier(ast.SkipParentheses(b.Left)) && !l.recordExplicitOwn(b.Right, key) {
				return &Refused{Where: l.program.Where(node), What: "record membership for " + key + ": the record's prototype chain isn't modeled", Fix: "use Object.hasOwn(record, key) for own membership, or a Map"}
			}
		}
		if b.OperatorToken.Kind == ast.KindEqualsToken && l.recordTarget(b.Left) {
			key := l.recordAccessPrototypeKey(b.Left)
			if key == "__proto__" && !l.recordExplicitOwn(recordReceiver(b.Left), key) {
				return &Refused{Where: l.program.Where(b.Left), What: "record assignment to __proto__: its inherited setter isn't modeled", Fix: "define an explicit computed own __proto__ data entry before assigning, or use a Map"}
			}
		}
	}
	if !l.recordTarget(node) || l.recordReadDiscarded(node) || l.recordReadOwnGuarded(node) {
		return nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		if parent.Kind == ast.KindDeleteExpression {
			return nil
		}
		if parent.Kind == ast.KindBinaryExpression {
			b := parent.AsBinaryExpression()
			if b.Left == at && b.OperatorToken.Kind == ast.KindEqualsToken {
				return nil
			}
			other := b.Right
			if other == at {
				other = b.Left
			}
			t := l.checker.GetTypeAtLocation(other)
			if (b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken) && t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 && !l.includesUndefined(t) {
				return nil
			}
		}
	}
	if key := l.recordAccessPrototypeKey(node); key != "" && node.Kind == ast.KindElementAccessExpression && ast.IsIdentifier(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)) && !l.recordExplicitOwn(recordReceiver(node), key) {
		return &Refused{Where: l.program.Where(node), What: "record read of " + key + ": the record's prototype chain isn't modeled", Fix: "check Object.hasOwn(record, key) immediately before reading its own value, or use a Map"}
	}
	// Only scalar reads whose narrowed slot has a visible intervening mutation
	// are refused. Runtime typeof/presence observations and reference errors keep
	// their existing semantics; unrelated calls do not invalidate this proof.
	if !recordTagObservation(node) && l.recordScalarNarrowed(node) && l.recordInterveningMutation(node) {
		return &Refused{Where: l.program.Where(node), What: "a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven", Fix: "snapshot the value before the call, or read and recheck its type after the call"}
	}
	return nil
}

func (l *lowering) recordAccessPrototypeKey(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		name := node.Name().Text()
		if recordPrototypeName(name) {
			return name
		}
		return ""
	}
	return l.recordPrototypeKey(node.AsElementAccessExpression().ArgumentExpression, 0)
}
func (l *lowering) recordPrototypeKey(node *ast.Node, depth int) string {
	if node == nil || depth > 12 {
		return ""
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		if recordPrototypeName(node.Text()) {
			return node.Text()
		}
		return ""
	}
	if !ast.IsIdentifier(node) {
		return ""
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return ""
	}
	d := symbol.Declarations[0]
	if d.Kind == ast.KindVariableDeclaration && d.Parent.Flags&ast.NodeFlagsConst != 0 {
		return l.recordPrototypeKey(d.AsVariableDeclaration().Initializer, depth+1)
	}
	if d.Kind != ast.KindParameter || d.Parent == nil {
		return ""
	}
	owner := d.Parent
	index := -1
	for i, p := range owner.Parameters() {
		if p == d {
			index = i
		}
	}
	if index < 0 {
		return ""
	}
	result := ""
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression {
			signature := l.checker.GetResolvedSignature(n)
			args := n.AsCallExpression().Arguments.Nodes
			if signature != nil && signature.Declaration() == owner && index < len(args) {
				if key := l.recordPrototypeKey(args[index], depth+1); key != "" {
					result = key
					return true
				}
			}
		}
		return n.ForEachChild(visit)
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if result != "" {
			break
		}
		if !file.IsDeclarationFile {
			file.AsNode().ForEachChild(visit)
		}
	}
	return result
}
func (l *lowering) recordExplicitOwn(receiver *ast.Node, key string) bool {
	receiver = ast.SkipParentheses(receiver)
	if ast.IsIdentifier(receiver) {
		s := l.symbol(receiver)
		if s == nil || len(s.Declarations) != 1 || s.Declarations[0].Kind != ast.KindVariableDeclaration {
			return false
		}
		receiver = s.Declarations[0].AsVariableDeclaration().Initializer
		if receiver == nil {
			return false
		}
		receiver = ast.SkipParentheses(receiver)
	}
	if receiver.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, p := range receiver.AsObjectLiteralExpression().Properties.Nodes {
		if p.Kind != ast.KindPropertyAssignment {
			continue
		}
		name := p.Name()
		if name.Kind == ast.KindComputedPropertyName {
			if l.recordPrototypeKey(name.AsComputedPropertyName().Expression, 0) == key {
				return true
			}
		} else if name.Text() == key && key != "__proto__" {
			return true
		}
	}
	return false
}
func (l *lowering) recordScalarNarrowed(node *ast.Node) bool {
	here := l.checker.GetTypeAtLocation(node)
	if here.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) == 0 {
		return false
	}
	element := l.recordElement(l.checker.GetTypeAtLocation(recordReceiver(node)))
	return element != nil && (element.Flags()&checker.TypeFlagsUnion != 0 || !l.includesUndefined(here))
}
func (l *lowering) recordInterveningMutation(node *ast.Node) bool {
	receiver := ast.SkipParentheses(recordReceiver(node))
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := l.symbol(receiver)
	var statement *ast.Node
	for at := node; at.Parent != nil; at = at.Parent {
		if at.Parent.Kind == ast.KindBlock {
			statement = at
			break
		}
		if ast.IsFunctionLike(at.Parent) {
			break
		}
	}
	if statement == nil {
		return false
	}
	found := false
	var mutation ast.Visitor
	mutation = func(n *ast.Node) bool {
		if n.Kind == ast.KindDeleteExpression && l.recordTarget(n.AsDeleteExpression().Expression) {
			if l.symbol(ast.SkipParentheses(recordReceiver(n.AsDeleteExpression().Expression))) == symbol && l.recordSameSlot(n.AsDeleteExpression().Expression, node) {
				found = true
			}
		}
		if n.Kind == ast.KindBinaryExpression && n.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && l.recordTarget(n.AsBinaryExpression().Left) {
			if l.symbol(ast.SkipParentheses(recordReceiver(n.AsBinaryExpression().Left))) == symbol && l.recordSameSlot(n.AsBinaryExpression().Left, node) {
				found = true
			}
		}
		return n.ForEachChild(mutation)
	}
	var calls ast.Visitor
	calls = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression {
			s := l.checker.GetResolvedSignature(n)
			if s != nil && s.Declaration() != nil && s.Declaration().Body() != nil {
				s.Declaration().Body().ForEachChild(mutation)
			}
		}
		return n.ForEachChild(calls)
	}
	for _, prior := range statement.Parent.AsBlock().Statements.Nodes {
		if prior == statement {
			break
		}
		prior.ForEachChild(calls)
		calls(prior)
	}
	return found
}

// Restrict invalidation to the same literal slot, rather than every mutation of
// its container. Unrelated record writes preserve a scalar narrowing.
func (l *lowering) recordSameSlot(a, b *ast.Node) bool {
	key := func(n *ast.Node) string {
		n = ast.SkipParentheses(n)
		if n.Kind == ast.KindPropertyAccessExpression {
			return n.Name().Text()
		}
		k := ast.SkipParentheses(n.AsElementAccessExpression().ArgumentExpression)
		if k.Kind == ast.KindStringLiteral || k.Kind == ast.KindNoSubstitutionTemplateLiteral {
			return k.Text()
		}
		return ""
	}
	x, y := key(a), key(b)
	return x != "" && x == y
}
