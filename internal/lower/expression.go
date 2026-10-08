package lower

import (
	"errors"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// typeOf is what's left at runtime of the type the checker proved for a node: a number, a boolean or
// a string. A union counts when every member is the same one ('Fizz' | 'Buzz' is a string).
func (l *lowering) typeOf(node *ast.Node) (ir.Type, error) {
	if l.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		if symbol := l.flagValueSymbol(ast.SkipParentheses(node)); symbol != nil {
			if stored, known := l.representation(l.checker.GetTypeOfSymbol(symbol)); known {
				return stored, nil
			}
		}
		return ir.Number, nil
	}
	if valueType, isKnown := l.representation(l.checker.GetTypeAtLocation(node)); isKnown {
		return valueType, nil
	}
	return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
}

func (l *lowering) representation(proven *checker.Type) (ir.Type, bool) {
	proven = l.concrete(proven)
	if kind := l.typedArrayKind(proven); kind != 0 {
		return kind, true
	}
	flags := proven.Flags()
	if flags&checker.TypeFlagsTypeParameter != 0 {
		// Inside a generic class, a type parameter is what this instantiation made it.
		substituted, isKnown := l.substitution[proven]
		return substituted, isKnown
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		// Target & WeakBrand is what a Weak<Target> reads as where it's present: the target.
		if target := l.weakTarget(proven); target != nil {
			return l.representation(target)
		}
		return l.objectIntersection(proven)
	}
	switch {
	case flags&checker.TypeFlagsNumberLike != 0:
		return ir.Number, true
	case flags&checker.TypeFlagsStringLike != 0:
		return ir.String, true
	case flags&checker.TypeFlagsBooleanLike != 0:
		return ir.Boolean, true
	case flags&checker.TypeFlagsObject != 0 && (l.checker.IsArrayType(proven) || l.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray")):
		return ir.Array, true
	case flags&checker.TypeFlagsObject != 0 && l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		// A Set is held as a Map whose values aren't used (set.go).
		return ir.Map, true
	case flags&checker.TypeFlagsObject != 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0:
		return ir.Object, true
	case flags&checker.TypeFlagsObject != 0:
		// An object with call signatures is a function, held as a closure.
		return ir.Closure, true
	case flags&checker.TypeFlagsUnion != 0:
		if l.includesNull(proven) {
			// A nullable match result uses NULL. A type also holding undefined needs a tag.
			if l.includesUndefined(proven) {
				return 0, false
			}
			for _, member := range proven.Types() {
				if member.Flags()&checker.TypeFlagsNull == 0 && !l.isLibraryType(member, "RegExpExecArray", "RegExpMatchArray") {
					return 0, false
				}
			}
		}
		var shared ir.Type
		mixed, weak := false, false
		for _, member := range proven.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				// undefined joins a union of references as a null pointer; it's checked below that
				// the rest are references.
				continue
			}
			if l.weakTarget(member) != nil {
				// Weak<Target> is (Target & WeakBrand) | undefined: kept weakly.
				weak = true
				continue
			}
			memberType, isKnown := l.representation(member)
			if !isKnown {
				return 0, false
			}
			if shared != 0 && memberType != shared {
				mixed = true
			}
			shared = memberType
		}
		if weak {
			return ir.Weak, shared == 0
		}
		if mixed {
			// Members held differently (string | number) are one Union, which holds undefined too.
			return ir.Union, true
		}
		if shared != 0 && !shared.IsReference() && l.includesUndefined(proven) {
			// number | undefined and boolean | undefined are each a present-and-value pair.
			if maybe := ir.Maybe(shared); maybe.IsMaybe() {
				return maybe, true
			}
			return 0, false
		}
		return shared, shared != 0
	}
	return 0, false
}

// isLibraryType reports whether a type is one of the library's, by name: Map, not a program's own
// interface that happens to be called Map.
func (l *lowering) isLibraryType(proven *checker.Type, names ...string) bool {
	symbol := proven.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return false
	}
	for _, name := range names {
		if symbol.Name == name {
			return true
		}
	}
	return false
}

func (l *lowering) includesUndefined(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return proven.Flags()&checker.TypeFlagsUndefined != 0
	}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			return true
		}
	}
	return false
}

func (l *lowering) includesNull(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return proven.Flags()&checker.TypeFlagsNull != 0
	}
	for _, member := range proven.Types() {
		if l.includesNull(member) {
			return true
		}
	}
	return false
}

// expression lowers a value. What's kept weakly (a Weak<Target> variable, field, element or map value)
// is read here as its target, so no value of a Weak type goes further; keeping one is fit's WeakOf.
func (l *lowering) expression(node *ast.Node) (ir.Expression, error) {
	if err := l.libraryIteratorUnsupportedUse(node); err != nil {
		return nil, err
	}
	if err := l.regexUnsupportedUse(node); err != nil {
		return nil, err
	}
	value, err := l.value(node)
	if literal := ast.SkipParentheses(node).Kind; err == nil && value.Type().IsReference() && literal != ast.KindArrayLiteralExpression && literal != ast.KindObjectLiteralExpression {
		// The checker lets { v: Box } be seen as { v: Weak<Box> } and back, an array of Box as one of
		// Weak<Box>, and (x: Weak<Box>) => ... as (x: Box) => ...; but one keeps a handle where the
		// other keeps the target, so the same object, array or function can't be both. A tuple is held
		// as an object, so it can't be seen as an array either, here or anywhere inside. A literal is
		// made as the type it's written into, so it never differs.
		if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
			if own := l.checker.GetTypeAtLocation(node); !l.typedArraySetArgument(node) && !l.sameKeeping(own, contextual, map[[2]*checker.Type]bool{}) {
				if l.typedArrayKind(l.checker.GetNonNullableType(own)) != 0 {
					return nil, l.notYet(node, "a typed array seen through a structural view that loses its buffer representation")
				}
				return nil, l.notYet(node, "a "+l.checker.TypeToString(own)+" seen as a "+l.checker.TypeToString(contextual)+" (one keeps something weakly that the other keeps strongly)")
			} else if tuple, array := l.tupleSeenAsArray(own, contextual, map[[2]*checker.Type]bool{}); tuple != nil {
				return nil, l.notYet(node, "a "+l.checker.TypeToString(tuple)+" seen as a "+l.checker.TypeToString(array)+" (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]])")
			}
		}
	}
	// Anything the check above doesn't see (a tuple inside a literal, or a union around it) is still
	// not yet: wherever a tuple flows into an array slot, at any depth.
	if skipped := ast.SkipParentheses(node); err == nil && skipped.Kind != ast.KindSpreadElement {
		if contextual := l.checker.GetContextualType(skipped, checker.ContextFlagsNone); contextual != nil && l.tupleWhereArrayGoes(l.checker.GetTypeAtLocation(skipped), contextual, 0) {
			return nil, l.notYet(skipped, "a tuple where an array goes (as "+l.checker.TypeToString(contextual)+")")
		}
	}
	if err != nil || value.Type() != ir.Weak {
		return value, err
	}
	read := l.checker.GetTypeAtLocation(node)
	to, present := ir.Object, false
	if target, isKnown := l.representation(read); isKnown && target != ir.Weak {
		// The checker narrowed it to present (Target & WeakBrand).
		to, present = target, true
	} else if read.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range read.Types() {
			if target := l.weakTarget(member); target != nil {
				if to, isKnown = l.representation(target); !isKnown {
					return nil, l.notYet(node, "a Weak of "+l.checker.TypeToString(target))
				}
			}
		}
	}
	if !to.IsReference() {
		return nil, l.notYet(node, "a Weak of "+l.checker.TypeToString(read))
	}
	return ir.WeakTarget{Value: value, To: to, Present: present}, nil
}

// sameKeeping reports whether a value of type from, seen as type to, keeps everything inside it the
// same way through both: every field to has, every element, map key and value, and every parameter
// and result of a function, Weak in both or in neither, all the way down.
func (l *lowering) sameKeeping(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	if fromKind, toKind := l.typedArrayKind(from), l.typedArrayKind(to); fromKind != 0 || toKind != 0 {
		return fromKind != 0 && fromKind == toKind
	}
	same := func(inside, viewed *checker.Type) bool {
		fromKept, _ := l.kept(inside)
		toKept, _ := l.kept(viewed)
		return (fromKept == ir.Weak) == (toKept == ir.Weak) && l.sameKeeping(inside, viewed, visited)
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		if !l.sameRestConvention(fromSignatures[0], toSignatures[0]) {
			return false
		}
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		if l.censusNeverRestSignature(toSignatures[0]) {
			toParameters = nil
		}
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if !same(l.checker.GetTypeOfSymbol(fromParameters[index]), l.checker.GetTypeOfSymbol(toParameters[index])) {
				return false
			}
		}
		return same(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]))
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := l.typeArguments(from), l.typeArguments(to)
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if !same(fromArguments[index], toArguments[index]) {
				return false
			}
		}
	default:
		for _, viewed := range l.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil && !same(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)) {
				return false
			}
		}
	}
	return true
}

// tupleSeenAsArray finds, in a value of type from seen as type to, the first tuple seen as an array:
// at the top, or as an element, field, map key or value, parameter or result, all the way down as
// sameKeeping walks. It returns the tuple's type and the array's, or nils.
func (l *lowering) tupleSeenAsArray(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) (*checker.Type, *checker.Type) {
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return nil, nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if checker.IsTupleType(from) && !checker.IsTupleType(to) {
		return from, to
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		// A function is handed the other's arguments, and its results are seen as the other's.
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		if l.censusNeverRestSignature(toSignatures[0]) {
			toParameters = nil
		}
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if tuple, array := l.tupleSeenAsArray(l.checker.GetTypeOfSymbol(toParameters[index]), l.checker.GetTypeOfSymbol(fromParameters[index]), visited); tuple != nil {
				return tuple, array
			}
		}
		return l.tupleSeenAsArray(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), visited)
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := l.typeArguments(from), l.typeArguments(to)
		if checker.IsTupleType(from) && l.checker.IsArrayType(to) {
			return from, to
		}
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if tuple, array := l.tupleSeenAsArray(fromArguments[index], toArguments[index], visited); tuple != nil {
				return tuple, array
			}
		}
	default:
		for _, viewed := range l.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil {
				if tuple, array := l.tupleSeenAsArray(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed), visited); tuple != nil {
					return tuple, array
				}
			}
		}
	}
	return nil, nil
}

// present is a type without undefined, and a Weak's narrowing without its brand: the one object,
// array, map or function type it is, or nil when it's none or several.
func (l *lowering) present(proven *checker.Type) *checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		var only *checker.Type
		for _, member := range proven.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
			if only != nil {
				return nil
			}
			only = member
		}
		if only == nil {
			return nil
		}
		proven = only
	}
	if target := l.weakTarget(proven); target != nil {
		proven = target
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	return proven
}

// kept is how an element or a map's value of a type is kept: as its representation, except that one
// narrowed from a Weak (Node & WeakBrand, as filter(x => x !== undefined) or every leaves an array of
// Weak<Node>) is still a Weak, a handle, since that's what the array holds.
func (l *lowering) kept(proven *checker.Type) (ir.Type, bool) {
	if l.weakTarget(proven) != nil {
		return ir.Weak, true
	}
	return l.representation(proven)
}

// typeArguments is a generic type's arguments, Map<K, V>'s K and V, through a Weak's narrowing
// (Map<K, V> & WeakBrand), or nil for a type that has none.
func (l *lowering) typeArguments(proven *checker.Type) []*checker.Type {
	if target := l.weakTarget(proven); target != nil {
		proven = target
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return nil
	}
	return l.checker.GetTypeArguments(proven)
}

// weakTarget is the Target of Target & WeakBrand, the present half of a Weak<Target>, or nil for
// any other type.
func (l *lowering) weakTarget(proven *checker.Type) *checker.Type {
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return nil
	}
	var target *checker.Type
	branded := false
	for _, member := range proven.Types() {
		if symbol := member.Symbol(); symbol != nil && symbol.Name == "WeakBrand" && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			branded = true
		} else if target == nil {
			target = member
		} else {
			return nil
		}
	}
	if !branded {
		return nil
	}
	return target
}

// value lowers a value, as expression does, but leaves a Weak as it's kept.
func (l *lowering) value(node *ast.Node) (ir.Expression, error) {
	if identity := l.enumNeverIdentity(node, map[*ast.Node]bool{}); identity != nil {
		value, err := l.enumNeverValue(node)
		if err != nil {
			return nil, err
		}
		return l.enumNeverCheck(node, value, identity), nil
	}
	return l.enumNeverValue(node)
}

func (l *lowering) enumNeverValue(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if value, handled, err := l.typedArrayExpression(node); handled {
		return value, err
	}
	if value, known, err := l.namespaceExpression(node); known {
		return value, err
	}
	if value, known, err := l.libraryMethodValue(node); known {
		return value, err
	}
	if value, known, err := l.enumExpression(node); known {
		return value, err
	}
	if observed, known := l.libraryArrayObservation(node); known {
		return observed, nil
	}
	switch node.Kind {
	case ast.KindNullKeyword:
		return ir.Null{}, nil
	case ast.KindRegularExpressionLiteral:
		return l.regexConstant(node)
	case ast.KindNumericLiteral:
		return l.numericLiteral(node)
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return ir.StringConstant{Index: l.constant(node.Text())}, nil
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return ir.BooleanConstant{Value: node.Kind == ast.KindTrueKeyword}, nil
	case ast.KindIdentifier:
		if value, known, err := l.libraryGlobalValue(node); known {
			return value, err
		}
		if value, handled := l.staticClassRead(node); handled {
			return value, nil
		}
		local, isLocal := l.local(node)
		if !isLocal && node.Text() == "undefined" {
			return ir.Undefined{}, nil
		}
		if !isLocal && (l.isLibraryGlobal(node, "NaN") || l.isLibraryGlobal(node, "Infinity")) {
			return ir.NumberConstant{Value: numberConstants[node.Text()]}, nil
		}
		if function, isFunction := l.functions[l.symbol(node)]; !isLocal && isFunction {
			return l.functionValue(node, function)
		}
		if _, isGeneric := l.generics[l.symbol(node)]; !isLocal && isGeneric {
			return nil, l.notYet(node, "a generic function as a value")
		}
		if !isLocal && l.isLibraryGlobal(node, "String") {
			return nil, l.notYet(node, "reading String as a first-class constructor (its any-typed call signature, construction and static members need an intrinsic value representation)")
		}
		if !isLocal {
			return nil, l.notYet(node, "reading "+node.Text())
		}
		read := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checkedModuleRead(node, local)})
		if l.result.Locals[local].Type == ir.Union {
			// Where the checker has narrowed it to fewer members held one way, it's read as that.
			parent := node.Parent
			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				parent = parent.Parent
			}
			observing := comparedWithUndefined(node) || (parent != nil && parent.Kind == ast.KindTypeOfExpression)
			if narrowed, isKnown := l.representation(l.checker.GetTypeAtLocation(node)); isKnown && narrowed != ir.Union && !observing {
				// Calls and captured writes can invalidate the checker's narrowing. Check the
				// held member before casting it, with ordinary IR shared by both backends.
				name := "object"
				switch narrowed.Present() {
				case ir.Number:
					name = "number"
				case ir.Boolean:
					name = "boolean"
				case ir.String:
					name = "string"
				case ir.Closure:
					name = "function"
				}
				if name == "object" {
					// typeof cannot distinguish differently held object members.
					declared := l.concrete(l.checker.GetTypeOfSymbol(l.symbol(node)))
					members := []*checker.Type{declared}
					if declared.Flags()&checker.TypeFlagsUnion != 0 {
						members = declared.Types()
					}
					for _, member := range members {
						if held, known := l.representation(member); known && held != narrowed && (held == ir.Object || held == ir.Array || held == ir.Map) {
							return nil, l.notYet(node, "a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables")
						}
					}
				}
				b := l.libraryArrayBuilder([]ir.Expression{read})
				held := b.read(b.parameters[0])
				matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant(name)}})
				if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
					matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
				}
				message := "union member where the checker narrowed it away: a call since the narrowing put it back"
				b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
				read = b.finish("narrowed_union_member", ir.Narrow{Value: held, To: narrowed})
			}
		}
		if declared := l.result.Locals[local].Type; declared.IsMaybe() {
			// Where the checker has narrowed it to what it holds, it's read as that.
			if narrowed, _ := l.representation(l.checker.GetTypeAtLocation(node)); narrowed == declared.Present() && !l.acceptsUndefined(node) {
				read = ir.Unwrap{Value: read}
			}
		}
		return l.defined(node, read), nil
	case ast.KindPrefixUnaryExpression:
		return l.prefix(node)
	case ast.KindTypeOfExpression:
		if l.isLibraryGlobal(node.AsTypeOfExpression().Expression, "Number") {
			return ir.StringConstant{Index: l.constant("function")}, nil
		}
		if value, intrinsic := l.stringTypeOf(node); intrinsic {
			return value, nil
		}
		operand, err := l.expression(node.AsTypeOfExpression().Expression)
		if err != nil {
			return nil, err
		}
		written := node.AsTypeOfExpression().Expression
		null := l.typeOfNull(written)
		if null && l.includesUndefined(l.concrete(l.checker.GetTypeAtLocation(written))) {
			switch operand.(type) {
			case ir.ArrayIndex, ir.MapGet, ir.ArrayPop:
				// The lookup still has a presence slot, so typeof can distinguish null from undefined.
			default:
				return nil, l.notYet(node, "typeof a value holding both null and undefined without a presence slot")
			}
		}
		return ir.TypeOf{Value: operand, Null: null}, nil
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			return l.coalesce(node)
		}
		if binary.OperatorToken.Kind == ast.KindInstanceOfKeyword {
			if lowered, isCaught := l.caughtInstanceOfError(node); isCaught {
				return lowered, nil
			}
			return l.classInstanceOf(node)
		}
		left, err := l.expression(binary.Left)
		if err != nil {
			return nil, err
		}
		right, err := l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		if binary.OperatorToken.Kind == ast.KindPlusToken {
			left, right = l.spelled(binary.Left, left), l.spelled(binary.Right, right)
		}
		return l.combine(node, binary.OperatorToken.Kind, left, right)
	case ast.KindTaggedTemplateExpression:
		return l.stringRawTemplate(node)
	case ast.KindTemplateExpression:
		return l.template(node)
	case ast.KindConditionalExpression:
		return l.conditional(node)
	case ast.KindObjectLiteralExpression:
		return l.objectLiteral(node)
	case ast.KindArrayLiteralExpression:
		return l.arrayLiteral(node)
	case ast.KindPropertyAccessExpression:
		return l.property(node)
	case ast.KindElementAccessExpression:
		return l.elementAccess(node)
	case ast.KindNewExpression:
		return l.newExpression(node)
	case ast.KindThisKeyword:
		if l.this < 0 {
			return nil, l.notYet(node, "this outside a method")
		}
		if err := l.useOfThis(node); err != nil {
			return nil, err
		}
		l.touch(l.this)
		return ir.Read{Local: l.this, Of: ir.Object}, nil
	case ast.KindArrowFunction:
		return l.closure(node)
	case ast.KindAsExpression:
		return l.cast(node)
	case ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		if err := l.provenRelation(node, satisfies.Expression, l.checker.GetTypeAtLocation(satisfies.Type)); err != nil {
			return nil, err
		}
		return l.expression(satisfies.Expression)
	case ast.KindFunctionExpression:
		return l.functionExpression(node)
	case ast.KindCallExpression:
		if err := l.optionalCall(node); err != nil {
			return nil, err
		}
		if lowered, isBuiltin, err := l.builtin(node); isBuiltin {
			if err == nil && lowered.Type() == 0 {
				// forEach is void; as a value it's undefined, which only places 0.1 refuses would use.
				return nil, l.notYet(node, "a void call used as a value")
			}
			return lowered, err
		}
		call, err := l.callOrMethod(node)
		if err != nil {
			return nil, err
		}
		if call.Type() == 0 {
			// The checker allows a void call where a value goes only in places 0.1 refuses anyway
			// (a template of void prints "undefined"); stage 0 says so rather than guess.
			return nil, l.notYet(node, "a void call used as a value")
		}
		return call, nil
	}
	return nil, l.notYet(node, describe(node))
}

// fit makes a value fit where a value of type to goes: a number or a boolean, or undefined, where
// number | undefined or boolean | undefined goes, since that is two words and they are one. It also
// unwraps a maybe value the checker narrowed to its present type. Anything else is left as it is.
func fit(value ir.Expression, to ir.Type) ir.Expression {
	if to == ir.Weak && value != nil && value.Type() != ir.Weak {
		return ir.WeakOf{Value: value}
	}
	if to == ir.Union && value != nil && value.Type() != ir.Union {
		return ir.Box{Value: value}
	}
	if _, isNull := value.(ir.Null); isNull && to.IsReference() && to != ir.Union {
		return ir.Null{Of: to}
	}
	if _, isUndefined := value.(ir.Undefined); isUndefined && to.IsReference() && to != ir.Union {
		// undefined going where a string, an array or a function may be missing is that reference,
		// missing: typed as it, so the C holding it is.
		return ir.Undefined{Of: to}
	}
	if !to.IsMaybe() || value == nil {
		if value != nil && value.Type().IsMaybe() && value.Type().Present() == to {
			return ir.Unwrap{Value: value}
		}
		return value
	}
	if value.Type() == to.Present() {
		return ir.MaybeOf{Value: value, Of: to}
	}
	if _, isUndefined := value.(ir.Undefined); isUndefined {
		return ir.MaybeOf{Of: to}
	}
	return value
}

// numericLiteral is the value the checker read from the literal, so 0x1F, 1_000 and 1e3 all mean
// what JavaScript says they mean without a second parser here.
func (l *lowering) numericLiteral(node *ast.Node) (ir.Expression, error) {
	literal := l.checker.GetTypeAtLocation(node)
	if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, errors.New("lower: " + l.program.Where(node) + ": the checker gave a numeric literal a type that isn't a number literal")
	}
	value := reflect.ValueOf(literal.AsLiteralType().Value())
	if value.Kind() != reflect.Float64 {
		return nil, errors.New("lower: " + l.program.Where(node) + ": a numeric literal's value isn't a float64")
	}
	return ir.NumberConstant{Value: value.Float()}, nil
}

func (l *lowering) prefix(node *ast.Node) (ir.Expression, error) {
	prefix := node.AsPrefixUnaryExpression()
	if prefix.Operator == ast.KindPlusToken || prefix.Operator == ast.KindMinusToken || prefix.Operator == ast.KindTildeToken {
		operand, err := l.libraryNumber(prefix.Operand)
		if err != nil {
			return nil, err
		}
		operator := ir.Plus
		if prefix.Operator == ast.KindMinusToken {
			operator = ir.Negate
		}
		if prefix.Operator == ast.KindTildeToken {
			operator = ir.BitNot
		}
		return ir.Unary{Operator: operator, Operand: operand}, nil
	}
	operand, err := l.expression(prefix.Operand)
	if err != nil {
		return nil, err
	}
	if prefix.Operator == ast.KindExclamationToken {
		return ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil
	}
	return nil, l.notYet(node, describe(node)+" on a "+typeName(operand.Type()))
}

var arithmetic = map[ast.Kind]ir.Operator{
	ast.KindPlusToken:             ir.Add,
	ast.KindMinusToken:            ir.Subtract,
	ast.KindAsteriskToken:         ir.Multiply,
	ast.KindSlashToken:            ir.Divide,
	ast.KindPercentToken:          ir.Remainder,
	ast.KindAsteriskAsteriskToken: ir.Power,
}

// bitwise are the binary bitwise operators, on two numbers.
var bitwise = map[ast.Kind]ir.Operator{
	ast.KindAmpersandToken:                         ir.BitAnd,
	ast.KindBarToken:                               ir.BitOr,
	ast.KindCaretToken:                             ir.BitXor,
	ast.KindLessThanLessThanToken:                  ir.ShiftLeft,
	ast.KindGreaterThanGreaterThanToken:            ir.ShiftRight,
	ast.KindGreaterThanGreaterThanGreaterThanToken: ir.ShiftRightUnsigned,
}

var comparisons = map[ast.Kind]ir.Operator{
	ast.KindLessThanToken:          ir.Less,
	ast.KindLessThanEqualsToken:    ir.LessOrEqual,
	ast.KindGreaterThanToken:       ir.Greater,
	ast.KindGreaterThanEqualsToken: ir.GreaterOrEqual,
}

// combine lowers a binary operator on two lowered operands.
func (l *lowering) combine(node *ast.Node, operator ast.Kind, left ir.Expression, right ir.Expression) (ir.Expression, error) {
	both := func(want ir.Type) bool { return left.Type() == want && right.Type() == want }
	if operator == ast.KindPlusToken && both(ir.String) {
		return ir.Concat{Parts: []ir.Expression{left, right}}, nil
	}
	if lowered, isArithmetic := arithmetic[operator]; isArithmetic && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if lowered, isBitwise := bitwise[operator]; isBitwise && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if lowered, isComparison := comparisons[operator]; isComparison && (both(ir.Number) || both(ir.String)) {
		// Strings compare in UTF-16 code unit order, as JavaScript's do.
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		_, leftNull := left.(ir.Null)
		_, rightNull := right.(ir.Null)
		if leftNull != rightNull {
			value := left
			if leftNull {
				value = right
			}
			if !value.Type().IsReference() {
				return nil, l.notYet(node, "null comparison with a scalar")
			}
			operand := node.AsBinaryExpression().Left
			if leftNull {
				operand = node.AsBinaryExpression().Right
			}
			test := ir.Expression(ir.IsNull{Value: value, AlwaysFalse: !l.includesNull(l.checker.GetTypeAtLocation(operand))})
			if operator == ast.KindExclamationEqualsEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return test, nil
		}
		// x === undefined tests for a missing reference, whatever x's type.
		_, leftUndefined := left.(ir.Undefined)
		_, rightUndefined := right.(ir.Undefined)
		if leftUndefined != rightUndefined {
			value := left
			if leftUndefined {
				value = right
			}
			if pair := ir.Maybe(value.Type()); pair.IsMaybe() {
				// A number or a boolean the checker narrowed to present: still evaluated, and never
				// undefined.
				value = fit(value, pair)
			}
			if !value.Type().IsReference() && !value.Type().IsMaybe() {
				return nil, l.notYet(node, "comparing a "+typeName(value.Type())+" with undefined")
			}
			test := ir.Expression(ir.IsUndefined{Value: value})
			operand := node.AsBinaryExpression().Left
			if leftUndefined {
				operand = node.AsBinaryExpression().Right
			}
			if l.includesNull(l.checker.GetTypeAtLocation(operand)) {
				test = ir.IsNull{Value: value, AlwaysFalse: true}
			}
			if operator == ast.KindExclamationEqualsEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return test, nil
		}
	}
	if (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) && left.Type() != right.Type() {
		// A pair against what it holds, or against another pair: each made a pair, and compared as one.
		if pair := left.Type(); pair.IsMaybe() && right.Type() == pair.Present() {
			right = fit(right, pair)
		} else if pair := right.Type(); pair.IsMaybe() && left.Type() == pair.Present() {
			left = fit(left, pair)
		} else if left.Type() == ir.Union || right.Type() == ir.Union {
			// A union against anything: both as unions, compared by member and value.
			left, right = fit(left, ir.Union), fit(right, ir.Union)
		}
	}
	if (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) && left.Type() == right.Type() {
		lowered := ir.Equal
		if operator == ast.KindExclamationEqualsEqualsToken {
			lowered = ir.NotEqual
		}
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if operator == ast.KindBarBarToken && left.Type() == ir.Closure && right.Type() == ir.Closure {
		// A present function is always truthy; nullable closures use undefined's
		// null pointer. Coalesce preserves selection and evaluates each side once.
		return ir.Coalesce{Value: left, Fallback: right, Of: ir.Closure}, nil
	}
	if value, known := l.censusBooleanLogical(node, operator, left, right); known {
		return value, nil
	}

	return nil, l.notYet(node, describe(node)+" with a "+typeName(left.Type())+" and a "+typeName(right.Type()))
}

// spelled is a string as + and a template write it: one that may be missing (a null reference) is
// written "undefined", as JavaScript writes it.
func (l *lowering) spelled(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() != ir.String || !(l.includesUndefined(l.checker.GetTypeAtLocation(node)) || l.narrowedAway(ast.SkipParentheses(node))) {
		return value
	}
	return ir.Coalesce{Value: value, Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}
}

// template lowers a template literal to a Concat, writing each value as String() would.
func (l *lowering) template(node *ast.Node) (ir.Expression, error) {
	template := node.AsTemplateExpression()
	parts := []ir.Expression{}
	if head := template.Head.Text(); head != "" {
		parts = append(parts, ir.StringConstant{Index: l.constant(head)})
	}
	for _, span := range template.TemplateSpans.Nodes {
		value, err := l.expression(span.AsTemplateSpan().Expression)
		if err != nil {
			return nil, err
		}
		switch value.Type() {
		case ir.Number:
			value = ir.NumberToString{Value: value}
		case ir.Boolean:
			value = ir.BooleanToString{Value: value}
		case ir.MaybeNumber, ir.MaybeBoolean:
			value = ir.MaybeToString{Value: value}
		case ir.Union:
			if !l.writable(l.checker.GetTypeAtLocation(span.AsTemplateSpan().Expression)) {
				return nil, l.notYet(span, "a template interpolating a union with an object, an array, a map or a function in it")
			}
			value = ir.UnionToString{Value: value}
		case ir.String:
			value = l.spelled(span.AsTemplateSpan().Expression, value)
		default:
			// JavaScript writes an object as "[object Object]", an array as its join, and a function as
			// its source; 0.1 has no use for any of it.
			return nil, l.notYet(span, "a template interpolating an object, an array, a map, a function or undefined")
		}
		parts = append(parts, value)
		if literal := span.AsTemplateSpan().Literal.Text(); literal != "" {
			parts = append(parts, ir.StringConstant{Index: l.constant(literal)})
		}
	}
	return ir.Concat{Parts: parts}, nil
}

func (l *lowering) conditional(node *ast.Node) (ir.Expression, error) {
	conditional := node.AsConditionalExpression()
	condition, err := l.condition(conditional.Condition)
	if err != nil {
		return nil, err
	}
	whenTrue, err := l.expression(conditional.WhenTrue)
	if err != nil {
		return nil, err
	}
	whenNot, err := l.expression(conditional.WhenFalse)
	if err != nil {
		return nil, err
	}
	if whenTrue.Type() != whenNot.Type() && l.acceptsUndefined(node) {
		// A stale narrowing in either branch may still hold undefined. Keep that representation
		// when the whole conditional is observed or written into a slot that accepts it.
		if pair := whenTrue.Type(); pair.IsMaybe() && whenNot.Type() == pair.Present() {
			whenNot = fit(whenNot, pair)
		} else if pair := whenNot.Type(); pair.IsMaybe() && whenTrue.Type() == pair.Present() {
			whenTrue = fit(whenTrue, pair)
		}
	}
	if whenTrue.Type() != whenNot.Type() {
		// flag ? 1 : undefined is number | undefined, and flag ? 1 : 'one' a union: each branch made
		// one.
		if of, err := l.typeOf(node); err == nil && (of.IsMaybe() || of == ir.Union || of.IsReference()) {
			whenTrue, whenNot = fit(whenTrue, of), fit(whenNot, of)
		}
	}
	if whenTrue.Type() != whenNot.Type() {
		// flag ? text : undefined is a reference that may be missing, held as text is, with undefined
		// the null one: the checker's type for the whole says which.
		_, trueUndefined := whenTrue.(ir.Undefined)
		_, notUndefined := whenNot.(ir.Undefined)
		present := whenTrue
		if trueUndefined {
			present = whenNot
		}
		if of, err := l.typeOf(node); err == nil && trueUndefined != notUndefined && present.Type().IsReference() && of == present.Type() {
			return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot, Of: of}, nil
		}
		return nil, l.notYet(node, "a conditional whose branches have different types")
	}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot}, nil
}

func typeName(valueType ir.Type) string {
	if valueType.IsTypedArray() {
		return typedArrayName(valueType)
	}
	switch valueType {
	case ir.Number:
		return "number"
	case ir.Boolean:
		return "boolean"
	case ir.String:
		return "string"
	case ir.MaybeNumber:
		return "number | undefined"
	case ir.MaybeBoolean:
		return "boolean | undefined"
	case ir.Union:
		return "union of differently held members"
	}
	return "value"
}

// writable reports whether every member of a union is one a template writes: a number, a boolean, a
// string or undefined.
func (l *lowering) writable(proven *checker.Type) bool {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		memberType, isKnown := l.representation(member)
		if !isKnown || (memberType != ir.Number && memberType != ir.Boolean && memberType != ir.String) {
			return false
		}
	}
	return true
}

// slotless reports whether a value of the type can't yet be held in one word: a field, an element, a
// map's value, a cell, or a function value's argument or result. number | undefined is packed into
// one (a reserved NaN is undefined); boolean | undefined has a tagged byte for object fields
// (censusFieldSlotless), but remains unsupported in the other slot contexts, and a
// Union has to be boxed on its way in, which stage 0 does only where a variable, a parameter or a
// result takes one.
func slotless(valueType ir.Type) bool {
	return valueType == ir.MaybeBoolean || valueType == ir.Union
}

// call lowers a call to one of the module's functions, or to a function value.
func (l *lowering) call(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	qualified := l.namespaceMember(callee)
	if declaration, isGeneric := l.generics[l.symbol(callee)]; (ast.IsIdentifier(callee) || qualified) && isGeneric {
		instance, err := l.instantiateFunction(node, declaration)
		if err != nil {
			return nil, err
		}
		return l.callFunction(call, instance)
	}
	function, isFunction := l.functions[l.symbol(callee)]
	if (!ast.IsIdentifier(callee) && !qualified) || !isFunction {
		if calleeType, _ := l.representation(l.checker.GetTypeAtLocation(callee)); calleeType == ir.Closure {
			return l.callClosure(node)
		}
		return nil, l.notYet(node, "a call to "+describe(callee))
	}
	return l.callFunction(call, function)
}

// callFunction lowers a call's arguments, in order, and the call to function.
func (l *lowering) callFunction(call *ast.CallExpression, function int) (ir.Expression, error) {
	if rest := l.censusRestDeclaration(function); rest != nil {
		return l.censusRestCall(call, function, rest)
	}

	arguments := []ir.Expression{}
	for _, argument := range call.Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, lowered)
	}
	return l.censusOverloadResult(call, ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns})
}

// coalesce lowers value ?? fallback, and value ?? panic('why'), evaluating the right side only when
// the left is missing.
func (l *lowering) coalesce(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	value, err := l.expression(binary.Left)
	if err != nil {
		return nil, err
	}
	// What ?? makes is the checker's: the left side's present members and the right side's, which
	// may be held differently (text ?? count is string | number, a Union).
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	if present := value.Type(); !present.IsMaybe() && !present.IsReference() {
		// A value that can't be missing: ?? never runs its right side.
		return fit(value, of), nil
	}
	right := ast.SkipParentheses(binary.Right)
	if right.Kind == ast.KindCallExpression && l.isPreludeFunction(right.AsCallExpression().Expression, "panic") && len(right.AsCallExpression().Arguments.Nodes) == 1 {
		message, err := l.expression(right.AsCallExpression().Arguments.Nodes[0])
		if err != nil {
			return nil, err
		}
		return ir.Coalesce{Value: value, Panic: message, Of: of}, nil
	}
	fallback, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	fallback = fit(fallback, of)
	if _, isUndefined := fallback.(ir.Undefined); fallback.Type() != of && !(isUndefined && of.IsReference()) {
		return nil, l.notYet(node, "?? whose sides have different types")
	}
	return ir.Coalesce{Value: value, Fallback: fallback, Of: of}, nil
}

// closure lowers an arrow function to a function of its own and the closure that captures it.
func (l *lowering) closure(node *ast.Node) (ir.Expression, error) {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "closure", Closure: true})
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	if l.instance != nil {
		l.instance.templates = append(l.instance.templates, template{closure: index, proven: l.checker.GetTypeAtLocation(node), isClosure: true})
	}
	l.closures = append(l.closures, index)
	err := l.lowerFunction(index, node, -1)
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}

// functionValue lowers a module function read as a value rather than called: a function value whose
// code forwards its arguments to the function, made once for each function read so, before the
// program runs. Its closure receives omitted arguments as undefined before forwarding them;
// the declared function retains its ordinary default-parameter prologue.
func (l *lowering) functionValue(node *ast.Node, target int) (ir.Expression, error) {
	symbol := l.symbol(node)
	if node.Kind == ast.KindShorthandPropertyAssignment {
		symbol = l.checker.GetShorthandAssignmentValueSymbol(node)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = l.checker.GetAliasedSymbol(symbol)
		}
		symbol = l.checker.GetExportSymbolOfSymbol(symbol)
	}
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && l.censusImplementation(declaration) != nil {
				return nil, l.notYet(node, "an overloaded function as a value")
			}
		}
	}
	if held, isMade := l.forwarders[target]; isMade {
		return ir.Read{Local: held, Of: ir.Closure}, nil
	}
	for _, parameter := range symbol.Declarations[0].Parameters() {
		if parameter.Name().Text() == "this" {
			return nil, l.notYet(parameter, "a function with a this parameter used as a value; pass the receiver explicitly or use an arrow")
		}
		declared := parameter.AsParameterDeclaration()
		if declared.DotDotDotToken != nil {
			return nil, l.notYet(node, "a function with a rest parameter, as a value")
		}
	}
	callee := l.result.Functions[target]
	if censusCallableSlotless(callee.Returns) {
		return nil, l.notYet(node, "a function value returning "+typeName(callee.Returns))
	}
	index := len(l.result.Functions)
	forwarder := ir.Function{Name: callee.Name + "_value", Closure: true, Returns: callee.Returns}
	arguments := []ir.Expression{}
	for _, parameter := range callee.Parameters {
		declared := l.result.Locals[parameter]
		if censusCallableSlotless(declared.Type) {
			return nil, l.notYet(node, "a function value taking "+typeName(declared.Type))
		}
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: declared.Name, Type: declared.Type, Function: index})
		forwarder.Parameters = append(forwarder.Parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: declared.Type})
	}
	call := ir.Call{Function: target, Arguments: arguments, Returns: callee.Returns}
	if callee.Returns == 0 {
		forwarder.Body = []ir.Statement{ir.Evaluate{Value: call}}
	} else {
		forwarder.Body = []ir.Statement{ir.Return{Value: call}}
	}
	l.result.Functions = append(l.result.Functions, forwarder)
	// One function value for the function, made before anything runs and held by a global of its
	// own, so reading the function twice gives the same value, === as JavaScript's.
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: callee.Name + "_value", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[target] = held
	return ir.Read{Local: held, Of: ir.Closure}, nil
}

// optionalCall refuses a call in an optional chain, text?.toUpperCase() or run?.(): the receiver has
// to be evaluated once and tested before the call, which stage 0 doesn't lower yet for the library's
// methods or a function value called directly. Lowered as a plain call, it ran the method on
// undefined. The one form lowered is object?.method(...) on a method the program declares (a
// class's, an interface's, or a function value in a field): the call goes through the object's
// methods, which stops at undefined as JavaScript's chain does (ir.Property's Method).
func (l *lowering) optionalCall(call *ast.Node) error {
	if call.Flags&ast.NodeFlagsOptionalChain == 0 {
		return nil
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken != nil && call.AsCallExpression().QuestionDotToken == nil {
		if method := l.checker.GetSymbolAtLocation(callee); method != nil && len(method.Declarations) > 0 && !load.IsLibrary(ast.GetSourceFileOfNode(method.Declarations[0])) {
			return nil
		}
	}
	return l.notYet(call, "a call through ?. (an optional call)")
}

// callClosure lowers a call through a function value.
func (l *lowering) callClosure(node *ast.Node) (ir.Expression, error) {
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression), checker.SignatureKindCall)
	if len(signatures) == 1 && l.censusNeverRestSignature(signatures[0]) {
		// never[] admits a zero-argument call in TypeScript. The erased slot does
		// not retain a source signature to prove its required arguments or ABI.
		return nil, l.notYet(node, "a call through an erased never-rest callable marker")
	}
	closure, err := l.expression(node.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	if property, isProperty := closure.(ir.Property); isProperty {
		// object.name(...) through an interface: the object may be a class's, whose methods aren't
		// fields (ir.Property's Method).
		property.Method = true
		closure = property
	}
	arguments := []ir.Expression{}
	if len(signatures) == 1 && restParameterIndex(signatures[0]) >= 0 {
		arguments, err = l.packRestArguments(node.AsCallExpression(), l.checker.GetResolvedSignature(node))
		if err != nil {
			return nil, err
		}
	} else {
		for _, argument := range node.AsCallExpression().Arguments.Nodes {
			lowered, err := l.expression(argument)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, lowered)
		}
	}
	var returns ir.Type
	if result := l.checker.GetTypeAtLocation(node); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = l.representation(result); !isKnown {
			return nil, l.notYet(node, "a call returning "+l.checker.TypeToString(result))
		}
	}
	// Each argument is made what the function value takes: a number or undefined where it takes
	// number | undefined is packed as one.
	if signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression), checker.SignatureKindCall); len(signatures) == 1 {
		for index, parameter := range signatures[0].Parameters() {
			if index < len(arguments) {
				if takes, isKnown := l.censusCallableParameter(parameter); isKnown {
					arguments[index] = fit(arguments[index], takes)
				}
			}
		}
	}
	for _, argument := range arguments {
		if censusCallableSlotless(argument.Type()) {
			return nil, l.notYet(node, "passing "+typeName(argument.Type())+" to a function value")
		}
	}
	if censusCallableSlotless(returns) {
		return nil, l.notYet(node, "a function value returning "+typeName(returns))
	}
	return ir.CallClosure{Closure: closure, Arguments: arguments, Returns: returns}, nil
}
