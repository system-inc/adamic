package lower

import (
	"slices"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Binding a type argument, inferred or written, to a constrained type parameter is a relation site
// (#xywvn06, Adamic's ruling): a generic body is proven once for every type argument its constraint
// allows (docs/rulings.md, the generic body), so a type argument that doesn't satisfy the constraint
// under Adamic's relation can't be bound. read<T extends Options>(options: T) called with a
// { readonly other: string } that holds overrideImportMode: "lie" reads the string as Options' number,
// which the hiding rule refuses at a direct call through a parameter typed Options. Each binding
// typeArgumentBindings finds is judged by one rule set (judgeBinding), the registry's relation rules for
// a value seen as a type: invariance (widened), the hiding rule and nominal ancestry, whichever road
// bound it, so an inferred binding and its written twin judge the same thing:
//   - Inferred from an argument whose parameter is the type parameter itself, or it joined with undefined
//     or null (read<T extends Options>(options: T)): the argument is the bound value, judged against the
//     constraint, every other type parameter as the call binds it, as that value, so a fresh literal stays
//     fresh and a refusal names the part of the argument it's about. An argument tsc wouldn't take into
//     the constraint holds the type parameter where it's taken in, not given out, and isn't judged there.
//   - Inferred from an argument where the type parameter is nested in its parameter's type (to: T[],
//     make: () => T), or from no argument (the result's slot): the binding is judged against the
//     constraint once, as itself, at the site (#c4vhbep). The argument isn't judged against its parameter
//     with the constraint put in for the type parameter, a slot no call makes: combine<T extends {}>'s
//     append(xs, ys) puts a T[] where append's T[] is combine's T[], and the substitution saw it as {}[],
//     a wall where nothing is widened. The argument against its parameter at the resolved binding is the
//     argument site's, which refuseWidening judges as it does any slot's (add(dogs, new Animal()), T
//     inferred as Animal, is Dog[] seen as Animal[] there), and what the body does through a T is reached
//     through the constraint, which judging the binding covers.
//   - Written (read<X>(x), new Box<X>(), extends Box<X>) or defaulted: the type argument is judged against
//     the constraint, at the type argument written, or at the site.
//   - A binding that can't be read is not yet, failing closed.
// A binding makes no slot and no view, since the body is lowered per instantiation, so the registry's
// representation stops (recordStorageView's storage gap) don't run at one. Declared keys at a binding are
// the declared-key rule's, which judges every binding typeArgumentBindings finds in .a first (1975); .ts
// keeps the checked read.
//
// The library exemption: a type parameter the library or the prelude declares isn't judged here. Its
// generic's body is the runtime's own, not a generic body Adamic proves once under the constraint, so the
// ruling's reason doesn't reach it, and lib's type-guard overloads (filter<S extends T>, find, every) bind
// S to a narrower type than the element on every guarded call. Its arguments keep every rule at their own
// sites, and the declared-key rule still judges its bindings in .a. Backed by data (#xywvn06): the
// admission delta with library bindings judged too, and lib_filter_hidden, a filter whose S hides a field
// of its element type, agreeing with Node natively.

// regularTypeOfObjectLiteral is a fresh object literal's type without its freshness, which tsc's
// relation would otherwise judge for excess properties: an argument { mode: 1, extra: "x" } bound to
// T extends Options binds T to its own type, extra and all, and Options never sees extra. It's the
// checker's own, pulled from its package as generic_function_value.go pulls instantiateSignatureInContextOf.
//
//go:linkname regularTypeOfObjectLiteral github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getRegularTypeOfObjectLiteral
func regularTypeOfObjectLiteral(receiver *checker.Checker, proven *checker.Type) *checker.Type

// constraintRelationRefusal judges each type argument node binds to a constrained type parameter.
func (l *lowering) constraintRelationRefusal(node *ast.Node) error {
	return l.typeArgumentBindings(node, l.constraintRelation)
}

func init() {
	refusalRules.register(3350, "constraintRelationRefusal", (*lowering).constraintRelationRefusal)
}

// constraintRelation judges one binding, as constraintRelationRefusal says.
func (l *lowering) constraintRelation(binding typeArgumentBinding) error {
	if l.libraryTypeParameter(binding.parameter) {
		return nil
	}
	if binding.argument == nil {
		return l.notYet(binding.site, "a generic instantiation whose type argument for "+l.checker.TypeToString(binding.parameter)+", constrained to "+l.checker.TypeToString(binding.constraint)+", can't be read")
	}
	if binding.argument == binding.constraint {
		return nil
	}
	if binding.written != nil {
		return l.judgeBinding(binding.written, nil, binding.argument, binding.constraint)
	}
	judged, err := l.boundArgumentsRefusal(binding)
	if err != nil || judged {
		return err
	}
	return l.judgeBinding(binding.site, nil, binding.argument, binding.constraint)
}

// boundArgumentsRefusal judges each argument a call's or a new's binding can be inferred from whose
// parameter is the binding's type parameter itself, or it joined with undefined or null, as that value
// against the constraint. judged is whether those arguments are all the binding is inferred from, so it
// needn't be judged as itself: an argument where the type parameter is nested leaves the binding to be
// judged against the constraint as itself, once (constraintRelation).
func (l *lowering) boundArgumentsRefusal(binding typeArgumentBinding) (judged bool, err error) {
	if binding.generic == nil || binding.site.Kind != ast.KindCallExpression && binding.site.Kind != ast.KindNewExpression {
		return false, nil
	}
	typeParameters := binding.generic.TypeParameters()
	bound := make([]*checker.Type, len(typeParameters))
	for index, typeParameter := range typeParameters {
		bound[index] = instantiateType(l.checker, typeParameter, binding.mapper)
		if typeParameter == binding.parameter {
			bound[index] = binding.constraint
		}
	}
	asConstraint := newTypeMapper(typeParameters, bound)
	parameters := binding.generic.Parameters()
	nested := false
	for index, argument := range binding.site.Arguments() {
		at := index
		rest := binding.generic.HasRestParameter() && at >= len(parameters)-1
		if rest {
			at = len(parameters) - 1
		}
		if at >= len(parameters) {
			continue
		}
		declared := l.checker.GetTypeOfSymbol(parameters[at])
		given, slot := instantiateType(l.checker, declared, binding.mapper), instantiateType(l.checker, declared, asConstraint)
		if given == slot {
			continue
		}
		if argument.Kind == ast.KindSpreadElement {
			return true, l.notYet(argument, "a spread argument "+l.checker.TypeToString(binding.parameter)+"'s type argument can be inferred from, judged against its constraint "+l.checker.TypeToString(binding.constraint))
		}
		// place is the argument's own parameter type as declared: a rest parameter's element.
		place := declared
		if rest {
			if !l.isArrayType(slot) {
				return true, l.notYet(argument, "an argument to a rest parameter that isn't an array, which "+l.checker.TypeToString(binding.parameter)+"'s type argument can be inferred from, judged against its constraint "+l.checker.TypeToString(binding.constraint))
			}
			slot, place = l.elementTypeOfArray(slot), l.elementTypeOfArray(declared)
		}
		// Only a slot tsc's relation takes the argument into is one the relation rules can judge it in,
		// as they judge only what tsc accepted. One it doesn't is a place the type parameter is taken in,
		// not given out, a parameter of a callback (value: TIn) => value is TOut, and the binding is judged
		// by its other arguments, or as its type argument when it has none.
		if !l.checker.IsTypeAssignableTo(regularTypeOfObjectLiteral(l.checker, l.checker.GetTypeAtLocation(argument)), slot) {
			continue
		}
		if !boundItself(place, binding.parameter) {
			nested = true
			continue
		}
		judged = true
		if err := l.judgeBinding(argument, argument, l.checker.GetTypeAtLocation(argument), slot); err != nil {
			return true, err
		}
	}
	return judged && !nested, nil
}

// boundItself is whether a parameter's declared type is the type parameter itself, or it joined with
// undefined or null, so the argument is the value the type parameter is bound to.
func boundItself(declared, parameter *checker.Type) bool {
	if declared == parameter {
		return true
	}
	if declared == nil || declared.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	found := false
	for _, member := range declared.Types() {
		switch {
		case member == parameter:
			found = true
		case member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) == 0:
			return false
		}
	}
	return found
}

// judgeBinding judges a binding as source seen as target, at where, by the registry's relation rules for
// a value seen as a type: a mutable slot that can be written what source can't hold, a readonly field
// made writable among them (widened, adamic/invariant-mutable), an optional property target declares
// that source can hide (optionalWidened, the hiding rule), and nominal ancestry (nominalMismatch,
// adamic/nominal-class), in the registry's order and in each rule's words. value is the expression that
// is the value, an argument a binding is inferred from or a part of one, or nil for a type argument no
// expression is. A value is judged as itself: a literal made right there is made as its slot, so its
// parts are judged against the slot's parts where they're written, as the registry judges a literal's
// parts at their own sites; a conditional is either branch; a fresh copy's own slots stay covariant
// (freshOrWidened); and a function's expression body is what it returns (optionalValue).
func (l *lowering) judgeBinding(where, value *ast.Node, source, target *checker.Type) error {
	if target.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return nil // Nothing is read through an any or an unknown that a value can hide or be written.
	}
	if value != nil {
		value = ast.SkipParentheses(value)
		switch value.Kind {
		case ast.KindObjectLiteralExpression:
			return l.judgeBoundObject(value, target)
		case ast.KindArrayLiteralExpression:
			return l.judgeBoundArray(value, target)
		case ast.KindConditionalExpression:
			conditional := value.AsConditionalExpression()
			for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
				if err := l.judgeBinding(branch, branch, l.checker.GetTypeAtLocation(branch), target); err != nil {
					return err
				}
			}
			return nil
		}
	}
	var found *widening
	if value != nil {
		found = l.freshOrWidened(value, source, target)
	} else {
		found = l.widened(source, target, map[[2]*checker.Type]bool{})
	}
	if found != nil {
		return l.wideningRefusal(where, source, target, found)
	}
	var hidden *optionalWidening
	if value != nil {
		hidden = l.optionalValue(value, target)
	} else {
		hidden = l.optionalWidened(source, target, nil, map[[2]*checker.Type]bool{})
	}
	if hidden != nil {
		return l.optionalWideningRefusal(where, hidden)
	}
	if mismatch := l.nominalMismatch(source, target, map[[2]*checker.Type]bool{}); mismatch != nil {
		return l.nominalViewRefusal(where, mismatch)
	}
	return nil
}

// judgeBoundObject judges an object literal a binding is inferred from against target, as made as each
// member of target it can be: each property's value against the property it's written into, a spread's
// value against target past the names written after it, and a method's type against its property's.
// A property target doesn't declare is never seen through it. A computed name fails closed, and so does a
// literal no member of target takes, which is judged whole as its type.
func (l *lowering) judgeBoundObject(literal *ast.Node, target *checker.Type) error {
	own := l.checker.GetTypeAtLocation(literal)
	taken := false
	for _, member := range l.definedMembers(l.checker.GetNonNullableType(target)) {
		if !l.checker.IsTypeAssignableTo(regularTypeOfObjectLiteral(l.checker, own), member) {
			continue
		}
		taken = true
		properties := literal.AsObjectLiteralExpression().Properties.Nodes
		for index, property := range properties {
			if property.Kind == ast.KindSpreadAssignment {
				spread := property.AsSpreadAssignment().Expression
				if ast.SkipParentheses(spread).Kind == ast.KindObjectLiteralExpression {
					if err := l.judgeBinding(spread, spread, l.checker.GetTypeAtLocation(spread), member); err != nil {
						return err
					}
					continue
				}
				skip := map[string]bool{}
				for _, after := range properties[index+1:] {
					if after.Name() != nil && after.Name().Kind != ast.KindComputedPropertyName {
						skip[after.Name().Text()] = true
					}
				}
				from := l.checker.GetTypeAtLocation(spread)
				if found := l.widenedProperties(from, member, skip, true, map[[2]*checker.Type]bool{}); found != nil {
					return l.wideningRefusal(property, from, member, found)
				}
				if hidden := l.optionalWidened(from, member, skip, map[[2]*checker.Type]bool{}); hidden != nil {
					return l.optionalWideningRefusal(property, hidden)
				}
				continue
			}
			name := property.Name()
			if name == nil || name.Kind == ast.KindComputedPropertyName {
				return l.notYet(property, "a computed name in an object literal a type argument is inferred from, judged against its constraint "+l.checker.TypeToString(target))
			}
			slot, declared := l.carriedProperty(member, name)
			if !declared || slot == nil {
				continue
			}
			var err error
			switch property.Kind {
			case ast.KindPropertyAssignment:
				initializer := property.AsPropertyAssignment().Initializer
				err = l.judgeBinding(initializer, initializer, l.checker.GetTypeAtLocation(initializer), slot)
			case ast.KindShorthandPropertyAssignment:
				err = l.judgeBinding(name, name, l.checker.GetTypeAtLocation(name), slot)
			default:
				// A method or an accessor: its own type, declared where it's written, seen as the slot's.
				err = l.judgeBinding(property, nil, l.checker.GetTypeOfSymbol(l.checker.GetSymbolOfDeclaration(property)), slot)
			}
			if err != nil {
				return err
			}
		}
	}
	if !taken {
		return l.judgeBinding(literal, nil, own, target)
	}
	return nil
}

// judgeBoundArray judges an array literal a binding is inferred from against target: each element
// against the element type, or a tuple's against its place, and a spread's value against the array.
// Into anything else, or a tuple past a spread, it fails closed.
func (l *lowering) judgeBoundArray(literal *ast.Node, target *checker.Type) error {
	array := l.checker.GetNonNullableType(target)
	elements := literal.AsArrayLiteralExpression().Elements.Nodes
	for index, element := range elements {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		var slot *checker.Type
		switch {
		case element.Kind == ast.KindSpreadElement:
			spread := element.AsSpreadElement().Expression
			if err := l.judgeBinding(element, nil, l.checker.GetTypeAtLocation(spread), array); err != nil {
				return err
			}
			continue
		case checker.IsTupleType(array):
			arguments := l.checker.GetTypeArguments(array)
			if index < len(arguments) && !slices.ContainsFunc(elements[:index], func(before *ast.Node) bool { return before.Kind == ast.KindSpreadElement }) {
				slot = arguments[index]
			}
		case l.isArrayType(array):
			slot = l.elementTypeOfArray(array)
		}
		if slot == nil {
			return l.notYet(element, "an element of an array literal a type argument is inferred from, whose type in its constraint "+l.checker.TypeToString(target)+" can't be read")
		}
		if err := l.judgeBinding(element, element, l.checker.GetTypeAtLocation(element), slot); err != nil {
			return err
		}
	}
	return nil
}

// libraryTypeParameter is whether a type parameter is declared in the bundled library or the prelude,
// whose bindings the relation rule leaves alone (the library exemption, above).
func (l *lowering) libraryTypeParameter(parameter *checker.Type) bool {
	symbol := parameter.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return file != nil && (load.IsLibrary(file) || load.IsPrelude(file))
}
