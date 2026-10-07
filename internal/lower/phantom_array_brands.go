package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// An array brand adds only absent, undefined-valued fields. Required undefined
// is phantom too (October 7 ruling); phantom_array_presence.go forbids observing
// presence or writing the checker-only members. Inherited library members
// still belong to the array; they are never phantom fields to be erased.
func (l *lowering) phantomArrayParts(proven *checker.Type) (*checker.Type, []*ast.Symbol) {
	proven = l.concrete(proven)
	if proven == nil || l.checker.IsArrayType(proven) || isClassInstance(proven) {
		return nil, nil
	}
	var base *checker.Type
	var fields []*ast.Symbol
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, part := range proven.Types() {
			if l.checker.IsArrayType(part) {
				if base != nil {
					return nil, nil
				}
				base = part
			} else {
				if part.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetSignaturesOfType(part, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(part, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(part)) != 0 {
					return nil, nil
				}
				fields = append(fields, l.checker.GetPropertiesOfType(part)...)
			}
		}
	} else if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&(checker.ObjectFlagsInterface|checker.ObjectFlagsReference) != 0 {
		target := proven
		if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
			target = proven.Target()
		}
		declaration := l.classNodeFor(proven)
		if declaration == nil || declaration.Kind != ast.KindInterfaceDeclaration {
			return nil, nil
		}
		var mapper *typeMapper
		if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
			mapper = l.typeMapperOf(declaration, proven)
		}
		for _, inherited := range l.checker.GetBaseTypes(target) {
			if mapper != nil {
				inherited = instantiateType(l.checker, inherited, mapper)
			}
			if branded := l.phantomArrayBase(inherited); branded != nil {
				inherited = branded
			}
			if l.checker.IsArrayType(inherited) {
				if base != nil {
					return nil, nil
				}
				base = inherited
			} else {
				return nil, nil
			}
		}
		if base == nil {
			return nil, nil
		}
		for _, field := range l.checker.GetPropertiesOfType(proven) {
			original := l.checker.GetPropertyOfType(base, field.Name)
			if original != nil && slices.Equal(field.Declarations, original.Declarations) {
				continue
			}
			fields = append(fields, field)
		}
		// A new index signature changes the element view, rather than merely naming it.
		for _, declaration := range proven.Symbol().Declarations {
			if declaration.Kind != ast.KindInterfaceDeclaration {
				return nil, nil
			}
			for _, member := range declaration.Members() {
				if member.Kind != ast.KindPropertySignature {
					return nil, nil
				}
			}
		}
	}
	if base == nil || len(fields) == 0 {
		return nil, nil
	}
	return base, fields
}

var phantomArrayNames = []string{"at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find", "findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes", "indexOf", "join", "keys", "lastIndexOf", "length", "map", "pop", "push", "reduce", "reduceRight", "reverse", "shift", "slice", "some", "sort", "splice", "toReversed", "toSorted", "toSpliced", "unshift", "values", "with"}

func arrayBrandMember(name string) bool {
	return slices.Contains(phantomArrayNames, name) || primitiveMember(checker.TypeFlagsString, name) && !slices.Contains(phantomStringNames, name)
}

func (l *lowering) phantomArrayBase(proven *checker.Type) *checker.Type {
	base, fields := l.phantomArrayParts(proven)
	if base == nil {
		return nil
	}
	for _, field := range fields {
		if arrayBrandMember(field.Name) || !phantomField(l.checker.GetTypeOfSymbol(field), true) {
			return nil
		}
	}
	return base
}

func (l *lowering) phantomArrayView(proven *checker.Type) *checker.Type {
	proven = l.concrete(proven)
	if base := l.phantomArrayBase(proven); base != nil {
		return base
	}
	return proven
}

func (l *lowering) phantomArrayRefusal(node *ast.Node) error {
	if node.Kind != ast.KindIntersectionType && node.Kind != ast.KindInterfaceDeclaration {
		return nil
	}
	base, fields := l.phantomArrayParts(l.checker.GetTypeAtLocation(node))
	// Written intersections can reduce to never when an own name conflicts.
	if base == nil && node.Kind == ast.KindIntersectionType {
		for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
			proven := l.checker.GetTypeAtLocation(part)
			if l.checker.IsArrayType(proven) {
				base = proven
			} else if proven.Flags()&checker.TypeFlagsObject != 0 {
				fields = append(fields, l.checker.GetPropertiesOfType(proven)...)
			}
		}
	}
	if base == nil {
		return nil
	}
	// Data-bearing array intersections are not phantom brands. Keep their existing
	// lowering and relation diagnostics rather than erasing their real fields.
	for _, field := range fields {
		if !phantomField(l.checker.GetTypeOfSymbol(field), true) {
			return nil
		}
	}
	for _, field := range fields {
		if arrayBrandMember(field.Name) {
			return &Refused{Where: l.program.Where(node), What: "an array brand member " + field.Name + " that exists on the array", Fix: "use a member name the array does not have on its own properties or prototype chain"}
		}
	}
	return nil
}

func (l *lowering) phantomArrayCast(node, expression *ast.Node, source, target *checker.Type) (bool, error) {
	if l.phantomArrayBase(source) == nil && l.phantomArrayBase(target) == nil {
		return false, nil
	}
	from, to := l.phantomArrayView(source), l.phantomArrayView(target)
	if !l.checker.IsArrayType(from) || !l.checker.IsArrayType(to) {
		if l.phantomArrayBase(target) != nil {
			return true, &Refused{Where: l.program.Where(node), What: "an array brand cast from a value that is not an array", Fix: "use a compatible array with proven elements, or narrow the value before adding an array brand (adamic/no-unchecked-cast)"}
		}
		return false, nil
	}
	if !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {
		return true, l.notYet(node, "an array brand cast that keeps elements weakly in one view and strongly in the other")
	}
	return true, l.provenTypesRelation(node, expression, from, to)
}
