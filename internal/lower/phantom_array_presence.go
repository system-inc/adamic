package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// An array brand is phantom only while every observation of its member's presence is refused
// (rulings.md:31): a value of SortedArray<T> is held as its base array (unbranded.go), where the brand
// member is never present, so `" __sortedArrayBrand" in x` answers false while the checker narrows on
// it. Over `SortedArray<number> | { n: number }` the checker then takes the brand's arm and a function
// typed number returns undefined on Node (Language's b09, #gb0fqaa). So each road that observes it is
// refused at its node, judged on the checker's own type of the value observed, before any unbranding and
// before lowering's own handling of that node, by the design Diagnostics ruled on #gb0fqaa (Nov 5):
//  1. `k in x`, where k is a literal naming a brand member, or isn't a literal.
//  2. `Object.hasOwn(x, k)` and `x.hasOwnProperty(k)`, with the same key rule.
//  3. `Object.keys`, `values`, `entries`, `getOwnPropertyNames`, and `Object.assign`'s every argument.
//  4. `JSON.stringify(x)`.
//  5. An object spread `{ ...x }`. An array spread `[...x]` iterates elements, which observes nothing.
//  6. A write to a brand member: an element write whose key is a literal naming one, or isn't a literal
//     and isn't a number, and a property write naming one.
//
// Left to Language: a value that reaches the observation through a type parameter, an intersection,
// unknown or any, and the roads not listed (propertyIsEnumerable, Reflect, for-in, delete, a write
// through a destructuring pattern).
func (l *lowering) phantomArrayPresenceRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindInKeyword {
			return l.phantomPresenceKeyed(node, binary.Right, binary.Left, "an `in` test")
		}
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return l.phantomPresenceWrite(node, ast.SkipParentheses(binary.Left))
		}
	case ast.KindSpreadAssignment:
		return l.phantomPresenceWhole(node, node.AsSpreadAssignment().Expression, "an object spread")
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return nil
		}
		access := callee.AsPropertyAccessExpression()
		arguments := call.Arguments.Nodes
		name := access.Name().Text()
		switch {
		case l.isLibraryGlobal(access.Expression, "Object") && name == "hasOwn" && len(arguments) == 2:
			return l.phantomPresenceKeyed(node, arguments[0], arguments[1], "Object.hasOwn")
		case name == "hasOwnProperty" && len(arguments) == 1 && l.libraryMember(callee):
			return l.phantomPresenceKeyed(node, access.Expression, arguments[0], "hasOwnProperty")
		case l.isLibraryGlobal(access.Expression, "Object") && (name == "keys" || name == "values" || name == "entries" || name == "getOwnPropertyNames" || name == "assign"):
			for _, argument := range arguments {
				if err := l.phantomPresenceWhole(node, argument, "Object."+name); err != nil {
					return err
				}
			}
		case l.isLibraryGlobal(access.Expression, "JSON") && name == "stringify" && len(arguments) > 0:
			return l.phantomPresenceWhole(node, arguments[0], "JSON.stringify")
		}
	}
	return nil
}

func init() {
	refusalRules.register(450, "phantomArrayPresenceRefusal", (*lowering).phantomArrayPresenceRefusal)
}

// arrayBrandIn is the phantom extension of an array that t is, or has as a union arm, and the phantom
// members it adds over the array it brands, through every brand between them; or nil.
func (l *lowering) arrayBrandIn(t *checker.Type) (*checker.Type, []string) {
	arms := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		arms = t.Types()
	}
	for _, arm := range arms {
		base, members := l.phantomExtension(arm)
		for base != nil {
			next, more := l.phantomExtension(base)
			if next == nil {
				break
			}
			base, members = next, append(slices.Clip(members), more...)
		}
		if base != nil && l.arrayLikeType(base) {
			return arm, members
		}
	}
	return nil, nil
}

// phantomPresenceKeyed refuses a keyed test of value's presence (road 1 or 2): a key that's a literal
// naming a brand member, or a key that isn't a literal, which may name one.
func (l *lowering) phantomPresenceKeyed(node, value, key *ast.Node, road string) error {
	brand, members := l.arrayBrandIn(l.checker.GetTypeAtLocation(value))
	if brand == nil {
		return nil
	}
	key = ast.SkipParentheses(key)
	switch key.Kind {
	case ast.KindPrivateIdentifier:
		return nil
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
		if !slices.Contains(members, key.Text()) {
			return nil
		}
		return l.phantomPresenceRefused(node, brand, []string{key.Text()}, road)
	}
	return l.phantomPresenceRefused(node, brand, members, road)
}

// phantomPresenceWhole refuses a road that observes every own member of value at once (3 to 5).
func (l *lowering) phantomPresenceWhole(node, value *ast.Node, road string) error {
	brand, members := l.arrayBrandIn(l.checker.GetTypeAtLocation(value))
	if brand == nil {
		return nil
	}
	return l.phantomPresenceRefused(node, brand, members, road)
}

// phantomPresenceWrite refuses a write that makes a brand member present (road 6): an element write
// whose key is a literal naming one, or a key that's neither a literal nor a number, and a property write
// naming one. A write at a number is an element's.
func (l *lowering) phantomPresenceWrite(node, target *ast.Node) error {
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		brand, members := l.arrayBrandIn(l.checker.GetTypeAtLocation(access.Expression))
		if brand == nil || !slices.Contains(members, access.Name().Text()) {
			return nil
		}
		return l.phantomPresenceRefused(node, brand, []string{access.Name().Text()}, "a write")
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		brand, members := l.arrayBrandIn(l.checker.GetTypeAtLocation(access.Expression))
		if brand == nil {
			return nil
		}
		key := ast.SkipParentheses(access.ArgumentExpression)
		switch {
		case key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral:
			if !slices.Contains(members, key.Text()) {
				return nil
			}
			return l.phantomPresenceRefused(node, brand, []string{key.Text()}, "a write")
		case l.checker.GetTypeAtLocation(key).Flags()&checker.TypeFlagsNumberLike != 0:
			return nil
		}
		return l.phantomPresenceRefused(node, brand, members, "a write")
	}
	return nil
}

// phantomPresenceRefused is the refusal, in the words Diagnostics gave (#gb0fqaa).
func (l *lowering) phantomPresenceRefused(node *ast.Node, brand *checker.Type, members []string, road string) error {
	return &Refused{Where: l.program.Where(node), What: "an observation of " + l.checker.TypeToString(brand) + "'s phantom member " + strings.Join(members, ", ") + " (" + road + ")", Fix: "a phantom brand is never present at run time, so test or copy the value as its base array (`readonly T[]`), or test the other arm's own field (rulings.md:31)"}
}
