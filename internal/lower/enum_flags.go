package lower

import (
	"math"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Shape is the opt-in. Numerically identical implicit or arithmetic initializers stay closed.
func (l *lowering) flagEnum(symbol *ast.Symbol) bool {
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindEnumDeclaration {
		return false
	}
	declaration := symbol.ValueDeclaration
	members := declaration.AsEnumDeclaration().Members.Nodes
	if len(members) == 0 {
		return false
	}
	for _, member := range members {
		initializer := member.AsEnumMember().Initializer
		if initializer == nil || !l.flagInitializer(initializer, declaration, false) {
			return false
		}
		constant := l.checker.GetConstantValue(member)
		if constant == nil || reflect.TypeOf(constant).Kind() != reflect.Float64 {
			return false
		}
		value := reflect.ValueOf(constant).Float()
		if value < 0 || value > math.MaxInt32 || value != math.Trunc(value) {
			return false
		}
	}
	return true
}

func (l *lowering) flagLiteral(node *ast.Node, value float64) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindNumericLiteral {
		return false
	}
	literal := l.checker.GetTypeAtLocation(node)
	if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return false
	}
	number := reflect.ValueOf(literal.AsLiteralType().Value())
	return number.Kind() == reflect.Float64 && number.Float() == value
}

func (l *lowering) flagInitializer(node, declaration *ast.Node, memberAllowed bool) bool {
	node = ast.SkipParentheses(node)
	if member := l.enumMember(node); member != nil {
		return member.Parent == declaration
	}
	if l.flagLiteral(node, 0) && !memberAllowed {
		return true
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindLessThanLessThanToken:
		if memberAllowed || !l.flagLiteral(binary.Left, 1) {
			return false
		}
		for bit := 0; bit <= 30; bit++ {
			if l.flagLiteral(binary.Right, float64(bit)) {
				return true
			}
		}
	case ast.KindBarToken:
		return l.flagInitializer(binary.Left, declaration, true) && l.flagInitializer(binary.Right, declaration, true)
	}
	return false
}

// A member slot still denotes that member. Only the complete enum type accepts combinations.
func (l *lowering) flagTarget(proven *checker.Type) *ast.Symbol {
	if proven == nil {
		return nil
	}
	proven = l.checker.GetNonNullableType(proven)
	if proven == nil || proven.Flags()&checker.TypeFlagsEnumLike == 0 {
		return nil
	}
	symbol := l.enumIdentity(proven)
	if !l.flagEnum(symbol) || proven != l.checker.GetTypeAtLocation(symbol.ValueDeclaration.Name()) {
		return nil
	}
	return symbol
}

// No runtime validation is needed: AND cannot add bits, while OR and XOR need both proofs.
// Complement, arithmetic and shifts never gain a proof, even if their current value fits.
func (l *lowering) flagDomain(node *ast.Node, target *ast.Symbol) bool {
	return l.flagDomainSeen(node, target, map[*ast.Node]bool{})
}

func (l *lowering) flagDomainSeen(node *ast.Node, target *ast.Symbol, seen map[*ast.Node]bool) bool {
	node = assignmentRight(node)
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandToken:
			return l.flagDomainSeen(binary.Left, target, seen) || l.flagDomainSeen(binary.Right, target, seen)
		case ast.KindBarToken:
			return l.flagDomainSeen(binary.Left, target, seen) && l.flagDomainSeen(binary.Right, target, seen)
		case ast.KindCaretToken:
			return l.flagDomainSeen(binary.Left, target, seen) && l.flagDomainSeen(binary.Right, target, seen)
		default:
			proven := l.checker.GetTypeAtLocation(node)
			return proven.Flags()&checker.TypeFlagsEnumLike != 0 && l.enumIdentity(proven) == target
		}
	}
	if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {
		return false
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		return l.flagDomainSeen(conditional.WhenTrue, target, seen) && l.flagDomainSeen(conditional.WhenFalse, target, seen)
	}
	proven := l.checker.GetTypeAtLocation(node)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 && ast.IsIdentifier(node) {
		symbol := l.flagValueSymbol(node)
		if symbol != nil && symbol.ValueDeclaration != nil {
			declaration := symbol.ValueDeclaration
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && !seen[declaration] {
				// A fresh inline iterable of primitives has no alias that can replace an element.
				// The const binding receives only those proven values, even if tsc infers number[].
				if loop := declaration.Parent.Parent; loop != nil && loop.Kind == ast.KindForOfStatement {
					iterable := ast.SkipParentheses(loop.AsForInOrOfStatement().Expression)
					if iterable.Kind == ast.KindArrayLiteralExpression {
						seen[declaration] = true
						defer delete(seen, declaration)
						for _, element := range iterable.AsArrayLiteralExpression().Elements.Nodes {
							if !l.flagDomainSeen(element, target, seen) {
								return false
							}
						}
						return true
					}
				}
				if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
					seen[declaration] = true
					result := l.flagDomainSeen(initializer, target, seen)
					delete(seen, declaration)
					return result
				}
			}
		}
	}
	return proven.Flags()&checker.TypeFlagsEnumLike != 0 && l.enumIdentity(proven) == target
}

func (l *lowering) flagWriteRefusal(node *ast.Node, target *checker.Type) error {
	return &Refused{Where: l.program.Where(node), What: "a number outside the proven flag domain assigned to " + l.checker.TypeToString(target) + "; its domain is a closed union of non-negative int32 bit subsets", Fix: "use this enum's members with | or ^ on two proven flags, or & with one proven flag; keep complement, arithmetic and shifts in a number (adamic/enum-flags)"}
}

// A literal shift through the sign bit is explicitly refused rather than silently opting out.
func (l *lowering) flagInitializerBound(node *ast.Node) error {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind == ast.KindBarToken {
		if err := l.flagInitializerBound(binary.Left); err != nil {
			return err
		}
		return l.flagInitializerBound(binary.Right)
	}
	if binary.OperatorToken.Kind == ast.KindLessThanLessThanToken && l.flagLiteral(binary.Left, 1) {
		right := ast.SkipParentheses(binary.Right)
		if right.Kind == ast.KindNumericLiteral {
			literal := l.checker.GetTypeAtLocation(right)
			value := reflect.ValueOf(literal.AsLiteralType().Value()).Float()
			if value < 0 || value > 30 || value != math.Trunc(value) {
				return &Refused{Where: l.program.Where(node), What: "a flag initializer outside the non-negative int32 bound", Fix: "use 1 << n with a literal n from 0 to 30 (adamic/enum-flags)"}
			}
		}
	}
	return nil
}

func (l *lowering) flagUpdate(node, updated *ast.Node) bool {
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	symbol := l.symbol(updated)
	if symbol == nil {
		return false
	}
	target := l.flagTarget(l.checker.GetTypeOfSymbol(symbol))
	if target == nil {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandEqualsToken:
		return true
	case ast.KindBarEqualsToken, ast.KindCaretEqualsToken:
		return l.flagDomain(binary.Right, target)
	}
	return false
}

// Excluding declared members does not prove a flag is one remaining member: it may be a mask.
func (l *lowering) flagMemberWidened(node *ast.Node, contextual *checker.Type) *widening {
	target := l.checker.GetNonNullableType(contextual)
	if target.Flags()&checker.TypeFlagsEnumLike == 0 || !l.flagEnum(l.enumIdentity(target)) || l.flagTarget(target) != nil {
		return nil
	}
	if symbol := l.flagValueSymbol(node); symbol != nil {
		declared := l.checker.GetTypeOfSymbol(symbol)
		if l.flagTarget(declared) != nil {
			return &widening{source: declared, target: target, enum: true}
		}
	}
	return nil
}

// Shorthand names denote a field symbol at the syntax node, but the proof belongs to its value.
func (l *lowering) flagValueSymbol(node *ast.Node) *ast.Symbol {
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol := l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = l.checker.GetAliasedSymbol(symbol)
		}
		if symbol != nil {
			return l.checker.GetExportSymbolOfSymbol(symbol)
		}
		return nil
	}
	return l.symbol(node)
}
