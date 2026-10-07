package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// provenRelation holds satisfies and erased upcasts to the same writable-slot and nominal
// proofs as assignments. Contextual typing may build a fresh literal at the target type;
// values it contains are still checked at their own assignment sites by the refusal walk.
func (l *lowering) provenRelation(where, expression *ast.Node, target *checker.Type) error {
	source := l.checker.GetTypeAtLocation(expression)
	refuse := func(part, fix string) error {
		return &Refused{Where: l.program.Where(where), What: "an unproven relation from " + l.checker.TypeToString(source) + " to " + l.checker.TypeToString(target) + ": " + part, Fix: fix}
	}
	if !l.checker.IsTypeAssignableTo(source, target) && !l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
		return refuse("the source is not assignable to the target", "use a compatible type or narrow the value before using it")
	}
	if found := l.freshOrWidened(expression, source, target); found != nil {
		return l.wideningRefusal(where, source, target, found)
	}
	if mismatch := l.nominalMismatch(source, target, map[[2]*checker.Type]bool{}); mismatch != nil {
		return refuse("nominal ancestry for "+l.checker.TypeToString(mismatch), "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)")
	}
	if field := l.optionalRelationFailure(source, target, expression, map[[2]*checker.Type]bool{}); field != "" {
		return refuse("optional field "+field+" has no proven compatible presence/type", "keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)")
	}
	return nil
}

// Optional fields absent from a structural view may exist with an unrelated type at runtime.
// Follow containers, callback arguments/results and nested fields, as the other relation walks do.
func (l *lowering) optionalRelationFailure(from, to *checker.Type, origin *ast.Node, seen map[[2]*checker.Type]bool) string {
	if from == nil || to == nil || from == to || seen[[2]*checker.Type{from, to}] {
		return ""
	}
	seen[[2]*checker.Type{from, to}] = true
	defer delete(seen, [2]*checker.Type{from, to})
	if origin != nil {
		origin = ast.SkipParentheses(origin)
	}
	exact := origin != nil && origin.Kind == ast.KindObjectLiteralExpression
	if exact {
		for _, property := range origin.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindSpreadAssignment {
				exact = false
			}
		}
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if field := l.optionalRelationFailure(member, to, origin, seen); field != "" {
				return field
			}
		}
		return ""
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range to.Types() {
			if l.checker.IsTypeAssignableTo(from, member) {
				if field := l.optionalRelationFailure(from, member, origin, seen); field != "" {
					return field
				}
			}
		}
		return ""
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.optionalRelationFailure(l.checker.GetBaseConstraintOfType(from), to, nil, seen)
	}
	if !l.structured(from) || !l.structured(to) {
		return ""
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		parameters := toSignatures[0].Parameters()
		if l.censusNeverRestSignature(toSignatures[0]) {
			parameters = nil
		}
		for index, parameter := range parameters {
			if index >= len(fromSignatures[0].Parameters()) {
				break
			}
			if field := l.optionalRelationFailure(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(fromSignatures[0].Parameters()[index]), nil, seen); field != "" {
				return "parameter." + field
			}
		}
		return l.optionalRelationFailure(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), nil, seen)
	}
	// Class type arguments are invariant even when the class only exposes readonly fields.
	if isClassInstance(to) {
		if declaration := l.classNodeFor(to); declaration != nil {
			if actual := l.classView(from, declaration); actual != nil {
				fromArguments, toArguments := l.checker.GetTypeArguments(actual), l.checker.GetTypeArguments(to)
				for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
					if field := l.optionalRelationFailure(fromArguments[index], toArguments[index], nil, seen); field != "" {
						return "type argument." + field
					}
					if field := l.optionalRelationFailure(toArguments[index], fromArguments[index], nil, seen); field != "" {
						return "type argument." + field
					}
				}
			}
		}
	}
	fresh := exact || (origin != nil && (origin.Kind == ast.KindArrayLiteralExpression || l.freshValue(origin)))
	for _, target := range l.containers(to) {
		for _, source := range l.containers(from) {
			if !l.sameContainer(source, target) {
				continue
			}
			fromArguments, toArguments := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			if origin != nil && origin.Kind == ast.KindArrayLiteralExpression && len(toArguments) == 1 && !checker.IsTupleType(target) {
				for _, element := range origin.AsArrayLiteralExpression().Elements.Nodes {
					if element.Kind == ast.KindSpreadElement {
						if field := l.optionalRelationFailure(l.checker.GetTypeAtLocation(element.AsSpreadElement().Expression), target, nil, seen); field != "" {
							return field
						}
					} else if field := l.optionalRelationFailure(l.checker.GetTypeAtLocation(element), toArguments[0], element, seen); field != "" {
						return "element." + field
					}
				}
				continue
			}
			if checker.IsTupleType(source) && !checker.IsTupleType(target) && len(toArguments) == 1 {
				elements := make([]*checker.Type, len(fromArguments))
				for index := range elements {
					elements[index] = toArguments[0]
				}
				toArguments = elements
			}
			mutable := !fresh && !l.isLibraryType(target, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(target) && target.TargetTupleType().IsReadonly())
			for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
				if field := l.optionalRelationFailure(fromArguments[index], toArguments[index], nil, seen); field != "" {
					return "element." + field
				}
				if mutable {
					if field := l.optionalRelationFailure(toArguments[index], fromArguments[index], nil, seen); field != "" {
						return "element." + field
					}
				}
			}
		}
	}
	if len(l.containers(to)) > 0 {
		return ""
	}
	for _, property := range l.checker.GetPropertiesOfType(to) {
		inside := l.checker.GetPropertyOfType(from, property.Name)
		if inside == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 && !exact {
				return property.Name
			}
			continue
		}
		if field := l.optionalRelationFailure(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(property), relationFieldExpression(origin, property.Name), seen); field != "" {
			return property.Name + "." + field
		}
		if !fresh && !l.checker.IsReadonlySymbol(property) && property.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod == 0 {
			if field := l.optionalRelationFailure(l.checker.GetTypeOfSymbol(property), l.checker.GetTypeOfSymbol(inside), nil, seen); field != "" {
				return property.Name + "." + field
			}
		}
	}
	return ""
}

// An explicitly written field may itself be an exact literal. Shorthands and spread fields
// remain structural views, whose runtime fields their types may hide.
func relationFieldExpression(origin *ast.Node, name string) *ast.Node {
	if origin == nil || origin.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	for _, property := range origin.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindPropertyAssignment && property.Name().Text() == name {
			return property.AsPropertyAssignment().Initializer
		}
	}
	return nil
}
