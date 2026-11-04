package lower

import (
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
// typeArgumentBindings finds is judged by the registry with the constraint as its slot, exactly as an
// argument passed to a parameter of the constraint's type is:
//   - Inferred from an argument: the argument is judged against its parameter's type with the type
//     parameter replaced by its constraint, every other bound as the call binds it, by the whole
//     registry (refuseInSlot), so a fresh literal stays fresh and a refusal is the direct twin's own,
//     at the argument. A function written in place there was typed by the checker from the real
//     instantiation, not lib's any, so it's judged on the type-to-type road against that slot as that
//     value, its expression body as what it returns (boundArgumentRefusal). An argument tsc wouldn't
//     take into that slot holds the type parameter where it's taken in, not given out, and isn't judged
//     there.
//   - Written (read<X>(x), new Box<X>(), extends Box<X>), defaulted, or inferred from no argument (the
//     result's slot): the type argument is judged against the constraint on the registry's
//     type-to-type road (typeRelationRefusal), at the type argument written, or at the site.
//   - A binding that can't be read is not yet, failing closed.
// A library generic's body is the runtime's own, not a generic body proven under its constraint, so its
// bindings are left to the rules at its arguments.

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
		return l.typeRelationRefusal(binding.written, binding.argument, binding.constraint)
	}
	judged, err := l.boundArgumentsRefusal(binding)
	if err != nil || judged {
		return err
	}
	return l.typeRelationRefusal(binding.site, binding.argument, binding.constraint)
}

// boundArgumentsRefusal judges each argument a call's or a new's binding can be inferred from: one
// whose parameter's type, with the binding's type parameter as its constraint, is another slot than
// the one the call gives it. judged is whether there was one.
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
		if rest {
			if !l.checker.IsArrayType(slot) {
				return true, l.notYet(argument, "an argument to a rest parameter that isn't an array, which "+l.checker.TypeToString(binding.parameter)+"'s type argument can be inferred from, judged against its constraint "+l.checker.TypeToString(binding.constraint))
			}
			slot = l.checker.GetElementTypeOfArrayType(slot)
		}
		// Only a slot tsc's relation takes the argument into is one the registry can judge it in, as it
		// judges only what tsc accepted. One it doesn't is a place the type parameter is taken in, not
		// given out, a parameter of a callback (value: TIn) => value is TOut, and the binding is judged by
		// its other arguments, or on the type-to-type road when it has none.
		if !l.checker.IsTypeAssignableTo(l.checker.GetTypeAtLocation(argument), slot) {
			continue
		}
		judged = true
		if err := l.boundArgumentRefusal(argument, slot); err != nil {
			return true, err
		}
	}
	return judged, nil
}

// boundArgumentRefusal judges an argument against its parameter typed by the constraint: by the whole
// registry, as a value written into that slot, or, for a function written in place, which refuseInSlot
// can't read, on the type-to-type road at the argument, as that value.
func (l *lowering) boundArgumentRefusal(argument *ast.Node, slot *checker.Type) error {
	switch ast.SkipParentheses(argument).Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return l.relationRefusal(argument, argument, l.checker.GetTypeAtLocation(argument), slot)
	}
	return l.refuseInSlot(argument, slot)
}

// typeRelationRefusal judges a type source seen as target, where no expression is the value, at where.
func (l *lowering) typeRelationRefusal(where *ast.Node, source, target *checker.Type) error {
	return l.relationRefusal(where, nil, source, target)
}

// relationRefusal judges a value of type source seen as target as the registry's relation rules judge
// two types, in its order: a mutable slot that can be written what source can't hold (widened,
// adamic/invariant-mutable), an optional property target declares that source can hide (optionalWidened,
// the hiding rule), and nominal ancestry (nominalMismatch, adamic/nominal-class). When value is the
// expression that is the value, it's judged as that value: a fresh container's own slots stay covariant
// (freshOrWidened), and a function's expression body is what it returns (optionalValue).
func (l *lowering) relationRefusal(where, value *ast.Node, source, target *checker.Type) error {
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

// libraryTypeParameter is whether a type parameter is declared in the bundled library or the prelude.
func (l *lowering) libraryTypeParameter(parameter *checker.Type) bool {
	symbol := parameter.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return file != nil && (load.IsLibrary(file) || load.IsPrelude(file))
}
