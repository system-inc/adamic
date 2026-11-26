package lower

import (
	"errors"
	"strconv"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) isArrayPredicateCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return call.QuestionDotToken == nil && len(call.Arguments.Nodes) == 1 && callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken == nil && callee.Name().Text() == "isArray" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") && l.arrayPredicateMember(callee)
}

// The prelude's readonly unknown[] overload merges into the library's isArray, so the builtin is
// the member whose every declaration is the library's or the prelude's, never the program's.
func (l *lowering) arrayPredicateMember(callee *ast.Node) bool {
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
			if file := ast.GetSourceFileOfNode(declaration); !load.IsLibrary(file) && !load.IsPrelude(file) {
				return false
			}
		}
	}
	return true
}

func (l *lowering) arrayIsArray(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isArrayPredicateCall(node) {
		return nil, false, nil
	}
	argument := node.AsCallExpression().Arguments.Nodes[0]
	if !l.admitArrayPredicate(argument, l.checker.GetTypeAtLocation(argument)) && !l.exactObject(argument, 0) {
		return nil, true, l.notYet(argument, "Array.isArray on a tuple or an erased object/any/unknown view")
	}
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayIsArray{Value: fit(value, ir.Union)}, true, nil
}

// isArray's predicate, the prelude's readonly unknown[] overload in place of the bundled any[]
// one, describes the runtime test, not an element contract. Only array members already promised
// by the declared input survive.
func (l *lowering) arrayPredicateMembers(proven *checker.Type, array bool) *checker.Type {
	members := []*checker.Type{}
	var collect func(*checker.Type)
	collect = func(member *checker.Type) {
		member = l.concrete(member)
		if member.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range member.Types() {
				collect(part)
			}
			return
		}
		if member.Flags()&checker.TypeFlagsUnknown != 0 {
			if array {
				members = append(members, predicateReadonlyArray(l.checker, member, true))
			} else {
				members = append(members, member)
			}
			return
		}
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return
		}
		_, _, nodeArray := l.nodeArrayLayoutOf(member)
		if (l.isArrayType(member) || nodeArray) == array {
			members = append(members, member)
		}
	}
	collect(proven)
	return l.checker.GetUnionType(members)
}

// Tuples currently have an object representation. Erased object views can hide
// one too, so the native array brand cannot decide their JavaScript arrayness.
// A plain object arm is in the domain only where a property rules an array out
// there (arrayExcludedArm), or its index does (arrayExcludedIndexArm).
func (l *lowering) arrayPredicateDomain(proven *checker.Type) bool {
	return l.arrayPredicateArms(proven, nil)
}

// admitArrayPredicate is arrayPredicateDomain for a test lowering builds at node, the array brand's: each
// arm it admits by arrayExcludedArm or arrayExcludedIndexArm is recorded with node, for checkArrayArmViews to hold
// once lowering is done.
func (l *lowering) admitArrayPredicate(node *ast.Node, proven *checker.Type) bool {
	var arms []*checker.Type
	if !l.arrayPredicateArms(proven, &arms) {
		return false
	}
	if len(arms) > 0 {
		l.arrayArmSites = append(l.arrayArmSites, arrayArmSite{node: node, arms: arms})
	}
	return true
}

// arrayPredicateArms is arrayPredicateDomain, adding to arms, when it isn't nil, each arm arrayExcludedArm or
// arrayExcludedIndexArm admits.
func (l *lowering) arrayPredicateArms(proven *checker.Type, arms *[]*checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.arrayPredicateArms(member, arms) {
				return false
			}
		}
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnknown != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 || checker.IsTupleType(proven) {
		return false
	}
	if _, _, known := l.nodeArrayLayoutOf(proven); known {
		return true
	}
	if l.isArrayType(proven) {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(proven) || l.isLibraryType(proven, "Uint8Array", "Int32Array", "Float64Array") {
		return true
	}
	if !l.arrayExcludedArm(proven) && !l.arrayExcludedIndexArm(proven) {
		return false
	}
	if arms != nil {
		*arms = append(*arms, proven)
	}
	return true
}

// arrayExcludedArm is whether proven, a plain object type, rules an array out (Adamic's ruling on #s8gy89r): it
// declares a required property no type stored as an array can have, one that isn't numeric-named, isn't length,
// and isn't in arrayStoredPropertyNames, which holds lib's Array and ReadonlyArray members and every name a type
// of the closed world stored as an array declares or inherits. Every value at the arm has that property and no
// array has it, so none is an array, and a list of array types the program grows can't outgrow the test.
// Asking whether readonly unknown[] is assignable to the arm was ruled out, since it fails open: a tuple, an
// element-typed interface and tsc's TextRange each take an array readonly unknown[] doesn't. A class is no arm
// here, since its instance is in the domain above: `class X extends Array` is NotYet where it's constructed
// (Language's a01 on #mzc2mmn), so no class instance is an array.
//
// "Every value at the arm has the property" holds for the values the program builds. A checked view checks at a
// read, never at its cast, so a view can put a value behind the arm that lacks the property, unread, and a tuple
// held as an object behind it reads as no array to the brand where Node calls it one (Lower's probe on #9z2ksv7:
// `[number, number] | Other` cast as Named, overlapViewCast, printed named natively where Node printed array,
// J-k8ssmqr). checkArrayArmViews holds every arm admitted here against every checked view the program makes, so
// none puts a value stored as an array behind it. The other roads a tuple could take to an arm unread stop today,
// each measured in .ts (J-k8ssmqr) and pinned in TestArrayPredicateArmViewRoadsStop, so a road that opens goes red
// there and the pass and this domain are judged again:
//   - from unknown (`hidden as Named`), any (`JSON.parse(...) as Named`), object (`o as Named`), {} (`e as Named`),
//     and a function's unknown parameter (`value as Named`): castProof's "a cast the runtime can't check";
//   - a union member cast (`(Named | [number, number]) as Named`): unionMemberProof's same refusal, with no tag;
//   - a structural view over an interface holding a tuple: "optional property ... absent from structural source
//     [number, number]" (adamic/no-optional-widening), where the tuple is assigned.
func (l *lowering) arrayExcludedArm(proven *checker.Type) bool {
	if l.arrayStored(proven, map[*checker.Type]bool{}) {
		return false
	}
	stored := l.arrayStoredPropertyNames()
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsOptional == 0 && property.Name != "length" && !numericPropertyName(property.Name) && !stored[property.Name] {
			return true
		}
	}
	return false
}

// arrayExcludedIndexArm is the arm rule's second form, for an index-only record (rulings.md:260's amendment, Nov 26,
// @system_adamic on #gcec0bx): proven, a plain object type, rules an array out by its index instead of by a property.
// The premise, in Adamic's words: at an index-only arm, every value assignable to it without a hatch has every property
// readable as the index's value type. Every array-stored type has callable members (its methods), and condition 2 keeps
// every callable out of that value type. So no array-stored value reaches the arm except through a hatch, and condition
// 4's view check covers the hatch. That makes :260's "a kept arm that no array could be" hold by the index instead of by
// a property. Language's attack on the premise (#gcec0bx) found it stronger: tsc gives an implicit index signature only
// to an object literal's type, never to Array, ReadonlyArray or a tuple, so every array-stored type fails a string index
// (TS2322), even `[]` and `readonly never[]`, and its callable members are a second wall. It also found why the index
// must be a string's: a number-only index arm takes an array tsc-clean (`const asArm: NumberOnly = ["a"]`, her i01),
// since a number index constrains only numeric keys, and a symbol-only or template-pattern-only arm is no string index
// either (her i05 and i02). The arm is
// admitted only when all four hold, and any other keeps the stop and its words:
//  1. proven isn't stored as an array, and has no property, exactly one index signature, and that one a string's: a
//     number, symbol or template-pattern index alone fails, and so does one beside the string's; and no call or
//     construct signature (a class instance never gets here: arrayPredicateArms takes it first);
//  2. the index's value type admits no callable and no any, unknown, object or {} (indexValueAdmitsNoCallable);
//  3. no type stored as an array in the closed world declares a string index of its own (arrayStoredIndex, found by
//     arrayStoredPropertyNames' walk, failing closed as it does): one whose index its methods fit would be assignable
//     to the arm with no hatch;
//  4. arrayPredicateArms records the arm in arms exactly as it does one arrayExcludedArm admits, so checkArrayArmViews
//     holds it against every checked view the program makes.
//
// Lowering reads an instantiation's types (concrete), so in a generic function the index's value is what the
// instantiation made it; a type parameter is met here only where no instantiation fixes it, and is read at its
// constraint.
func (l *lowering) arrayExcludedIndexArm(proven *checker.Type) bool {
	if l.arrayStored(proven, map[*checker.Type]bool{}) || len(l.checker.GetPropertiesOfType(proven)) != 0 {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(proven)
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	if !l.indexValueAdmitsNoCallable(infos[0].ValueType(), map[*checker.Type]bool{}) {
		return false
	}
	l.arrayStoredPropertyNames()
	return !l.arrayStoredIndex
}

// indexValueAdmitsNoCallable is arrayExcludedIndexArm's condition 2 on value, an index-only arm's index value type: no
// value of it can be a function, and it's none of any, unknown, object or {}, which each take one. It's read through
// aliases (the type an alias names), unions and intersections, member by member, and through a type parameter, an
// indexed access or a conditional at its base constraint, so one with no constraint, or whose constraint fails, fails;
// keyof, a template literal and a string mapping are strings or keys whatever they're instantiated with, and pass.
// An object member passes only when it isn't callable: no call or construct signature, not lib's Function, and not
// empty, since every function is assignable to {}. Its properties aren't walked: the index value is what every property
// of the arm reads as, and the question is only whether that type can be a function. seen ends a value type that
// refers to itself (the JSON union holds itself): a member met again was judged where it was first met.
func (l *lowering) indexValueAdmitsNoCallable(value *checker.Type, seen map[*checker.Type]bool) bool {
	if value == nil {
		return false
	}
	if seen[value] {
		return true
	}
	seen[value] = true
	flags := value.Flags()
	switch {
	case flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0:
		return false
	case flags&checker.TypeFlagsUnionOrIntersection != 0:
		for _, member := range value.Types() {
			if !l.indexValueAdmitsNoCallable(member, seen) {
				return false
			}
		}
		return true
	case flags&checker.TypeFlagsInstantiableNonPrimitive != 0:
		constraint := l.checker.GetBaseConstraintOfType(value)
		return constraint != nil && constraint != value && l.indexValueAdmitsNoCallable(constraint, seen)
	case flags&checker.TypeFlagsObject != 0:
		if len(l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(value, checker.SignatureKindConstruct)) != 0 || l.isLibraryType(value, "Function", "CallableFunction", "NewableFunction") {
			return false
		}
		for _, info := range l.checker.GetIndexInfosOfType(value) {
			// An index of any takes every object, a function's too, as tsc relates it.
			if info.ValueType().Flags()&checker.TypeFlagsAny != 0 {
				return false
			}
		}
		return len(l.checker.GetPropertiesOfType(value)) != 0 || len(l.checker.GetIndexInfosOfType(value)) != 0
	}
	return true
}

// numericPropertyName is whether name reads as a number, an index an array may hold an element at: `0`, `"1"`,
// and, failing closed, any spelling ParseFloat takes (`1.5`, `1e3`, `NaN`, `Infinity`).
func numericPropertyName(name string) bool {
	_, err := strconv.ParseFloat(name, 64)
	return err == nil || errors.Is(err, strconv.ErrRange)
}

// arrayStored is whether a value of proven is stored as an array: an array or a tuple, a phantom extension of one
// (phantomExtension, tsc's SortedArray), lib's Array or ReadonlyArray, an interface or class whose base chain
// reaches one (NodeArray's), or an intersection holding one.
func (l *lowering) arrayStored(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range proven.Types() {
			if l.arrayStored(member, seen) {
				return true
			}
		}
		return false
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if l.isArrayType(proven) || checker.IsTupleType(proven) || l.isLibraryType(proven, "Array", "ReadonlyArray") {
		return true
	}
	if base, _ := l.phantomExtension(proven); base != nil && l.arrayStored(base, seen) {
		return true
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsClassOrInterface == 0 {
		return false
	}
	for _, base := range l.checker.GetBaseTypes(target) {
		if l.arrayStored(base, seen) {
			return true
		}
	}
	return false
}

// arrayStoredPropertyNames is every property name a type stored as an array can have, found once: lib's Array
// and ReadonlyArray members, and every name each type of the closed world that arrayStored calls an array declares
// or inherits. The closed world is each interface, class and type alias every file the program holds declares
// (the files kindTagCarriersOf walks, never moduleOrder's runnable modules alone, which leave out a module the
// program only imports types from), each intersection written in one, since `T & { pos: number }` over an array
// type argument is an array holding pos, and each Object.assign's result. An intersection holding a type parameter
// counts as one stored as an array, failing closed: its argument may be one. The bundled library and the prelude
// are walked too, for their interfaces and classes that extend another, such as lib's RegExpExecArray, whose index
// and input an array the program holds can have. The same walk sets arrayStoredIndex, whether any of those types has
// an index signature other than a number's (a string's, failing closed on any other key), for arrayExcludedIndexArm.
func (l *lowering) arrayStoredPropertyNames() map[string]bool {
	if l.arrayStoredNames != nil {
		return l.arrayStoredNames
	}
	names := map[string]bool{}
	index := false
	add := func(proven *checker.Type) {
		for _, property := range l.checker.GetPropertiesOfType(proven) {
			names[property.Name] = true
		}
		for _, info := range l.checker.GetIndexInfosOfType(proven) {
			if info.KeyType().Flags() != checker.TypeFlagsNumber {
				index = true
			}
		}
	}
	add(predicateReadonlyArray(l.checker, l.checker.GetUnknownType(), false))
	add(predicateReadonlyArray(l.checker, l.checker.GetUnknownType(), true))
	var collect func(*checker.Type)
	collect = func(proven *checker.Type) {
		if proven == nil {
			return
		}
		switch {
		case proven.Flags()&checker.TypeFlagsUnion != 0:
			for _, member := range proven.Types() {
				collect(member)
			}
		case proven.Flags()&checker.TypeFlagsIntersection != 0:
			for _, member := range proven.Types() {
				if member.Flags()&checker.TypeFlagsTypeParameter != 0 || l.arrayStored(member, map[*checker.Type]bool{}) {
					add(proven)
					return
				}
			}
		case l.arrayStored(proven, map[*checker.Type]bool{}):
			add(proven)
		}
	}
	declared := func(node *ast.Node) {
		if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil {
			collect(l.checker.GetDeclaredTypeOfSymbol(symbol))
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindTypeAliasDeclaration:
			declared(node)
		case ast.KindIntersectionType:
			collect(l.checker.GetTypeFromTypeNode(node))
		case ast.KindCallExpression:
			// Object.assign builds its target at an intersection no declaration names, an array's with pos and end
			// when its target is one, as enum_tag_closed_kind.go's build sites count it.
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				collect(l.checker.GetTypeAtLocation(node))
			}
		}
		node.ForEachChild(visit)
		return false
	}
	// The library declares thousands of members and no intersection a value is built at, so only its declarations
	// that extend another are asked, and its bodies are never walked.
	var visitLibrary ast.Visitor
	visitLibrary = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration:
			if ast.GetHeritageClause(node, ast.KindExtendsKeyword) != nil {
				declared(node)
			}
			return false
		case ast.KindModuleDeclaration, ast.KindModuleBlock:
			node.ForEachChild(visitLibrary)
		}
		return false
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if load.IsLibrary(file) || load.IsPrelude(file) {
			file.AsNode().ForEachChild(visitLibrary)
		} else {
			file.AsNode().ForEachChild(visit)
		}
	}
	l.arrayStoredNames, l.arrayStoredIndex = names, index
	return names
}

// The pinned checker owns construction of the readonly array type, just as it
// owns predicate flow facts. This introduces no any element contract.
//
//go:linkname predicateReadonlyArray github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).createArrayTypeEx
func predicateReadonlyArray(receiver *checker.Checker, element *checker.Type, readonly bool) *checker.Type

// arrayArmSite is a test lowering built at node whose array brand decides arms, each admitted by arrayExcludedArm or
// arrayExcludedIndexArm.
type arrayArmSite struct {
	node *ast.Node
	arms []*checker.Type
}

// checkArrayArmViews holds each arm a lowered Array.isArray, or a narrowed read checked by the array brand, leans on
// (arrayArmSites) against every checked view the program made, once lowering is done: a view reaching the arm
// (viewsReaching), over a source that may hold a value stored as an array, puts such a value behind the arm with
// nothing read, so the brand would call it no array where Node calls it one. That site is NotYet. A source may hold one when it names none, or any member of it is erased (any, unknown, a type
// parameter, object), an intersection, or an object type other than a class instance (a tuple, an array, an
// interface an array satisfies, or one a view of its own stands behind), failing closed.
func (l *lowering) checkArrayArmViews() error {
	for _, site := range l.arrayArmSites {
		for _, arm := range site.arms {
			for _, view := range l.viewsReaching(arm) {
				if !l.mayHoldArrayStored(view.source) {
					continue
				}
				source, where := "a value of no named type", "an unnamed site"
				if view.source != nil {
					source = l.checker.TypeToString(view.source)
				}
				if view.node != nil {
					where = l.program.Where(view.node)
				}
				target := "no named type"
				if view.target != nil {
					target = l.checker.TypeToString(view.target)
				}
				return l.notYet(site.node, "Array.isArray over "+l.checker.TypeToString(arm)+", which the checked view of "+target+" over "+source+" at "+where+" can hold an array behind, unread")
			}
		}
	}
	return nil
}

// mayHoldArrayStored is whether a value of source, a checked view's, may be stored as an array (checkArrayArmViews).
func (l *lowering) mayHoldArrayStored(source *checker.Type) bool {
	if source == nil {
		return true
	}
	for _, member := range castMembers(source) {
		switch {
		case member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsNonPrimitive|checker.TypeFlagsIntersection) != 0:
			return true
		case member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member):
			return true
		}
	}
	return false
}
