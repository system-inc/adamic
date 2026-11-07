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
//   - Inferred from an argument: the argument is judged against its parameter's type with the type
//     parameter replaced by its constraint, every other bound as the call binds it, as that value, so a
//     fresh literal stays fresh and a refusal names the part of the argument it's about. An argument tsc
//     wouldn't take into that slot holds the type parameter where it's taken in, not given out, and isn't
//     judged there.
//   - Written (read<X>(x), new Box<X>(), extends Box<X>), defaulted, or inferred from no argument (the
//     result's slot): the type argument is judged against the constraint, at the type argument written,
//     or at the site.
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
func (lowering *lowering) constraintRelationRefusal(node *ast.Node) error {
	return lowering.typeArgumentBindings(node, lowering.constraintRelation)
}

func init() {
	refusalRules.register(3350, "constraintRelationRefusal", (*lowering).constraintRelationRefusal)
}

// constraintRelation judges one binding, as constraintRelationRefusal says.
func (lowering *lowering) constraintRelation(binding typeArgumentBinding) error {
	if lowering.libraryTypeParameter(binding.parameter) {
		return nil
	}
	if binding.argument == nil {
		return lowering.notYet(binding.site, "a generic instantiation whose type argument for "+
			lowering.checker.TypeToString(binding.parameter)+
			", constrained to "+
			lowering.checker.TypeToString(binding.constraint)+
			", can't be read")
	}
	if binding.argument == binding.constraint {
		return nil
	}
	if binding.written != nil {
		return lowering.judgeBinding(binding.written, nil, binding.argument, binding.constraint)
	}
	judged, boundArgumentsRefusalError := lowering.boundArgumentsRefusal(binding)
	if boundArgumentsRefusalError != nil || judged {
		return boundArgumentsRefusalError
	}
	return lowering.judgeBinding(binding.site, nil, binding.argument, binding.constraint)
}

// boundArgumentsRefusal judges each argument a call's or a new's binding can be inferred from: one
// whose parameter's type, with the binding's type parameter as its constraint, is another slot than
// the one the call gives it. judged is whether there was one.
func (lowering *lowering) boundArgumentsRefusal(
	binding typeArgumentBinding,
) (judged bool, boundArgumentsRefusalError error) {
	if binding.generic == nil ||
		binding.site.Kind != ast.KindCallExpression && binding.site.Kind != ast.KindNewExpression {
		return false, nil
	}
	typeParameters := binding.generic.TypeParameters()
	bound := make([]*checker.Type, len(typeParameters))
	for index, typeParameter := range typeParameters {
		bound[index] = instantiateType(lowering.checker, typeParameter, binding.mapper)
		if typeParameter == binding.parameter {
			bound[index] = binding.constraint
		}
	}
	asConstraint := newTypeMapper(typeParameters, bound)
	parameters := binding.generic.Parameters()
	for index, argument := range binding.site.Arguments() {
		at := index
		rest := binding.generic.HasRestParameter() && at >= len(parameters)-1
		if rest {
			at = len(parameters) - 1
		}
		if at >= len(parameters) {
			continue
		}
		declared := lowering.checker.GetTypeOfSymbol(parameters[at])
		given, slot := instantiateType(
			lowering.checker,
			declared,
			binding.mapper,
		), instantiateType(lowering.checker, declared, asConstraint)
		if given == slot {
			continue
		}
		if argument.Kind == ast.KindSpreadElement {
			return true, lowering.notYet(argument, "a spread argument "+
				lowering.checker.TypeToString(binding.parameter)+
				"'s type argument can be inferred from, judged against its constraint "+
				lowering.checker.TypeToString(binding.constraint))
		}
		if rest {
			if !lowering.isArrayType(slot) {
				return true, lowering.notYet(argument, "an argument to a rest parameter that isn't an array, which "+
					lowering.checker.TypeToString(binding.parameter)+
					"'s type argument can be inferred from, judged against its constraint "+
					lowering.checker.TypeToString(binding.constraint))
			}
			slot = lowering.elementTypeOfArray(slot)
		}
		// Only a slot tsc's relation takes the argument into is one the relation rules can judge it in,
		// as they judge only what tsc accepted. One it doesn't is a place the type parameter is taken in,
		// not given out, a parameter of a callback (value: TIn) => value is TOut, and the binding is judged
		// by its other arguments, or as its type argument when it has none.
		if !lowering.checker.IsTypeAssignableTo(
			regularTypeOfObjectLiteral(lowering.checker, lowering.checker.GetTypeAtLocation(argument)),
			slot,
		) {
			continue
		}
		judged = true
		if judgeBindingError := lowering.judgeBinding(
			argument,
			argument,
			lowering.checker.GetTypeAtLocation(argument),
			slot,
		); judgeBindingError != nil {
			return true, judgeBindingError
		}
	}
	return judged, nil
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
func (lowering *lowering) judgeBinding(where, value *ast.Node, source, target *checker.Type) error {
	if target.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return nil // Nothing is read through an any or an unknown that a value can hide or be written.
	}
	if value != nil {
		value = ast.SkipParentheses(value)
		switch value.Kind {
		case ast.KindObjectLiteralExpression:
			return lowering.judgeBoundObject(value, target)
		case ast.KindArrayLiteralExpression:
			return lowering.judgeBoundArray(value, target)
		case ast.KindConditionalExpression:
			conditional := value.AsConditionalExpression()
			for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
				if judgeBindingError := lowering.judgeBinding(
					branch,
					branch,
					lowering.checker.GetTypeAtLocation(branch),
					target,
				); judgeBindingError != nil {
					return judgeBindingError
				}
			}
			return nil
		}
	}
	var found *widening
	if value != nil {
		found = lowering.freshOrWidened(value, source, target)
	} else {
		found = lowering.widened(source, target, map[[2]*checker.Type]bool{})
	}
	if found != nil {
		return lowering.wideningRefusal(where, source, target, found)
	}
	var hidden *optionalWidening
	if value != nil {
		hidden = lowering.optionalValue(value, target)
	} else {
		hidden = lowering.optionalWidened(source, target, nil, map[[2]*checker.Type]bool{})
	}
	if hidden != nil {
		return lowering.optionalWideningRefusal(where, hidden)
	}
	if mismatch := lowering.nominalMismatch(source, target, map[[2]*checker.Type]bool{}); mismatch != nil {
		return lowering.nominalViewRefusal(where, mismatch)
	}
	return nil
}

// judgeBoundObject judges an object literal a binding is inferred from against target, as made as each
// member of target it can be: each property's value against the property it's written into, a spread's
// value against target past the names written after it, and a method's type against its property's.
// A property target doesn't declare is never seen through it. A computed name fails closed, and so does a
// literal no member of target takes, which is judged whole as its type.
func (lowering *lowering) judgeBoundObject(literal *ast.Node, target *checker.Type) error {
	own := lowering.checker.GetTypeAtLocation(literal)
	taken := false
	for _, member := range lowering.definedMembers(lowering.checker.GetNonNullableType(target)) {
		if !lowering.checker.IsTypeAssignableTo(regularTypeOfObjectLiteral(lowering.checker, own), member) {
			continue
		}
		taken = true
		properties := literal.AsObjectLiteralExpression().Properties.Nodes
		for index, property := range properties {
			if property.Kind == ast.KindSpreadAssignment {
				spread := property.AsSpreadAssignment().Expression
				if ast.SkipParentheses(spread).Kind == ast.KindObjectLiteralExpression {
					if judgeBindingError := lowering.judgeBinding(
						spread,
						spread,
						lowering.checker.GetTypeAtLocation(spread),
						member,
					); judgeBindingError != nil {
						return judgeBindingError
					}
					continue
				}
				skip := map[string]bool{}
				for _, after := range properties[index+1:] {
					if after.Name() != nil && after.Name().Kind != ast.KindComputedPropertyName {
						skip[after.Name().Text()] = true
					}
				}
				from := lowering.checker.GetTypeAtLocation(spread)
				if found := lowering.widenedProperties(
					from,
					member,
					skip,
					true,
					map[[2]*checker.Type]bool{},
				); found != nil {
					return lowering.wideningRefusal(property, from, member, found)
				}
				if hidden := lowering.optionalWidened(from, member, skip, map[[2]*checker.Type]bool{}); hidden != nil {
					return lowering.optionalWideningRefusal(property, hidden)
				}
				continue
			}
			name := property.Name()
			if name == nil || name.Kind == ast.KindComputedPropertyName {
				return lowering.notYet(
					property,
					"a computed name in an object literal a type argument is inferred from, judged against its constraint "+
						lowering.checker.TypeToString(target),
				)
			}
			slot, declared := lowering.carriedProperty(member, name)
			if !declared || slot == nil {
				continue
			}
			var judgeBindingError error
			switch property.Kind {
			case ast.KindPropertyAssignment:
				initializer := property.AsPropertyAssignment().Initializer
				judgeBindingError = lowering.judgeBinding(
					initializer,
					initializer,
					lowering.checker.GetTypeAtLocation(initializer),
					slot,
				)
			case ast.KindShorthandPropertyAssignment:
				judgeBindingError = lowering.judgeBinding(name, name, lowering.checker.GetTypeAtLocation(name), slot)
			default:
				// A method or an accessor: its own type, declared where it's written, seen as the slot's.
				judgeBindingError = lowering.judgeBinding(
					property,
					nil,
					lowering.checker.GetTypeOfSymbol(lowering.checker.GetSymbolOfDeclaration(property)),
					slot,
				)
			}
			if judgeBindingError != nil {
				return judgeBindingError
			}
		}
	}
	if !taken {
		return lowering.judgeBinding(literal, nil, own, target)
	}
	return nil
}

// judgeBoundArray judges an array literal a binding is inferred from against target: each element
// against the element type, or a tuple's against its place, and a spread's value against the array.
// Into anything else, or a tuple past a spread, it fails closed.
func (lowering *lowering) judgeBoundArray(literal *ast.Node, target *checker.Type) error {
	array := lowering.checker.GetNonNullableType(target)
	elements := literal.AsArrayLiteralExpression().Elements.Nodes
	for index, element := range elements {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		var slot *checker.Type
		switch {
		case element.Kind == ast.KindSpreadElement:
			spread := element.AsSpreadElement().Expression
			if judgeBindingError := lowering.judgeBinding(
				element,
				nil,
				lowering.checker.GetTypeAtLocation(spread),
				array,
			); judgeBindingError != nil {
				return judgeBindingError
			}
			continue
		case checker.IsTupleType(array):
			arguments := lowering.checker.GetTypeArguments(array)
			if index < len(arguments) && !slices.ContainsFunc(
				elements[:index],
				func(before *ast.Node) bool { return before.Kind == ast.KindSpreadElement },
			) {
				slot = arguments[index]
			}
		case lowering.isArrayType(array):
			slot = lowering.elementTypeOfArray(array)
		}
		if slot == nil {
			return lowering.notYet(
				element,
				"an element of an array literal a type argument is inferred from, whose type in its constraint "+
					lowering.checker.TypeToString(target)+
					" can't be read",
			)
		}
		if judgeBindingError := lowering.judgeBinding(
			element,
			element,
			lowering.checker.GetTypeAtLocation(element),
			slot,
		); judgeBindingError != nil {
			return judgeBindingError
		}
	}
	return nil
}

// libraryTypeParameter is whether a type parameter is declared in the bundled library or the prelude,
// whose bindings the relation rule leaves alone (the library exemption, above).
func (lowering *lowering) libraryTypeParameter(parameter *checker.Type) bool {
	symbol := parameter.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return file != nil && (load.IsLibrary(file) || load.IsPrelude(file))
}
