package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A served view can call RegExp's test with its receiver, or observe its presence, but cannot
// hand out the prototype function. A relation triggers the rule for the whole program; reads
// are judged after lowering, including ones through another view or earlier in the source.
type regexpServedRead struct {
	node     *ast.Node
	receiver *checker.Type
	class    viewRead
	allowed  bool
}

const regexpServedValueStop = "test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to"

func libraryServedRegExpTest(member *ast.Symbol) bool {
	if member.Name != "test" || !libraryServedMethods["RegExp"]["test"] {
		return false
	}
	for _, declaration := range member.Declarations {
		owner := declaration.Parent
		if declaration.Kind == ast.KindMethodSignature && owner != nil && owner.Kind == ast.KindInterfaceDeclaration && owner.Name().Text() == "RegExp" && load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return true
		}
	}
	return false
}

// A program with a served view refuses even an unrelated object's own test read as a value.
// This over-stop fails closed: a second structural view need not retain the first view's identity.
// Keep the receiver's declared type, not its instantiation or constraint: those can hide a view.
func (l *lowering) recordRegExpServedMember(node *ast.Node, receiver *checker.Type, expression ir.Expression, allowed bool) {
	if receiver == nil {
		return
	}
	l.regexpServedReads = append(l.regexpServedReads, regexpServedRead{node: node, receiver: receiver, class: viewReadOf(l.result, expression), allowed: allowed})
}

// Whole readers weigh types naming test; a direct named read also counts a record key.
func (l *lowering) recordRegExpServedWhole(node *ast.Node, receiver *checker.Type, expression ir.Expression, allowed bool) {
	if l.regexpServedHasTest(receiver) {
		l.recordRegExpServedMember(node, receiver, expression, allowed)
	}
}

func (l *lowering) regexpServedHasTest(receiver *checker.Type) bool {
	if receiver.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	return l.checker.GetPropertyOfType(receiver, "test") != nil
}

// Only the library's RegExp itself is exempt. A union or intersection, a type parameter,
// any and unknown do not prove that every value is typed by that library declaration.
func regexpServedReceiverIsRegExp(receiver *checker.Type) bool {
	if receiver.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection|checker.TypeFlagsInstantiable|checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return false
	}
	symbol := receiver.Symbol()
	if symbol == nil || symbol.Name != "RegExp" || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return false
		}
	}
	return true
}

// recordRegExpServedRead weighs member reads by viewReadOf, not another list of readers. A
// helper call made for element access can hide the property in its body; its unclassed value
// road fails closed. Whole-value readers are also weighed from wholeObjectCopies below.
func (l *lowering) recordRegExpServedRead(node *ast.Node, expression ir.Expression) {
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if node.Name().Text() != "test" {
			return
		}
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		if ast.IsAssignmentTarget(node) {
			return
		}
		key := l.checker.GetTypeAtLocation(node.AsElementAccessExpression().ArgumentExpression)
		if !regexpServedKeyMayTest(key) {
			return
		}
		receiver = node.AsElementAccessExpression().Expression
	default:
		if node.Kind == ast.KindCallExpression && viewReadOf(l.result, expression) != viewReadsNoField {
			for _, argument := range node.AsCallExpression().Arguments.Nodes {
				l.recordRegExpServedWhole(node, l.checker.GetTypeAtLocation(argument), expression, false)
			}
		}
		return
	}
	// A read by a computed key may name test even when its key is not a literal. No value road
	// through that receiver is served, so it needs no second key classifier here.
	var member ir.Expression
	walk(expression, func(value any) bool {
		if property, ok := value.(ir.Property); ok && property.Name == "test" {
			member = property
		}
		return true
	})
	if member == nil {
		member = expression
	}
	allowed := node.Kind == ast.KindPropertyAccessExpression && (called(node) || l.regexpServedPresence(node) || node.Parent != nil && node.Parent.Kind == ast.KindTypeOfExpression)
	l.recordRegExpServedMember(node, l.checker.GetTypeAtLocation(receiver), member, allowed)
}

// A finite key that cannot name test reads no served function; an unbounded key fails closed.
func regexpServedKeyMayTest(key *checker.Type) bool {
	if key.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range key.Types() {
			if regexpServedKeyMayTest(member) {
				return true
			}
		}
		return false
	}
	if key.Flags()&checker.TypeFlagsStringLiteral != 0 {
		return key.AsLiteralType().Value() == "test"
	}
	return key.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsESSymbolLike) == 0
}

// Presence consumes only a boolean answer. typeof keeps its existing presence-proof stop;
// comparisons other than the two undefined tests can observe the function's identity.
func (l *lowering) regexpServedPresence(node *ast.Node) bool {
	held := node
	for held.Parent != nil && held.Parent.Kind == ast.KindParenthesizedExpression {
		held = held.Parent
	}
	parent := held.Parent
	if parent == nil || parent.Kind == ast.KindTypeOfExpression {
		return false
	}
	if presenceConsumer(node) {
		return true
	}
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	other := binary.Left
	if other == held {
		other = binary.Right
	}
	return ast.IsIdentifier(other) && other.Text() == "undefined" && l.checker.GetTypeAtLocation(other).Flags()&checker.TypeFlagsUndefined != 0
}

// checkRegExpServedValues runs after lowering, so the relation's source order and the view's
// identity do not affect the judgment. viewReadOf classes every operand expression; the shared
// whole-value reader registry supplies spreads and copies without another list of readers.
func (l *lowering) checkRegExpServedValues() error {
	l.result.RegExpServedView = l.regexpServedView
	if !l.regexpServedView {
		return nil
	}
	for _, read := range l.regexpServedReads {
		if !regexpServedReceiverIsRegExp(read.receiver) && (!read.allowed || read.class == 0) {
			return l.notYet(read.node, regexpServedValueStop)
		}
	}
	for _, whole := range l.wholeObjectCopies {
		if l.regexpServedHasTest(whole.copied) && !regexpServedReceiverIsRegExp(whole.copied) {
			return l.notYet(whole.node, regexpServedValueStop)
		}
	}
	return nil
}

// A whole reader may already have stopped before it made an expression. Its unclassed
// candidate still belongs to this judgment when the relation has triggered the program rule.
// With no such relation, the original error and every admitted read are unchanged.
func (l *lowering) regexpServedPending(err error) error {
	if err != nil {
		if stop := l.checkRegExpServedValues(); stop != nil {
			return stop
		}
	}
	return err
}
