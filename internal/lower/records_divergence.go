package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Refuse the witnessed divergent shapes before either backend inserts a stop.
// Follow literal keys through const aliases and direct call arguments, but do not
// reject unrelated index operations or the lane's own-key observation paths.
func (lowering *lowering) recordDivergenceRefusal(node *ast.Node) error {
	if node.Kind == ast.KindBinaryExpression {
		expression := node.AsBinaryExpression()
		if expression.OperatorToken.Kind == ast.KindInKeyword &&
			lowering.recordElement(lowering.checker.GetTypeAtLocation(expression.Right)) != nil {
			if key := lowering.recordPrototypeKey(
				expression.Left,
				0,
			); key != "" &&
				ast.IsIdentifier(ast.SkipParentheses(expression.Left)) &&
				!lowering.recordExplicitOwn(expression.Right, key) {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "record membership for " + key + ": the record's prototype chain isn't modeled",
					Fix:   "use Object.hasOwn(record, key) for own membership, or a Map",
				}
			}
		}
		if expression.OperatorToken.Kind == ast.KindEqualsToken && lowering.recordTarget(expression.Left) {
			key := lowering.recordAccessPrototypeKey(expression.Left)
			if key == "__proto__" && !lowering.recordExplicitOwn(recordReceiver(expression.Left), key) {
				return &Refused{
					Where: lowering.program.Where(expression.Left),
					What:  "record assignment to __proto__: its inherited setter isn't modeled",
					Fix:   "define an explicit computed own __proto__ data entry before assigning, or use a Map",
				}
			}
		}
	}
	if !lowering.recordReadTarget(node) || lowering.recordReadDiscarded(node) || lowering.recordReadOwnGuarded(node) {
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
			expression := parent.AsBinaryExpression()
			if expression.Left == at && expression.OperatorToken.Kind == ast.KindEqualsToken {
				return nil
			}
			other := expression.Right
			if other == at {
				other = expression.Left
			}
			subjectType := lowering.checker.GetTypeAtLocation(other)
			if (expression.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
				expression.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken) &&
				subjectType.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 &&
				!lowering.includesUndefined(subjectType) {
				return nil
			}
		}
	}
	if key := lowering.recordAccessPrototypeKey(node); key != "" &&
		node.Kind == ast.KindElementAccessExpression &&
		ast.IsIdentifier(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)) &&
		!lowering.recordExplicitOwn(recordReceiver(node), key) {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "record read of " + key + ": the record's prototype chain isn't modeled",
			Fix:   "check Object.hasOwn(record, key) immediately before reading its own value, or use a Map",
		}
	}
	// Only scalar reads whose narrowed slot has a visible intervening mutation
	// are refused. Runtime typeof/presence observations and reference errors keep
	// their existing semantics; unrelated calls do not invalidate this proof.
	if !recordTagObservation(node) && lowering.recordScalarNarrowed(node) && lowering.recordInterveningMutation(node) {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven",
			Fix:   "snapshot the value before the call, or read and recheck its type after the call",
		}
	}
	return nil
}

func init() {
	refusalRules.register(1000, "recordDivergenceRefusal", (*lowering).recordDivergenceRefusal)
}

func (lowering *lowering) recordAccessPrototypeKey(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		name := node.Name().Text()
		if recordPrototypeName(name) {
			return name
		}
		return ""
	}
	return lowering.recordPrototypeKey(node.AsElementAccessExpression().ArgumentExpression, 0)
}
func (lowering *lowering) recordPrototypeKey(node *ast.Node, depth int) string {
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
	symbol := lowering.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return ""
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
		return lowering.recordPrototypeKey(declaration.AsVariableDeclaration().Initializer, depth+1)
	}
	if declaration.Kind != ast.KindParameter || declaration.Parent == nil {
		return ""
	}
	owner := declaration.Parent
	index := -1
	for parameterIndex, parameter := range owner.Parameters() {
		if parameter == declaration {
			index = parameterIndex
		}
	}
	if index < 0 {
		return ""
	}
	result := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			signature := lowering.checker.GetResolvedSignature(node)
			args := node.AsCallExpression().Arguments.Nodes
			if signature != nil && signature.Declaration() == owner && index < len(args) {
				if key := lowering.recordPrototypeKey(args[index], depth+1); key != "" {
					result = key
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range lowering.program.CompilerProgram().GetSourceFiles() {
		if result != "" {
			break
		}
		if !file.IsDeclarationFile {
			file.AsNode().ForEachChild(visit)
		}
	}
	return result
}
func (lowering *lowering) recordExplicitOwn(receiver *ast.Node, key string) bool {
	receiver = ast.SkipParentheses(receiver)
	if ast.IsIdentifier(receiver) {
		symbol := lowering.symbol(receiver)
		if symbol == nil ||
			len(symbol.Declarations) != 1 ||
			symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
			return false
		}
		receiver = symbol.Declarations[0].AsVariableDeclaration().Initializer
		if receiver == nil {
			return false
		}
		receiver = ast.SkipParentheses(receiver)
	}
	if receiver.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, node := range receiver.AsObjectLiteralExpression().Properties.Nodes {
		if node.Kind != ast.KindPropertyAssignment {
			continue
		}
		name := node.Name()
		if name.Kind == ast.KindComputedPropertyName {
			if lowering.recordPrototypeKey(name.AsComputedPropertyName().Expression, 0) == key {
				return true
			}
		} else if name.Text() == key && key != "__proto__" {
			return true
		}
	}
	return false
}
func (lowering *lowering) recordScalarNarrowed(node *ast.Node) bool {
	here := lowering.checker.GetTypeAtLocation(node)
	if here.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) == 0 {
		return false
	}
	element := lowering.recordElement(lowering.checker.GetTypeAtLocation(recordReceiver(node)))
	return element != nil && (element.Flags()&checker.TypeFlagsUnion != 0 || !lowering.includesUndefined(here))
}
func (lowering *lowering) recordInterveningMutation(node *ast.Node) bool {
	receiver := ast.SkipParentheses(recordReceiver(node))
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := lowering.symbol(receiver)
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
		if ast.IsFunctionLike(n) {
			return false
		}
		if n.Kind == ast.KindDeleteExpression && lowering.recordTarget(n.AsDeleteExpression().Expression) {
			if lowering.symbol(ast.SkipParentheses(recordReceiver(n.AsDeleteExpression().Expression))) == symbol &&
				lowering.recordSameSlot(n.AsDeleteExpression().Expression, node) {
				found = true
			}
		}
		if n.Kind == ast.KindBinaryExpression &&
			n.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken &&
			lowering.recordTarget(n.AsBinaryExpression().Left) {
			if lowering.symbol(ast.SkipParentheses(recordReceiver(n.AsBinaryExpression().Left))) == symbol &&
				lowering.recordSameSlot(n.AsBinaryExpression().Left, node) {
				found = true
			}
		}
		return n.ForEachChild(mutation)
	}
	var calls ast.Visitor
	calls = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindCallExpression {
			signature := lowering.checker.GetResolvedSignature(node)
			if signature != nil && signature.Declaration() != nil && signature.Declaration().Body() != nil {
				signature.Declaration().Body().ForEachChild(mutation)
			}
		}
		return node.ForEachChild(calls)
	}
	for _, prior := range statement.Parent.AsBlock().Statements.Nodes {
		if prior == statement {
			break
		}
		calls(prior)
	}
	return found
}

// Restrict invalidation to the same literal slot, rather than every mutation of
// its container. Unrelated record writes preserve a scalar narrowing.
func (lowering *lowering) recordSameSlot(left, right *ast.Node) bool {
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
	x, y := key(left), key(right)
	return x != "" && x == y
}
