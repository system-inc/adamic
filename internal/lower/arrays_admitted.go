package lower

import (
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Scratch for #jjxtwzs, never to land: Adamic ruled that ir.Object promises an adamic object, so an array
// admitted where a checked view reads an object interface must be held as ir.Union, chosen in the type's
// representation (representationReason). This measures how many types that is when the set is narrowed to
// what the program's checked view sites can hand an array: a type T is array-admitted only where some site's
// source can be an array and its target, or a field's type at any depth of the target, is T, and T is an
// object whose declared fields are all optional and name nothing an array has (758624e2's viewArraysAdmitted).
// A site whose source value at a field can't be named counts as one that may hold an array.

// arraysAdmittedSite is one checked view site the pre-pass found: the type the value had before it, nil where
// there's none to name, and the type it's viewed as.
type arraysAdmittedSite struct {
	source, target *checker.Type
}

// collectArraysAdmitted is the pre-pass, run before any body is lowered and from the checker alone. It finds
// every site, then grows the set to its fixed point: a source field whose type is itself an all-optional
// object can hold an array only once that type is in the set, so each round may admit more.
func (l *lowering) collectArraysAdmitted(modules []*ast.SourceFile) {
	l.arraysAdmittedTypes = map[*checker.Type]bool{}
	l.arraysAdmittedSymbols = map[*ast.Symbol]bool{}
	sites := l.arraysAdmittedSites(modules)
	for {
		before := len(l.arraysAdmittedTypes)
		visited := map[[2]*checker.Type]bool{}
		for _, site := range sites {
			var nesting typeNesting
			l.arraysAdmittedConsider(site.target, site.source, visited, &nesting)
		}
		if len(l.arraysAdmittedTypes) == before {
			break
		}
	}
	l.arraysAdmittedReady = true
	l.arraysAdmittedReport()
}

// arraysAdmittedSites walks every module for the sites lowering views at: a cast (castValue, interfaceCast,
// the cast proofs), a type predicate's declaration with its claimed parameter as the source (predicates.go),
// a call through a predicate with its argument as the source (predicates_proof.go), a new of a constructor that
// isn't a class, given no source (new_class_value.go), and a relation, a value going into a slot tsc's
// contextual type or impliedTarget gives it (relation_views.go, refuseOptionalWidening's sites).
func (l *lowering) arraysAdmittedSites(modules []*ast.SourceFile) []arraysAdmittedSite {
	var sites []arraysAdmittedSite
	add := func(source, target *checker.Type) {
		if target != nil {
			sites = append(sites, arraysAdmittedSite{source: source, target: target})
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindAsExpression:
			as := node.AsAsExpression()
			if !(as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const") {
				add(l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node))
			}
		case ast.KindTypeAssertionExpression:
			add(l.checker.GetTypeAtLocation(node.AsTypeAssertion().Expression), l.checker.GetTypeAtLocation(node))
		case ast.KindTypePredicate:
			predicate := node.AsTypePredicateNode()
			if predicate.Type != nil {
				var source *checker.Type
				if owner := node.Parent; owner != nil && ast.IsFunctionLike(owner) && predicate.ParameterName != nil && predicate.ParameterName.Kind == ast.KindIdentifier {
					for _, parameter := range owner.Parameters() {
						if parameter.Name() != nil && parameter.Name().Kind == ast.KindIdentifier && parameter.Name().Text() == predicate.ParameterName.Text() {
							source = l.checker.GetTypeAtLocation(parameter.Name())
						}
					}
				}
				add(source, l.checker.GetTypeAtLocation(predicate.Type))
			}
		case ast.KindCallExpression:
			if signature := l.checker.GetResolvedSignature(node); signature != nil {
				claim := predicateOfSignature(l.checker, signature)
				if claim != nil && claim.Type() != nil && (claim.Kind() == checker.TypePredicateKindIdentifier || claim.Kind() == checker.TypePredicateKindAssertsIdentifier) {
					var source *checker.Type
					arguments := node.AsCallExpression().Arguments.Nodes
					if index := int(claim.ParameterIndex()); index >= 0 && index < len(arguments) {
						source = l.checker.GetTypeAtLocation(arguments[index])
					}
					add(source, claim.Type())
				}
			}
		case ast.KindNewExpression:
			callee := node.AsNewExpression().Expression
			constructs := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(callee), checker.SignatureKindConstruct)
			symbol := l.checker.GetSymbolAtLocation(callee)
			if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
				symbol = l.checker.GetAliasedSymbol(symbol)
			}
			if len(constructs) == 1 && (symbol == nil || symbol.Flags&ast.SymbolFlagsClass == 0) {
				add(nil, l.checker.GetReturnTypeOfSignature(constructs[0]))
			}
		case ast.KindShorthandPropertyAssignment:
			if literal := l.slotType(node.Parent); literal != nil {
				add(l.checker.GetTypeAtLocation(node.Name()), l.checker.GetTypeOfPropertyOfType(literal, node.Name().Text()))
			}
		case ast.KindSpreadAssignment:
			if literal := l.slotType(node.Parent); literal != nil {
				add(l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression), literal)
			}
		}
		if node.Kind != ast.KindParenthesizedExpression && l.isExpression(node) && optionalValueItself(node) {
			var slot *checker.Type
			if viewSite(node) {
				slot = l.slotType(node)
			}
			if slot == nil {
				slot = l.impliedTarget(node)
			}
			if slot != nil {
				add(l.checker.GetTypeAtLocation(node), slot)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return sites
}

// arraysAdmittedConsider walks target as a view builds contracts for it, a union's members, an array's or a
// tuple's elements, a record's values and an object's fields, carrying the source's type at each step, and
// adds each object type that qualifies where the source there can be an array. A callable is a leaf. A walk
// that meets a type instantiating itself deeper at each step admits every qualifying type from then on
// (arraysAdmittedEvery), which expansiveGenerics should have refused before this runs.
func (l *lowering) arraysAdmittedConsider(target, source *checker.Type, visited map[[2]*checker.Type]bool, nesting *typeNesting) {
	if target == nil || l.arraysAdmittedEvery {
		return
	}
	pair := [2]*checker.Type{target, source}
	if visited[pair] {
		return
	}
	visited[pair] = true
	ok, leave := nesting.enter(target)
	defer leave()
	if !ok {
		l.arraysAdmittedEvery = true
		return
	}
	present := source
	if present != nil {
		present = l.checker.GetNonNullableType(present)
	}
	switch {
	case target.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range target.Types() {
			l.arraysAdmittedConsider(member, source, visited, nesting)
		}
	case l.isArrayType(target):
		l.arraysAdmittedConsider(l.elementTypeOfArray(target), l.arraysAdmittedIndexed(present), visited, nesting)
	case checker.IsTupleType(target):
		var positions []*checker.Type
		if present != nil && checker.IsTupleType(present) {
			positions = l.checker.GetTypeArguments(present)
		}
		for index, element := range l.checker.GetTypeArguments(target) {
			if index < len(positions) {
				l.arraysAdmittedConsider(element, positions[index], visited, nesting)
			} else {
				l.arraysAdmittedConsider(element, l.arraysAdmittedIndexed(present), visited, nesting)
			}
		}
	case target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0:
		if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
			return
		}
		if l.arraysAdmittedShape(target) && l.arraysAdmittedSourceArray(source) {
			l.arraysAdmittedTypes[target] = true
			if symbol := target.Symbol(); symbol != nil {
				l.arraysAdmittedSymbols[symbol] = true
			}
		}
		for _, info := range l.checker.GetIndexInfosOfType(target) {
			var held *checker.Type
			if present != nil {
				held = l.checker.GetIndexTypeOfType(present, l.checker.GetStringType())
			}
			l.arraysAdmittedConsider(info.ValueType(), held, visited, nesting)
		}
		for _, property := range l.checker.GetPropertiesOfType(target) {
			var held *checker.Type
			if present != nil {
				if field := l.checker.GetPropertyOfType(present, property.Name); field != nil {
					held = l.checker.GetTypeOfSymbol(field)
				}
			}
			l.arraysAdmittedConsider(l.checker.GetTypeOfSymbol(property), held, visited, nesting)
		}
	}
}

// arraysAdmittedIndexed is what a source holds at a number index, nil where there's no source or it names none.
func (l *lowering) arraysAdmittedIndexed(present *checker.Type) *checker.Type {
	if present == nil {
		return nil
	}
	if l.isArrayType(present) {
		return l.elementTypeOfArray(present)
	}
	return l.checker.GetIndexTypeOfType(present, l.checker.GetNumberType())
}

// arraysAdmittedSourceArray is whether a value of source can be an array: one that can't be named, any,
// unknown or object, a type parameter whose constraint can (or that has none), a union with a member that can,
// an intersection each of whose members can, an array or a tuple, an object with an index signature or no
// property, or one whose required fields are all an array's. An object whose fields are all optional is weak,
// and tsc relates no array to it, so it holds one only where it's in the set already, which the fixed point
// grows.
func (l *lowering) arraysAdmittedSourceArray(source *checker.Type) bool {
	if source == nil {
		return true
	}
	flags := source.Flags()
	switch {
	case flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0:
		return true
	case flags&checker.TypeFlagsUnion != 0:
		for _, member := range source.Types() {
			if l.arraysAdmittedSourceArray(member) {
				return true
			}
		}
		return false
	case flags&checker.TypeFlagsIntersection != 0:
		for _, member := range source.Types() {
			if !l.arraysAdmittedSourceArray(member) {
				return false
			}
		}
		return true
	case flags&checker.TypeFlagsInstantiable != 0:
		constraint := l.checker.GetBaseConstraintOfType(source)
		return constraint == nil || constraint == source || l.arraysAdmittedSourceArray(constraint)
	case flags&checker.TypeFlagsObject != 0:
		if l.isArrayType(source) || checker.IsTupleType(source) || l.isLibraryType(source, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray") {
			return true
		}
		if len(l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(source, checker.SignatureKindConstruct)) != 0 {
			return false
		}
		if len(l.checker.GetIndexInfosOfType(source)) != 0 {
			return true
		}
		properties := l.checker.GetPropertiesOfType(source)
		if len(properties) == 0 {
			return true
		}
		required := false
		for _, property := range properties {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				continue
			}
			required = true
			if !l.arraysAdmittedArrayName(property.Name) {
				return false
			}
		}
		if required {
			return true
		}
		return l.arraysAdmittedHeld(source)
	}
	return false
}

// arraysAdmittedShape is 758624e2's viewArraysAdmitted asked of a type: a plain object, no array, tuple,
// callable, record or class instance, with at least one field, every field optional, and no field named as
// an array's own or inherited property. A library without Array to ask leaves it false. It's memoized a lowering.
func (l *lowering) arraysAdmittedShape(target *checker.Type) bool {
	if shaped, known := l.arraysAdmittedShapes[target]; known {
		return shaped
	}
	shaped := l.arraysAdmittedShapeAsked(target)
	if l.arraysAdmittedShapes == nil {
		l.arraysAdmittedShapes = map[*checker.Type]bool{}
	}
	l.arraysAdmittedShapes[target] = shaped
	return shaped
}

// arraysAdmittedShapeAsked is arraysAdmittedShape's answer, asked of the checker.
func (l *lowering) arraysAdmittedShapeAsked(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(target) || checker.IsTupleType(target) || isClassInstance(target) {
		return false
	}
	if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return false
	}
	if l.checker.GetGlobalSymbol("Array", ast.SymbolFlagsInterface, nil) == nil {
		return false
	}
	properties := l.checker.GetPropertiesOfType(target)
	if len(properties) == 0 {
		return false
	}
	for _, property := range properties {
		if property.Flags&ast.SymbolFlagsOptional == 0 || l.arraysAdmittedArrayName(property.Name) {
			return false
		}
	}
	return true
}

// arraysAdmittedArrayName is whether an array holds name: length, any name a number spells, Object's members
// and the __proto__ family (objectPrototypeName), and whatever the global Array interface declares. Where the
// library has no Array to ask, every name is one.
func (l *lowering) arraysAdmittedArrayName(name string) bool {
	if name == "length" || l.objectPrototypeName(name) {
		return true
	}
	if _, err := strconv.ParseFloat(name, 64); err == nil || errors.Is(err, strconv.ErrRange) {
		return true
	}
	array := l.checker.GetGlobalSymbol("Array", ast.SymbolFlagsInterface, nil)
	if array == nil {
		return true
	}
	return l.checker.GetPropertyOfType(l.checker.GetDeclaredTypeOfSymbol(array), name) != nil
}

// arraysAdmittedHeld is whether target is in the set: itself, or a type of the same shape whose symbol a member
// has, so an instantiation of a member's declaration (concrete under a type mapper) counts too.
func (l *lowering) arraysAdmittedHeld(target *checker.Type) bool {
	if l.arraysAdmittedEvery {
		return l.arraysAdmittedShape(target)
	}
	if l.arraysAdmittedTypes[target] {
		return true
	}
	if symbol := target.Symbol(); symbol != nil && l.arraysAdmittedSymbols[symbol] {
		return l.arraysAdmittedShape(target)
	}
	return false
}

// arraysAdmittedType is representationReason's question: whether a type is held as a union because an array
// may stand for it. Asked before the pre-pass ran, or of a lowering that never runs it, it fails closed: every
// type of the shape is.
func (l *lowering) arraysAdmittedType(target *checker.Type) bool {
	if !l.arraysAdmittedReady {
		return l.arraysAdmittedShape(target)
	}
	return l.arraysAdmittedHeld(target)
}

// arraysAdmittedReport appends one line a lowering run to the file ADAMIC_SCRATCH_ARRAYS_SET names: the root
// file, the set's count, and its types joined by " | ", sorted. Nothing when the variable is unset.
func (l *lowering) arraysAdmittedReport() {
	path := os.Getenv("ADAMIC_SCRATCH_ARRAYS_SET")
	if path == "" {
		return
	}
	var names []string
	for admitted := range l.arraysAdmittedTypes {
		names = append(names, l.checker.TypeToString(admitted))
	}
	sort.Strings(names)
	root := ""
	if l.entry != nil {
		root = l.program.FileName(l.entry)
	}
	count := strconv.Itoa(len(names))
	if l.arraysAdmittedEvery {
		count += " (every type of the shape)"
	}
	line := root + "\t" + count + "\t" + strings.Join(names, " | ") + "\n"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line)
}
