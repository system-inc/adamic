package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// relativeRefinedField is a field of an untagged predicate target that its source doesn't already
// guarantee (relativePredicateFields), checked at the call: of is the representation its checked read
// takes, and undefined says the target holds it as undefined alone.
type relativeRefinedField struct {
	name      string
	declared  *checker.Type
	of        ir.Type
	undefined bool
}

// relativePredicateFields judges an untagged checked predicate target relative to its source, by Adamic's
// ruling 11 on the scout pass (#bbssdcy, Nov 24), which amends the letter of #c7w6j0y's rule 3 ("unless a
// structural check reads every required field it claims") and keeps its intent. A target field needs no
// read at the call when every object member of the source already holds it: declared there by name, never
// answered by an index signature (Language's condition (b), her p03), required there when the target
// requires it (her condition (a), p02: under exactOptionalPropertyTypes an optional field's declared type
// passes assignability alone), and of a type castAssignable to the target's. The premise is rulings.md:19:
// a value a checked view admitted stays checked at its own read of every field, under every type it's
// later seen as, so a field skipped here trusts nothing.
//
// Every other field is refined, and checked at the call by its own membership. A required field of a
// number, a string or a boolean, its literals included, is read through the view's checked read, which
// checks its presence, its representation and its literals. A required field holding a plain untagged
// record is read the same way, which checks that it's an object and reads none of its structure: its
// structure is checked as a view at its read. A required field held as undefined alone is read where
// every source member declares it as a reference, and tested undefined.
//
// A refined field whose membership the compiler can't express leaves its target where it was, with the
// words it had: ok is false, and the caller takes checkedPredicateTargetCause's road, by Adamic's road (a)
// (#bbssdcy, 19:54:35Z). An optional field, a union of representations, null, an array, a class instance,
// a callable, a dictionary, and a Map, a ReadonlyMap, a Set or a ReadonlySet, which share one
// representation (expression.go's representation), so no read of it can tell one from another. That is
// decided by what a membership can express, never by a name, and no collection-kind test stands in for
// one. So is a target a view can't check at its reads, one that reaches a callable contract anywhere in it
// (family 9 A's Program): its view would refuse what is a NotYet today. The depth bound (untaggedRecord) is
// unchanged and still judges every target this road doesn't take.
func (l *lowering) relativePredicateFields(source, target *checker.Type) ([]relativeRefinedField, bool) {
	if !l.plainViewObject(target) || l.checker.GetPropertyOfType(target, "kind") != nil || l.reachesCallableContract(target, map[*checker.Type]bool{}) {
		return nil, false
	}
	var members []*checker.Type
	sources := []*checker.Type{source}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		sources = source.Types()
	}
	for _, member := range sources {
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0 {
			continue
		}
		// The call tests that the value is an object, so every other source member is one held as an
		// object: no class instance, callable, array or tuple.
		if member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || l.callableViewContract(member) || l.arrayLikeType(member) || checker.IsTupleType(member) {
			return nil, false
		}
		if of, known := l.representation(member); !known || of != ir.Object {
			return nil, false
		}
		members = append(members, member)
	}
	if len(members) == 0 {
		return nil, false
	}
	var refined []relativeRefinedField
	for _, property := range l.checker.GetPropertiesOfType(target) {
		optional := property.Flags&ast.SymbolFlagsOptional != 0
		declared := l.checker.GetTypeOfSymbol(property)
		held, references := true, true
		for _, member := range members {
			// GetPropertyOfType answers a declared property alone, never an index signature's.
			field := l.checker.GetPropertyOfType(member, property.Name)
			if field == nil {
				held, references = false, false
				continue
			}
			of := l.checker.GetTypeOfSymbol(field)
			if !optional && field.Flags&ast.SymbolFlagsOptional != 0 || !l.castAssignable(of, declared) {
				held = false
			}
			if represented, known := l.representation(of); field.Flags&ast.SymbolFlagsOptional != 0 || !known || represented != ir.Object {
				references = false
			}
		}
		if held {
			continue
		}
		if optional || accessorSymbol(property) || property.Flags&ast.SymbolFlagsMethod != 0 || l.accessorNames[property.Name] {
			return nil, false
		}
		field := relativeRefinedField{name: property.Name, declared: declared}
		of, known := l.representation(declared)
		switch {
		case declared.Flags()&checker.TypeFlagsUndefined != 0 && references:
			field.undefined = true
		case !known || declared.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0:
			return nil, false
		case of == ir.Number || of == ir.String || of == ir.Boolean:
			field.of = of
		case of == ir.Object && l.plainViewObject(declared) && !l.isLibrary(declared):
			field.of = of
		default:
			return nil, false
		}
		refined = append(refined, field)
	}
	return refined, true
}

// relativePredicateMembership is the membership of an untagged target relativePredicateFields takes, for a
// check whose failure stops (a true side, an asserts claim, an overload's result): the value is an object,
// and each refined field holds what the target says. The refined fields are read as the tagged road reads
// kind, through the view's checked read, which stops at a field the value lacks or holds otherwise, so it's
// never a membership a false side may test. The target's view is registered (relativeView), so every read
// narrowed to the target, a refined record's structure included, is checked at the read. site is where the
// checked reads say they failed. ok is false where relativePredicateFields leaves the target on the pin's
// road.
func (l *lowering) relativePredicateMembership(site, node *ast.Node, value ir.Expression, source, target *checker.Type) ([]ir.Statement, ir.Expression, bool, error) {
	source, target = l.concrete(source), l.concrete(target)
	if value.Type() != ir.Object || target.Flags()&checker.TypeFlagsUnion != 0 {
		return nil, nil, false, nil
	}
	refined, ok := l.relativePredicateFields(source, target)
	if !ok {
		return nil, nil, false, nil
	}
	if err := l.relativeView(node, source, target); err != nil {
		return nil, nil, true, err
	}
	// The checked reads name the value as the program wrote it, the call's argument, where the site is a
	// call with one; a predicate's value or an overload's result has no argument text of its own.
	subject := "predicate argument"
	if site != nil && site.Kind == ast.KindCallExpression {
		if claim := predicateOfSignature(l.checker, l.checker.GetResolvedSignature(site)); claim != nil {
			if index := int(claim.ParameterIndex()); index >= 0 && index < len(site.AsCallExpression().Arguments.Nodes) {
				subject = sourceExpression(site.AsCallExpression().Arguments.Nodes[index])
			}
		}
	}
	membership := censusCondition(value)
	for _, field := range refined {
		var test ir.Expression
		if field.undefined {
			test = ir.IsUndefined{Value: ir.Property{Object: value, Name: field.name, Of: ir.Object}}
		} else {
			// Keyed by the target, as the view registered its fields (viewSchema), so the read is checked
			// natively too (Program.CheckedRead) though no read in the program names the field.
			read := ir.Property{Object: value, Name: field.name, Of: field.of, View: subject + "." + field.name, ViewType: l.checker.TypeToString(field.declared), ViewWhere: l.program.Where(site), ViewAllowed: l.viewLiterals(field.declared), ViewReceiverTypeID: l.viewFieldReceiverTypeID(nil, target), ViewTypeID: l.viewTypeID(field.declared)}
			read.ViewContract = l.result.ViewContractTypes[read.ViewTypeID]
			if !l.result.CheckedRead(read) {
				return nil, nil, true, l.notYet(node, "a checked predicate target's refined field "+field.name+" its view doesn't check")
			}
			test = ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: read}, Right: ir.StringConstant{Index: l.constant(relativeTypeOf[field.of])}}
		}
		membership = ir.Binary{Operator: ir.And, Left: membership, Right: test}
	}
	return nil, membership, true, nil
}

// relativeView registers target's checked view over source as the tagged road registers its own (the view
// is keyed by target's type across the program), marked narrowing, so a read the checker narrows to target
// reads through it, and untagged, since nothing at the call tells source's members apart by a tag:
// viewPairGuard pairs every source arm with the target's.
func (l *lowering) relativeView(node *ast.Node, source, target *checker.Type) error {
	views := len(l.viewSources)
	if _, err := l.view(node, nil, source, target); err != nil {
		return err
	}
	l.viewSources[views].narrows = true
	for index := views; index < len(l.viewSources); index++ {
		l.viewSources[index].untagged = true
	}
	return nil
}

// relativePredicateRemainder judges the false side of an untagged claim this rule takes (Adamic on #4cvdkk3,
// Nov 24, yes to Diagnostics' proposal, [L]): #w3e1vr8's closed-remainder road applied to an untagged
// target. The remainder, what the argument narrows to where the predicate answers false, is judged relative
// to the same source by relativePredicateFields, and its view is registered, so a lying false stops at the
// first read of the value in the false branch, never at the branch's entry, by ruling 12 A's stop placement:
// a continuation that doesn't read it runs on as Node does. A remainder may hold undefined or null beside its
// one object member, which the view's reads already stop on. ok is false, and the false side keeps
// predicateMembership's road, for a claim this rule doesn't take, an argument that isn't a reference, a
// remainder of two object members, or one this rule can't take. Language attacks the remainder judgment.
func (l *lowering) relativePredicateRemainder(node *ast.Node, call *ast.CallExpression, index int, source, claim *checker.Type) (bool, error) {
	source, claim = l.concrete(source), l.concrete(claim)
	if claim.Flags()&checker.TypeFlagsUnion != 0 || index < 0 || index >= len(call.Arguments.Nodes) {
		return false, nil
	}
	if _, ok := l.relativePredicateFields(source, claim); !ok {
		return false, nil
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[index])
	if argument.Kind != ast.KindIdentifier && argument.Kind != ast.KindPropertyAccessExpression && argument.Kind != ast.KindElementAccessExpression && argument.Kind != ast.KindThisKeyword {
		return false, nil
	}
	incoming := l.checker.GetTypeAtLocation(argument)
	flow := &ast.FlowNode{Flags: ast.FlowFlagsFalseCondition, Node: call.AsNode(), Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
	remainder := l.concrete(predicateFlowType(l.checker, argument, incoming, incoming, nil, flow))
	if remainder.Flags()&checker.TypeFlagsNever != 0 {
		return false, nil
	}
	object := remainder
	if remainder.Flags()&checker.TypeFlagsUnion != 0 {
		object = nil
		for _, member := range remainder.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
			if object != nil {
				return false, nil
			}
			object = member
		}
		if object == nil {
			return false, nil
		}
	}
	if _, ok := l.relativePredicateFields(source, object); !ok {
		return false, nil
	}
	return true, l.relativeView(node, source, object)
}

// relativeTypeOf is what typeof answers for a refined field's checked read.
var relativeTypeOf = map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean", ir.Object: "object"}

// predicateMembershipHeld is predicateMembership for a check whose failure stops: an untagged target
// relativePredicateMembership takes, and every other target as predicateMembership lowers it.
func (l *lowering) predicateMembershipHeld(site, node *ast.Node, value ir.Expression, source, target *checker.Type, local func(string, ir.Type) int) ([]ir.Statement, ir.Expression, error) {
	if setup, membership, ok, err := l.relativePredicateMembership(site, node, value, source, target); ok {
		return setup, membership, err
	}
	return l.predicateMembership(node, value, source, target, local, 0)
}

// isLibrary is whether t is declared by TypeScript's or Node's own declarations, or the prelude.
func (l *lowering) isLibrary(t *checker.Type) bool {
	if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		file := ast.GetSourceFileOfNode(symbol.Declarations[0])
		return load.IsLibrary(file) || load.IsPrelude(file)
	}
	return false
}

// reachesCallableContract is whether a checked view of t would meet a callable contract anywhere in it: t
// itself, a union's or an intersection's member, an array's or a tuple's element, or a field or method of
// a record the program declares, at any depth. A library type's own members are the library's.
func (l *lowering) reachesCallableContract(t *checker.Type, visited map[*checker.Type]bool) bool {
	if visited[t] {
		return false
	}
	visited[t] = true
	if l.callableViewContract(t) {
		return true
	}
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range t.Types() {
			if l.reachesCallableContract(member, visited) {
				return true
			}
		}
		return false
	}
	if l.isArrayType(t) || checker.IsTupleType(t) {
		for _, element := range l.checker.GetTypeArguments(t) {
			if l.reachesCallableContract(element, visited) {
				return true
			}
		}
		return false
	}
	if t.Flags()&checker.TypeFlagsObject == 0 || l.isLibrary(t) {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		if property.Flags&ast.SymbolFlagsMethod != 0 || l.reachesCallableContract(l.checker.GetTypeOfSymbol(property), visited) {
			return true
		}
	}
	return false
}
