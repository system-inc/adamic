package lower

import (
	"math"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Shape is the opt-in. Numerically identical implicit or arithmetic initializers stay closed.
func (lowering *lowering) flagEnum(symbol *ast.Symbol) bool {
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
		if initializer == nil || !lowering.flagInitializer(initializer, declaration, false) {
			return false
		}
		constant := lowering.checker.GetConstantValue(member)
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

func (lowering *lowering) flagLiteral(node *ast.Node, value float64) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindNumericLiteral {
		return false
	}
	literal := lowering.checker.GetTypeAtLocation(node)
	if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return false
	}
	number := reflect.ValueOf(literal.AsLiteralType().Value())
	return number.Kind() == reflect.Float64 && number.Float() == value
}

func (lowering *lowering) flagInitializer(node, declaration *ast.Node, memberAllowed bool) bool {
	node = ast.SkipParentheses(node)
	if member := lowering.enumMember(node); member != nil {
		return member.Parent == declaration
	}
	if lowering.flagLiteral(node, 0) && !memberAllowed {
		return true
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindLessThanLessThanToken:
		if memberAllowed || !lowering.flagLiteral(binary.Left, 1) {
			return false
		}
		for bit := 0; bit <= 30; bit++ {
			if lowering.flagLiteral(binary.Right, float64(bit)) {
				return true
			}
		}
	case ast.KindBarToken:
		return lowering.flagInitializer(binary.Left, declaration, true) &&
			lowering.flagInitializer(binary.Right, declaration, true)
	}
	return false
}

// A member slot still denotes that member. Only the complete enum type accepts combinations.
func (lowering *lowering) flagTarget(proven *checker.Type) *ast.Symbol {
	if proven == nil {
		return nil
	}
	proven = lowering.checker.GetNonNullableType(proven)
	if proven == nil || proven.Flags()&checker.TypeFlagsEnumLike == 0 {
		return nil
	}
	symbol := lowering.enumIdentity(proven)
	if !lowering.flagEnum(symbol) || proven != lowering.checker.GetTypeAtLocation(symbol.ValueDeclaration.Name()) {
		return nil
	}
	return symbol
}

// No runtime validation is needed: AND cannot add bits, while OR and XOR need both proofs.
// Complement, arithmetic and shifts never gain a proof, even if their current value fits.
func (lowering *lowering) flagDomain(node *ast.Node, target *ast.Symbol) bool {
	return lowering.flagDomainSeen(node, target, map[*ast.Node]bool{})
}

func (lowering *lowering) flagDomainSeen(node *ast.Node, target *ast.Symbol, seen map[*ast.Node]bool) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandToken:
			return lowering.flagDomainSeen(binary.Left, target, seen) ||
				lowering.flagDomainSeen(binary.Right, target, seen)
		case ast.KindBarToken:
			return lowering.flagDomainSeen(binary.Left, target, seen) &&
				lowering.flagDomainSeen(binary.Right, target, seen)
		case ast.KindCaretToken:
			return lowering.flagDomainSeen(binary.Left, target, seen) &&
				lowering.flagDomainSeen(binary.Right, target, seen)
		default:
			proven := lowering.checker.GetTypeAtLocation(node)
			return proven.Flags()&checker.TypeFlagsEnumLike != 0 && lowering.enumIdentity(proven) == target
		}
	}
	if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {
		return false
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		return lowering.flagDomainSeen(conditional.WhenTrue, target, seen) &&
			lowering.flagDomainSeen(conditional.WhenFalse, target, seen)
	}
	proven := lowering.checker.GetTypeAtLocation(node)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 && ast.IsIdentifier(node) {
		symbol := lowering.flagValueSymbol(node)
		if symbol != nil && symbol.ValueDeclaration != nil {
			declaration := symbol.ValueDeclaration
			if declaration.Kind == ast.KindVariableDeclaration &&
				declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
				!seen[declaration] {
				// A fresh inline iterable of primitives has no alias that can replace an element.
				// The const binding receives only those proven values, even if tsc infers number[].
				if loop := declaration.Parent.Parent; loop != nil && loop.Kind == ast.KindForOfStatement {
					iterable := ast.SkipParentheses(loop.AsForInOrOfStatement().Expression)
					if iterable.Kind == ast.KindArrayLiteralExpression {
						seen[declaration] = true
						defer delete(seen, declaration)
						for _, element := range iterable.AsArrayLiteralExpression().Elements.Nodes {
							if !lowering.flagDomainSeen(element, target, seen) {
								return false
							}
						}
						return true
					}
				}
				if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
					seen[declaration] = true
					result := lowering.flagDomainSeen(initializer, target, seen)
					delete(seen, declaration)
					return result
				}
			}
		}
	}
	return proven.Flags()&checker.TypeFlagsEnumLike != 0 && lowering.enumIdentity(proven) == target
}

func (lowering *lowering) flagWriteRefusal(node *ast.Node, target *checker.Type) error {
	return &Refused{
		Where: lowering.program.Where(node),
		What: "a number outside the proven flag domain assigned to " +
			lowering.checker.TypeToString(target) +
			"; its domain is a closed union of non-negative int32 bit subsets",
		Fix: "use this enum's members with | or ^ on two proven flags, or & with one proven flag; keep complement, arithmetic and shifts in a number (adamic/enum-flags)",
	}
}

func (lowering *lowering) flagUpdate(node, updated *ast.Node) bool {
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	symbol := lowering.symbol(updated)
	if symbol == nil {
		return false
	}
	target := lowering.flagTarget(lowering.checker.GetTypeOfSymbol(symbol))
	if target == nil {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandEqualsToken:
		return true
	case ast.KindBarEqualsToken, ast.KindCaretEqualsToken:
		return lowering.flagDomain(binary.Right, target)
	}
	return false
}

// Excluding declared members does not prove a flag is one remaining member: it may be a mask.
func (lowering *lowering) flagMemberWidened(node *ast.Node, contextual *checker.Type) *widening {
	target := lowering.checker.GetNonNullableType(contextual)
	if target.Flags()&checker.TypeFlagsEnumLike == 0 ||
		!lowering.flagEnum(lowering.enumIdentity(target)) ||
		lowering.flagTarget(target) != nil {
		return nil
	}
	if symbol := lowering.flagValueSymbol(node); symbol != nil {
		declared := lowering.checker.GetTypeOfSymbol(symbol)
		if lowering.flagTarget(declared) != nil {
			return &widening{source: declared, target: target, enum: true}
		}
	}
	return nil
}

// Shorthand names denote a field symbol at the syntax node, but the proof belongs to its value.
func (lowering *lowering) flagValueSymbol(node *ast.Node) *ast.Symbol {
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol := lowering.checker.GetShorthandAssignmentValueSymbol(node.Parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = lowering.checker.GetAliasedSymbol(symbol)
		}
		if symbol != nil {
			return lowering.checker.GetExportSymbolOfSymbol(symbol)
		}
		return nil
	}
	return lowering.symbol(node)
}
