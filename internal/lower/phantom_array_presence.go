package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// An object or array brand is phantom only while every observation of its member's presence is refused
// (rulings.md:31). Object extensions and branded supertypes need the same guard (#3hsrk2t):
// a false presence test can narrow a live object to never, allowing an unchecked layout read.
// A value of SortedArray<T> is held as its base array (unbranded.go), where the brand
// member is never present, so `" __sortedArrayBrand" in x` answers false while the checker narrows on
// it. Over `SortedArray<number> | { n: number }` the checker then takes the brand's arm and a function
// typed number returns undefined on Node (Language's b09, #gb0fqaa). So each road that observes it is
// refused at its node, judged on the checker's own type of the value observed, before any unbranding and
// before lowering's own handling of that node, by the design Diagnostics ruled on #gb0fqaa (Nov 5):
//  1. `k in x`, where k is a literal naming a brand member, or isn't a literal.
//  2. `Object.hasOwn(x, k)`, with the same key rule; and any method call on x that isn't one of its base
//     array's own members (by the checker's symbol, never by name) and takes a property key, a string or
//     symbol argument. Object brands refuse only Object.prototype.hasOwnProperty and
//     propertyIsEnumerable, by symbol; their real methods remain real operations.
//     Object prototype methods use the same literal-key rule as Object.hasOwn.
//     Arrays retain the broader fail-closed method rule (#gb0fqaa).
//  3. `Object.keys`, `values`, `entries`, `getOwnPropertyNames`, and `Object.assign`'s every argument.
//  4. `JSON.stringify(x)`.
//  5. An object spread `{ ...x }`. An array spread `[...x]` iterates elements, which observes nothing.
//  6. A write to a brand member: an element write whose key is a literal naming one, or isn't a literal
//     and may name one, and a property write naming one. Numeric array writes remain element writes.
//  7. An object brand's phantom field read or binding. The base has no stored field, so
//     these otherwise reach a checked read that stops at run time. Array reads keep their existing handling.
//
// A type parameter is judged by its constraint (phantomBrandIn), so b09's lie reached generically is refused
// here too (Language's l13 and l14). Erased observations through `T extends object`, unknown or `object` see
// no brand and answer the truth, false, and the checker never narrows a branded type from them. The roads
// still held only by later stops (for-in, Reflect, destructuring the brand member, delete, Object.keys
// through object or a generic) are pinned with those stops in TestUnbrandedPresenceHeldElsewhere.
func (l *lowering) phantomPresenceRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindInKeyword {
			return l.phantomPresenceKeyed(node, binary.Right, binary.Left, "an `in` test")
		}
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			target := ast.SkipParentheses(binary.Left)
			if binary.OperatorToken.Kind == ast.KindEqualsToken && target.Kind == ast.KindObjectLiteralExpression {
				for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
					if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
						continue
					}
					name := property.Name()
					if name != nil && (name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral || name.Kind == ast.KindNumericLiteral) {
						if err := l.phantomObjectRead(property, binary.Right, l.checker.GetTypeAtLocation(binary.Right), name.Text()); err != nil {
							return err
						}
					}
				}
			}
			return l.phantomPresenceWrite(node, target)
		}
	case ast.KindPropertyAccessExpression:
		if l.isExpression(node) && !ast.IsAssignmentTarget(node) {
			access := node.AsPropertyAccessExpression()
			return l.phantomObjectRead(node, access.Expression, l.checker.GetTypeAtLocation(access.Expression), access.Name().Text())
		}
	case ast.KindBindingElement:
		if node.Parent != nil && node.Parent.Kind == ast.KindObjectBindingPattern && node.AsBindingElement().DotDotDotToken == nil {
			name := node.AsBindingElement().PropertyName
			if name == nil {
				name = node.Name()
			}
			declaration := node.Parent.Parent
			// A signature's parameter pattern reads nothing at run time.
			if declaration != nil && declaration.Kind == ast.KindParameter && declaration.Parent.Body() == nil {
				return nil
			}
			var origin *ast.Node
			if declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
				origin = declaration.Initializer()
			}
			if name != nil && (name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral || name.Kind == ast.KindNumericLiteral) {
				return l.phantomObjectRead(node, origin, l.checker.GetTypeAtLocation(node.Parent), name.Text())
			}
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
		case l.isLibraryGlobal(access.Expression, "Object") && (name == "keys" || name == "values" || name == "entries" || name == "getOwnPropertyNames" || name == "assign"):
			for _, argument := range arguments {
				if err := l.phantomPresenceWhole(node, argument, "Object."+name); err != nil {
					return err
				}
			}
		case l.isLibraryGlobal(access.Expression, "JSON") && name == "stringify" && len(arguments) > 0:
			return l.phantomPresenceWhole(node, arguments[0], "JSON.stringify")
		case l.holdsArrayBrand(access.Expression) && !l.arrayOwnMember(callee) && slices.ContainsFunc(arguments, l.propertyKeyArgument):
			return l.phantomPresenceWhole(node, access.Expression, name)
		case l.objectPrototypeKeyMember(callee) && len(arguments) > 0:
			// Object keys include numbers. The array road above deliberately
			// keeps its element-key exception, so select only object brands here.
			brand, _ := l.phantomBrandMatching(l.checker.GetTypeAtLocation(access.Expression), func(base *checker.Type, members []string) bool { return !l.phantomArrayBase(base, members) })
			if brand != nil {
				return l.phantomPresenceKeyed(node, access.Expression, arguments[0], name)
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(450, "phantomPresenceRefusal", (*lowering).phantomPresenceRefusal)
}

// phantomBrandIn finds an extension or branded supertype in t, its union arms or
// a type parameter's constraint. The returned names are its phantom members.
func (l *lowering) phantomBrandIn(t *checker.Type) (*checker.Type, []string) {
	return l.phantomBrandMatching(t, nil)
}

// A keyed observation must find the arm that owns its key, rather than let an
// earlier brand with another name hide it. Array methods select array bases only.
func (l *lowering) phantomBrandMatching(t *checker.Type, match func(*checker.Type, []string) bool) (*checker.Type, []string) {
	var arms []*checker.Type
	var add func(t *checker.Type, depth int)
	add = func(t *checker.Type, depth int) {
		switch {
		case t == nil || depth > 8:
		case t.Flags()&checker.TypeFlagsUnion != 0:
			for _, part := range t.Types() {
				add(part, depth+1)
			}
		case t.Flags()&checker.TypeFlagsTypeParameter != 0:
			// Judged by its constraint: `T extends SortedArray<number> | { n: number }` narrows on the brand.
			add(l.checker.GetBaseConstraintOfType(t), depth+1)
		default:
			arms = append(arms, t)
		}
	}
	add(t, 0)
	for _, arm := range arms {
		base, members := l.phantomExtension(arm)
		for base != nil {
			next, more := l.phantomExtension(base)
			if next == nil {
				break
			}
			base, members = next, append(slices.Clip(members), more...)
		}
		// Nominal classes keep real fields, as brandedMembership also requires.
		supertype := arm.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 && !isClassInstance(arm) && l.brandedSupertype(arm)
		if supertype {
			for _, property := range l.checker.GetPropertiesOfType(arm) {
				if property.Flags&ast.SymbolFlagsOptional == 0 && !checker.IsKnownSymbol(property) && phantomField(l.checker.GetTypeOfSymbol(property)) && !slices.Contains(members, property.Name) {
					members = append(members, property.Name)
				}
			}
		}
		if (base != nil || supertype) && (match == nil || match(base, members)) {
			return arm, members
		}
	}
	return nil, nil
}

// phantomPresenceKeyed refuses a keyed test of value's presence (road 1 or 2): a key that's a literal
// naming a brand member, or a key that isn't a literal, which may name one.
func (l *lowering) phantomPresenceKeyed(node, value, key *ast.Node, road string) error {
	key = ast.SkipParentheses(key)
	if key.Kind == ast.KindPrivateIdentifier {
		return nil
	}
	var match func(*checker.Type, []string) bool
	literal := key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral
	if literal {
		match = func(_ *checker.Type, members []string) bool { return slices.Contains(members, key.Text()) }
	}
	brand, members := l.phantomBrandMatching(l.checker.GetTypeAtLocation(value), match)
	if brand == nil {
		return nil
	}
	if literal {
		members = []string{key.Text()}
	}
	return l.phantomPresenceRefused(node, brand, members, road)
}

// Only object brands take the read road. Their checked field reads require a
// stored member that a phantom extension or supertype deliberately does not hold.
// A const literal explicitly carrying the field already reads correctly and keeps doing so.
func (l *lowering) phantomObjectRead(node, origin *ast.Node, receiver *checker.Type, field string) error {
	brand, _ := l.phantomBrandMatching(receiver, func(base *checker.Type, members []string) bool {
		return !l.phantomArrayBase(base, members) && slices.Contains(members, field)
	})
	if brand == nil {
		return nil
	}
	// An optional read already maps an absent member to undefined. A literal
	// explicitly carrying the field already has the storage its read requires.
	if member := l.checker.GetPropertyOfType(receiver, field); member != nil && member.Flags&ast.SymbolFlagsOptional != 0 {
		return nil
	}
	if origin != nil && l.literalReceiverGivesField(origin, field) {
		return nil
	}
	return l.phantomPresenceRefused(node, brand, []string{field}, "a read")
}

// phantomPresenceWhole refuses a road that observes every own member of value at once (3 to 5).
func (l *lowering) phantomPresenceWhole(node, value *ast.Node, road string) error {
	brand, members := l.phantomBrandIn(l.checker.GetTypeAtLocation(value))
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
		brand, _ := l.phantomBrandMatching(l.checker.GetTypeAtLocation(access.Expression), func(_ *checker.Type, members []string) bool { return slices.Contains(members, access.Name().Text()) })
		if brand == nil {
			return nil
		}
		return l.phantomPresenceRefused(node, brand, []string{access.Name().Text()}, "a write")
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral && key.Kind != ast.KindNumericLiteral && l.checker.GetTypeAtLocation(key).Flags()&checker.TypeFlagsNumberLike != 0 {
			// An array's numeric write is an element's. An object brand may name
			// a numeric phantom member, so that exception cannot apply to it.
			brand, members := l.phantomBrandMatching(l.checker.GetTypeAtLocation(access.Expression), func(base *checker.Type, members []string) bool {
				return !l.phantomArrayBase(base, members) && slices.ContainsFunc(members, func(name string) bool { _, err := strconv.ParseFloat(name, 64); return err == nil })
			})
			if brand == nil {
				return nil
			}
			return l.phantomPresenceRefused(node, brand, members, "a write")
		}
		return l.phantomPresenceKeyed(node, access.Expression, key, "a write")
	}
	return nil
}

// phantomPresenceRefused is the refusal, in the words Diagnostics gave (#gb0fqaa).
func (l *lowering) phantomPresenceRefused(node *ast.Node, brand *checker.Type, members []string, road string) error {
	fix := "a phantom brand is never present at run time, so test the value's real fields, or test it as its base type (rulings.md:31)"
	if array, _ := l.phantomBrandMatching(brand, l.phantomArrayBase); array != nil {
		fix = "a phantom brand is never present at run time, so test or copy the value as its base array (`readonly T[]`), or test the other arm's own field (rulings.md:31)"
	}
	return &Refused{Where: l.program.Where(node), What: "an observation of " + l.checker.TypeToString(brand) + "'s phantom member " + strings.Join(members, ", ") + " (" + road + ")", Fix: fix}
}

// arrayOwnMember is whether a method call's member is declared by Array or ReadonlyArray itself, the base a
// phantom array brand is held as, by the checker's symbol: every root, every declaration, in the library.
// Object.prototype's members (hasOwnProperty, propertyIsEnumerable) aren't.
func (l *lowering) arrayOwnMember(callee *ast.Node) bool {
	symbol := l.memberSymbol(callee)
	if symbol == nil {
		return false
	}
	roots := l.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			owner := declaration.Parent
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) || owner == nil || owner.Kind != ast.KindInterfaceDeclaration || (owner.Name().Text() != "Array" && owner.Name().Text() != "ReadonlyArray") {
				return false
			}
		}
	}
	return true
}

// propertyKeyArgument is whether an argument could name a property: a string or symbol, or a type that
// might be one (any, unknown, a type parameter), judged on every union arm.
func (l *lowering) propertyKeyArgument(argument *ast.Node) bool {
	argumentType := l.checker.GetTypeAtLocation(argument)
	if argumentType.Flags()&checker.TypeFlagsUnion != 0 {
		return slices.ContainsFunc(argumentType.Types(), func(part *checker.Type) bool {
			return part.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsESSymbolLike|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsInstantiable) != 0
		})
	}
	return argumentType.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsESSymbolLike|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsInstantiable) != 0
}

// holdsArrayBrand is whether value's type is, or holds, a phantom array brand.
func (l *lowering) holdsArrayBrand(value *ast.Node) bool {
	brand, _ := l.phantomBrandMatching(l.checker.GetTypeAtLocation(value), l.phantomArrayBase)
	return brand != nil
}

// Object brands may have real methods taking keys. Only the library Object's
// key-taking prototype members observe presence; a same-named own method does not.
// Array brands retain the broader arrayOwnMember road above.
func (l *lowering) objectPrototypeKeyMember(callee *ast.Node) bool {
	if name := callee.Name().Text(); name != "hasOwnProperty" && name != "propertyIsEnumerable" {
		return false
	}

	symbol := l.memberSymbol(callee)
	if symbol == nil {
		return false
	}
	roots := l.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if (root.Name != "hasOwnProperty" && root.Name != "propertyIsEnumerable") || len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			owner := declaration.Parent
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) || owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name().Text() != "Object" {
				return false
			}
		}
	}
	return true
}

func (l *lowering) phantomArrayBase(base *checker.Type, _ []string) bool {
	return base != nil && l.arrayLikeType(base)
}
