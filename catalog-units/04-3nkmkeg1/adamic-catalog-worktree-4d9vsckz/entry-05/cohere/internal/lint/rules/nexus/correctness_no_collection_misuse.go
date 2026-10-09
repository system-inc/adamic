package nexus

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const (
	correctnessNoCollectionMisuseInOnArrayId      = "inOnArray"
	correctnessNoCollectionMisuseSizeComparisonId = "impossibleSizeComparison"
	correctnessNoCollectionMisuseBracketAccessId  = "bracketAccessOnCollection"
	correctnessNoCollectionMisuseObjectMethodId   = "objectMethodOnCollection"
)

// correctnessNoCollectionMisuseInOnArrayText is one of the rule's messages, whose wording lives in
// `policy/messages/correctness-no-collection-misuse.json`.
var correctnessNoCollectionMisuseInOnArrayText = policy.MessageOf("nexus/correctness-no-collection-misuse", correctnessNoCollectionMisuseInOnArrayId)

func correctnessNoCollectionMisuseInOnArrayMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCollectionMisuseInOnArrayId,
		Description: correctnessNoCollectionMisuseInOnArrayText.Render(nil),
	}
}

// correctnessNoCollectionMisuseSizeComparisonText is one of the rule's messages, whose wording lives in
// `policy/messages/correctness-no-collection-misuse.json`.
var correctnessNoCollectionMisuseSizeComparisonText = policy.MessageOf("nexus/correctness-no-collection-misuse", correctnessNoCollectionMisuseSizeComparisonId)

func correctnessNoCollectionMisuseSizeComparisonMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCollectionMisuseSizeComparisonId,
		Description: correctnessNoCollectionMisuseSizeComparisonText.Render(nil),
	}
}

// correctnessNoCollectionMisuseBracketAccessText is one of the rule's messages, whose wording lives in
// `policy/messages/correctness-no-collection-misuse.json`.
var correctnessNoCollectionMisuseBracketAccessText = policy.MessageOf("nexus/correctness-no-collection-misuse", correctnessNoCollectionMisuseBracketAccessId)

func correctnessNoCollectionMisuseBracketAccessMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCollectionMisuseBracketAccessId,
		Description: correctnessNoCollectionMisuseBracketAccessText.Render(nil),
	}
}

// correctnessNoCollectionMisuseObjectMethodText is one of the rule's messages, whose wording lives in
// `policy/messages/correctness-no-collection-misuse.json`.
var correctnessNoCollectionMisuseObjectMethodText = policy.MessageOf("nexus/correctness-no-collection-misuse", correctnessNoCollectionMisuseObjectMethodId)

func correctnessNoCollectionMisuseObjectMethodMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCollectionMisuseObjectMethodId,
		Description: correctnessNoCollectionMisuseObjectMethodText.Render(nil),
	}
}

// CorrectnessNoCollectionMisuse reports four ways of treating a built-in collection that cannot do
// what they read as doing, each decided from the default library's declarations of the types
// involved: `in` used for array membership, a length or size compared where the answer never
// changes, brackets on a Map or Set, and `Object.keys` and its siblings on a Map or Set.
//
//	invalid: const roles: Role[] = ...; if('Admin' in roles) {}
//	invalid: if(items.length < 0) return;
//	invalid: while(queue.size !== -1) {}
//	invalid: const cache = new Map<string, number>(); cache[key] = 1;
//	invalid: Object.keys(new Map([['a', 1]]));
//	valid:   if(roles.includes('Admin')) {}
//	valid:   if(0 in sparse) {}
//	valid:   if('length' in roles) {}
//	valid:   if(items.length === 0) return;
//	valid:   cache.get(key); cache[Symbol.iterator]; cache['size'];
//	valid:   Object.keys(record);
//
// # Where it came from
//
// The JavaScript catalog pass of the new-rules sweep (`#tevhg3f`, items 9 and 10, after sonarjs's
// `no-in-misuse`, `no-collection-size-mischeck` and unicorn's `no-collection-bracket-access`,
// `no-object-methods-with-collections`, `no-impossible-length-comparison`), built in task
// `#j03vwm6`. The probes counted zero of each in ahra and Structure, so this is a guard on a class
// that is cheap to write and silent when it happens: every one of these shapes runs without throwing.
//
// # What makes each one exact
//
//   - **`in` on an array.** The right side's type is an array or tuple in every member of its
//     union, and the left side's type is a string literal, or a union of them, where no literal is a
//     canonical array index (`'0'`, `'12'`) and none names a property the array type has
//     (`'length'`, `'map'`). `in` on an array then asks for a property nothing declares, so it is
//     false. A wide `string`, a number and a symbol on the left are left alone: `index in array` is
//     how a hole in a sparse array is found, and a wide string may hold an index.
//   - **A length or size comparison.** `.length` on an array, a tuple, a string or a typed array,
//     or `.size` on a Map or Set (ReadonlyMap, ReadonlySet included), every member of the receiver's
//     type being one (so `items?.length` on a list that may be `undefined` is out), compared with a
//     numeric literal, possibly negated, such that no value from 0 up can change the answer: `< 0`,
//     `<= -1`, `=== -1` and `== -1` are always false, `>= 0`, `> -1`, `!== -1` and `!= -1` always
//     true, either way round.
//   - **Brackets on a Map or Set.** The object's type is Map, ReadonlyMap, WeakMap, Set,
//     ReadonlySet or WeakSet from the default library in every member, and the key's type cannot
//     name one of the collection's members: a string literal that is not one, a number, a boolean,
//     a bigint, `null` or `undefined`. `map['size']` and `map[Symbol.iterator]` read real members
//     and stay silent; so does a wide `string` or `symbol` key, which might. Under `noImplicitAny`
//     TypeScript already rejects most of these; this catches them where it is off or suppressed.
//   - **`Object` methods on a Map or Set.** `Object.keys`, `Object.values`, `Object.entries` or
//     `Object.getOwnPropertyNames`, with `Object` and the method both resolving to the default
//     library, called on one of the six collection types. A collection holds its entries in
//     internal slots, so its own properties are empty.
//
// The collection types are matched by the symbol the default library declares, so a project's own
// class named `Map`, a subclass of Map (which may declare fields of its own), and an intersection
// with extra members are all left alone.
//
// # What it declines: `indexOf(...) > 0`
//
// The sweep listed it beside these (sonarjs `index-of-compare-to-positive-number`), and it is not
// exact. `> 0` means "found, and not at the start" as often as it means a membership test that
// misses index 0: scriptaculous's `fontSize.indexOf(fontSizeType) > 0` requires a number before the
// unit, and TypeScript's own `generatedFile.test.mts` asserts `calls.indexOf("build") > 0`, that
// the build ran and was not first. Nothing in the types tells the two intents apart, so it is not
// reported. The membership spellings (`!== -1`, `>= 0`) belong to `@typescript-eslint/prefer-includes`,
// which is on in ahra.
//
// # No fix
//
// Each repair changes what the code means (`.includes()` for `in`, `.get()` for brackets, which
// comparison was intended), so each is the author's call.
var CorrectnessNoCollectionMisuse = rule.Rule{
	Name: "nexus/correctness-no-collection-misuse",

	// Every check rests on the type of an operand: an array, a string, a Map or Set.
	NeedsTypeChecker: true,

	// Whether a collection type or `Object` is the default library's is read off the lib list.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	// The rule reads types and the default library's declarations, never an imported body.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind == ast.KindInKeyword {
					if correctnessNoCollectionMisuseIsInOnArray(ctx, binary) {
						ctx.ReportNode(node, correctnessNoCollectionMisuseInOnArrayMessage())
					}
					return
				}
				if correctnessNoCollectionMisuseIsImpossibleSizeComparison(ctx, binary) {
					ctx.ReportNode(node, correctnessNoCollectionMisuseSizeComparisonMessage())
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				access := node.AsElementAccessExpression()
				if access.ArgumentExpression == nil ||
					!correctnessNoCollectionMisuseIsCollection(ctx, access.Expression, correctnessNoCollectionMisuseAnyCollection) {
					return
				}
				if correctnessNoCollectionMisuseKeyMissesEveryMember(ctx, access.Expression, access.ArgumentExpression) {
					ctx.ReportNode(node, correctnessNoCollectionMisuseBracketAccessMessage())
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if correctnessNoCollectionMisuseIsObjectMethodOnCollection(ctx, node) {
					ctx.ReportNode(node, correctnessNoCollectionMisuseObjectMethodMessage())
				}
			},
		}
	},
}

// correctnessNoCollectionMisuseAnyCollection are the default library's keyed collections.
var correctnessNoCollectionMisuseAnyCollection = map[string]bool{
	"Map": true, "ReadonlyMap": true, "Set": true, "ReadonlySet": true, "WeakMap": true, "WeakSet": true,
}

// correctnessNoCollectionMisuseSizedCollection are the collections with a `size`.
var correctnessNoCollectionMisuseSizedCollection = map[string]bool{
	"Map": true, "ReadonlyMap": true, "Set": true, "ReadonlySet": true,
}

// correctnessNoCollectionMisuseTypedArrays are the default library's typed arrays, whose `length` is
// an element count and never negative.
var correctnessNoCollectionMisuseTypedArrays = map[string]bool{
	"Int8Array": true, "Uint8Array": true, "Uint8ClampedArray": true, "Int16Array": true, "Uint16Array": true,
	"Int32Array": true, "Uint32Array": true, "Float16Array": true, "Float32Array": true, "Float64Array": true,
	"BigInt64Array": true, "BigUint64Array": true,
}

// correctnessNoCollectionMisuseObjectMethods are the `Object` methods that list own properties.
var correctnessNoCollectionMisuseObjectMethods = map[string]bool{
	"keys": true, "values": true, "entries": true, "getOwnPropertyNames": true,
}

// correctnessNoCollectionMisuseParts is the members of an expression's type, through a type
// parameter's constraint. An empty answer means nothing can be said.
func correctnessNoCollectionMisuseParts(ctx rule.Context, expression *ast.Node) []*checker.Type {
	expressionType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
	if expressionType == nil {
		return nil
	}
	return type_checking.UnionTypeParts(expressionType)
}

// correctnessNoCollectionMisuseIsLibraryType says whether a type's own symbol is one of the named
// default-library types, never a subclass or a project type of the same name.
func correctnessNoCollectionMisuseIsLibraryType(ctx rule.Context, part *checker.Type, names map[string]bool) bool {
	symbol := checker.Type_symbol(part)
	if symbol == nil {
		return false
	}
	return names[symbol.Name] && type_checking.IsSymbolFromDefaultLibrary(ctx.Program, symbol)
}

// correctnessNoCollectionMisuseIsCollection says whether every member of an expression's type is one
// of the named default-library collections.
func correctnessNoCollectionMisuseIsCollection(ctx rule.Context, expression *ast.Node, names map[string]bool) bool {
	parts := correctnessNoCollectionMisuseParts(ctx, expression)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !correctnessNoCollectionMisuseIsLibraryType(ctx, part, names) {
			return false
		}
	}
	return true
}

// correctnessNoCollectionMisuseIsInOnArray says whether `left in right` asks an array or tuple for
// a property no member of its type declares and no index can be.
func correctnessNoCollectionMisuseIsInOnArray(ctx rule.Context, binary *ast.BinaryExpression) bool {
	arrays := correctnessNoCollectionMisuseParts(ctx, binary.Right)
	if len(arrays) == 0 {
		return false
	}
	for _, part := range arrays {
		if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, part) {
			return false
		}
	}
	names := correctnessNoCollectionMisuseParts(ctx, binary.Left)
	if len(names) == 0 {
		return false
	}
	for _, part := range names {
		if !type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) {
			return false
		}
		name, isString := part.AsLiteralType().Value().(string)
		if !isString || correctnessNoCollectionMisuseIsArrayIndex(name) {
			return false
		}
		for _, array := range arrays {
			if checker.Checker_getPropertyOfType(ctx.TypeChecker, array, name) != nil {
				return false
			}
		}
	}
	return true
}

// correctnessNoCollectionMisuseIsArrayIndex says whether a property name is a canonical array index:
// the decimal spelling of an integer from 0 to 2^32 - 2, with no sign and no leading zero.
func correctnessNoCollectionMisuseIsArrayIndex(name string) bool {
	value, err := strconv.ParseUint(name, 10, 32)
	return err == nil && value < 1<<32-1 && strconv.FormatUint(value, 10) == name
}

// correctnessNoCollectionMisuseIsImpossibleSizeComparison says whether a comparison sets a length
// or size against a literal no value from 0 up can change the answer of.
func correctnessNoCollectionMisuseIsImpossibleSizeComparison(ctx rule.Context, binary *ast.BinaryExpression) bool {
	operator := binary.OperatorToken.Kind
	switch {
	case correctnessNoCollectionMisuseIsSize(ctx, binary.Left):
		value, isLiteral := correctnessNoCollectionMisuseNumericLiteral(binary.Right)
		return isLiteral && correctnessNoCollectionMisuseIsDecided(operator, value)
	case correctnessNoCollectionMisuseIsSize(ctx, binary.Right):
		value, isLiteral := correctnessNoCollectionMisuseNumericLiteral(binary.Left)
		return isLiteral && correctnessNoCollectionMisuseIsDecided(correctnessNoCollectionMisuseMirror(operator), value)
	}
	return false
}

// correctnessNoCollectionMisuseMirror is the operator that says the same thing with its operands
// swapped: `0 > size` is `size < 0`.
func correctnessNoCollectionMisuseMirror(operator ast.Kind) ast.Kind {
	switch operator {
	case ast.KindLessThanToken:
		return ast.KindGreaterThanToken
	case ast.KindGreaterThanToken:
		return ast.KindLessThanToken
	case ast.KindLessThanEqualsToken:
		return ast.KindGreaterThanEqualsToken
	case ast.KindGreaterThanEqualsToken:
		return ast.KindLessThanEqualsToken
	}
	return operator
}

// correctnessNoCollectionMisuseIsDecided says whether `size <operator> value` has one answer for every
// size from 0 up.
func correctnessNoCollectionMisuseIsDecided(operator ast.Kind, value float64) bool {
	switch operator {
	case ast.KindLessThanToken, ast.KindGreaterThanEqualsToken:
		return value <= 0
	case ast.KindLessThanEqualsToken, ast.KindGreaterThanToken,
		ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken,
		ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken:
		return value < 0
	}
	return false
}

// correctnessNoCollectionMisuseNumericLiteral reads a numeric literal, or one under a unary minus,
// through parentheses. A literal's text is the canonical spelling of its value.
func correctnessNoCollectionMisuseNumericLiteral(node *ast.Node) (float64, bool) {
	node = ast.SkipParentheses(node)
	sign := 1.0
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindMinusToken {
			return 0, false
		}
		sign = -1
		node = ast.SkipParentheses(unary.Operand)
	}
	if node.Kind != ast.KindNumericLiteral {
		return 0, false
	}
	value, err := strconv.ParseFloat(node.Text(), 64)
	if err != nil {
		return 0, false
	}
	return sign * value, true
}

// correctnessNoCollectionMisuseIsSize says whether an expression is `.length` of an array, tuple,
// string or typed array, or `.size` of a Map or Set, with every member of the receiver's type one of
// those, so a receiver that may be `undefined` (`items?.length`) is not.
func correctnessNoCollectionMisuseIsSize(ctx rule.Context, expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	if expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := expression.AsPropertyAccessExpression()
	if access.Name().Kind != ast.KindIdentifier {
		return false
	}
	receivers := correctnessNoCollectionMisuseParts(ctx, access.Expression)
	if len(receivers) == 0 {
		return false
	}
	switch access.Name().Text() {
	case "length":
		for _, part := range receivers {
			if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, part) &&
				!type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLike) &&
				!correctnessNoCollectionMisuseIsLibraryType(ctx, part, correctnessNoCollectionMisuseTypedArrays) {
				return false
			}
		}
		return true
	case "size":
		for _, part := range receivers {
			if !correctnessNoCollectionMisuseIsLibraryType(ctx, part, correctnessNoCollectionMisuseSizedCollection) {
				return false
			}
		}
		return true
	}
	return false
}

// correctnessNoCollectionMisuseKeyMissesEveryMember says whether every value a key's type allows
// becomes a property name that no member of the collection's type declares.
func correctnessNoCollectionMisuseKeyMissesEveryMember(ctx rule.Context, collection *ast.Node, key *ast.Node) bool {
	keyType := ctx.TypeChecker.GetTypeAtLocation(key)
	if keyType == nil {
		return false
	}
	collections := correctnessNoCollectionMisuseParts(ctx, collection)
	for part := range type_checking.UnionTypePartsSeq(keyType) {
		switch {
		case type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral):
			name, isString := part.AsLiteralType().Value().(string)
			if !isString {
				return false
			}
			for _, collectionPart := range collections {
				if checker.Checker_getPropertyOfType(ctx.TypeChecker, collectionPart, name) != nil {
					return false
				}
			}
		case type_checking.IsTypeFlagSet(part, checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|
			checker.TypeFlagsBigIntLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined):
		default:
			return false
		}
	}
	return true
}

// correctnessNoCollectionMisuseIsObjectMethodOnCollection says whether a call is the default
// library's `Object.keys`, `values`, `entries` or `getOwnPropertyNames` on a Map or Set.
func correctnessNoCollectionMisuseIsObjectMethodOnCollection(ctx rule.Context, node *ast.Node) bool {
	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	object := ast.SkipParentheses(access.Expression)
	if access.Name().Kind != ast.KindIdentifier || !correctnessNoCollectionMisuseObjectMethods[access.Name().Text()] ||
		object.Kind != ast.KindIdentifier || object.Text() != "Object" {
		return false
	}
	if !type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(object)) ||
		!type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(access.Name())) {
		return false
	}
	return correctnessNoCollectionMisuseIsCollection(ctx, call.Arguments.Nodes[0], correctnessNoCollectionMisuseAnyCollection)
}
