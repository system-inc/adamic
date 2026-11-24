package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// libraryMember follows the declaration, not the spelling: an own callback named toString is
// still an own field. Every declaration must be from the library to specialize a call.
func (l *lowering) memberSymbol(node *ast.Node) *ast.Symbol {
	symbol := l.checker.GetSymbolAtLocation(node)
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			symbol = l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(access.Expression), key.Text())
		}
	}
	return symbol
}

func (l *lowering) libraryMember(node *ast.Node) bool {
	return l.librarySymbol(l.memberSymbol(node))
}

// A union can combine an own callback with an inherited method. That still cannot be an
// unconditional own-field load: any library root is enough to refuse the read.
func (l *lowering) inheritedLibraryMember(node *ast.Node) bool {
	return l.inheritedLibrarySymbol(l.memberSymbol(node))
}

func (l *lowering) inheritedLibrarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return true
			}
		}
	}
	return false
}

// iteratorResultField reports whether done or value is read off a library iterator result as an own
// field, never a prototype member: a result object next returns holds both as its own on Node, whatever
// made it. It's the one home of that rule (#5t28afa), asked by every property read, element read and
// destructuring, in two arms:
//   - every library root of symbol is IteratorReturnResult's, both fields required and own. The prelude's
//     Map and Set iterators return one as their done arm, so inference reads that arm as the iterator's
//     return rather than its value;
//   - for a read off receiver (nil for a destructuring, which keeps the first arm alone, as before this
//     move), every member of the receiver's type is the library's IteratorYieldResult or
//     IteratorReturnResult, the result a hand-driven next() gives (#qk8rztp S2c-2a).
func (l *lowering) iteratorResultField(receiver *ast.Node, symbol *ast.Symbol) bool {
	if symbol == nil || (symbol.Name != "done" && symbol.Name != "value") {
		return false
	}
	if l.iteratorReturnResultRoots(symbol) {
		return true
	}
	if receiver == nil {
		return false
	}
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if !l.isLibraryType(member, "IteratorYieldResult", "IteratorReturnResult") {
			return false
		}
	}
	return true
}

// iteratorReturnResultRoots reports whether every library root of symbol is declared in
// IteratorReturnResult.
func (l *lowering) iteratorReturnResultRoots(symbol *ast.Symbol) bool {
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				continue
			}
			owner := declaration.Parent
			if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name().Text() != "IteratorReturnResult" {
				return false
			}
		}
	}
	return true
}

func (l *lowering) librarySymbol(symbol *ast.Symbol) bool {
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
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
		}
	}
	return true
}

func (l *lowering) prototypeRead(node *ast.Node, name string) error {
	if name == "isPrototypeOf" {
		return &Refused{Where: l.program.Where(node), What: "isPrototypeOf", Fix: "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant"}
	}
	return &Refused{Where: l.program.Where(node), What: "inherited library member " + name + " read as an own field", Fix: "prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)"}
}

func (l *lowering) objectPrototypeCall(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	return l.objectPrototypeCallArguments(node, receiver, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) objectPrototypeCallArguments(node, receiver *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !l.libraryMember(callee) {
		return nil, false, nil
	}
	switch name {
	case "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf":
	default:
		return nil, false, nil
	}
	if name == "isPrototypeOf" {
		return nil, true, &Refused{Where: l.program.Where(node), What: "isPrototypeOf", Fix: "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant"}
	}
	if name != "valueOf" && l.isLibraryType(l.checker.GetTypeAtLocation(receiver), errorClasses...) {
		if name == "hasOwnProperty" && callee.Kind == ast.KindPropertyAccessExpression &&
			ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression) == ast.SkipParentheses(receiver) &&
			l.prototypeHazardWithErrorTable(receiver, name, true) == "" {
			value, handled, err := l.hasOwnPropertyArguments(node, receiver, written)
			if own, ok := value.(ir.HasOwn); handled && err == nil && ok && own.Object.Type() == ir.Object {
				return ir.ObjectCall{Method: "hasOwnErrorView", Arguments: []ir.Expression{own.Object, own.Key}, Returns: ir.Boolean}, true, nil
			}
		}
		return nil, true, l.notYet(node, name+" on Error (its prototype and non-enumerable own descriptors differ from plain objects)")
	}
	if name == "toString" || name == "toLocaleString" {
		// valueOf answers the object itself on every library object with a prototype, an iterator and an Error
		// alike, which the arm below already gives. A RegExp match's groups object is tested for at run time below,
		// and a library object with a tag of its own reads it at run time (libraryTagAtRunTime).
		if library := l.libraryObjectOwnAnswer(receiver); library != "" && library != regexpGroupsAnswer {
			return nil, true, l.notYet(node, name+" on "+library+", whose own answer isn't Object.prototype's")
		}
	}
	of, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	if l.exactPlainObject(receiver) {
		of = ir.Object
	}
	// A RegExp match's groups object has a null prototype, so on Node a direct call of any of these on one finds
	// nothing to call and throws a TypeError, where a program's own object of the same type answers as
	// Object.prototype does: the value tells them apart, never the type (#eksbz4c). Object.prototype's own
	// method called through .call reaches the groups object as this and answers, so only a direct call is
	// tested. Each arm below that answers an object makes its answer in regexpGroupsMissingMethod.
	missing := false
	if access := ast.SkipParentheses(node.AsCallExpression().Expression); access.Kind == ast.KindPropertyAccessExpression && access.AsPropertyAccessExpression().Expression == receiver && l.regexpGroupsAdmitted(receiver) {
		if of != ir.Object {
			return nil, true, l.notYet(node, name+" on "+regexpGroupsAnswer+" held as "+typeName(of))
		}
		missing = true
	}
	if of == ir.Object || of == ir.Union {
		if reason := l.prototypeHazard(receiver, name); reason != "" {
			return nil, true, l.notYet(node, name+" through an object view ("+reason+")")
		}
	}
	if name == "toString" && of == ir.Number {
		return nil, false, nil
	} // Number's radix overload.
	if name == "toLocaleString" && of == ir.Number {
		return nil, true, &Refused{Where: l.program.Where(node), What: name + " on a number", Fix: "JavaScript uses locale-sensitive number formatting; use toString for deterministic formatting"}
	}
	if name == "toLocaleString" && of == ir.Array {
		element, err := l.elementType(receiver)
		if err != nil {
			return nil, true, err
		}
		if element != ir.String && element != ir.Boolean {
			return nil, true, &Refused{Where: l.program.Where(node), What: name + " on this array", Fix: "JavaScript formats each element with its own toLocaleString, which may be locale-sensitive or user-defined; use toString for deterministic formatting"}
		}
	}
	if name == "hasOwnProperty" || name == "propertyIsEnumerable" {
		if of == ir.Record && name == "hasOwnProperty" {
			return l.recordObjectCallArguments(node, "hasOwn", append([]*ast.Node{receiver}, written...))
		}
		if of != ir.Object {
			return l.nonObjectOwnPropertyArguments(node, receiver, name, of, written)
		}
		if checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {
			return nil, true, l.notYet(node, name+" on a tuple (its representation includes absent optional slots and no length descriptor)")
		}
		// Every own field of a plain object or a class instance is enumerable. Methods are on its
		// prototype, absent from its shape, and defineProperty is refused.
		lowered, handled, err := l.hasOwnPropertyArguments(node, receiver, written)
		if !missing || !handled || err != nil {
			return lowered, handled, err
		}
		owned, isOwned := lowered.(ir.HasOwn)
		if !isOwned {
			return nil, true, l.notYet(node, name+" on "+regexpGroupsAnswer+" lowered other than as an own-property test")
		}
		return l.regexpGroupsMissingMethod(node, name, []ir.Expression{owned.Object, owned.Key}, func(reads []ir.Expression) ir.Expression {
			return ir.HasOwn{Object: reads[0], Key: reads[1]}
		})
	}
	if len(written) != 0 {
		return nil, true, l.notYet(node, name+" with arguments")
	}
	if checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, true, l.notYet(node, name+" on a tuple (its native representation is not an array)")
	}
	if name == "toLocaleString" {
		toString := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(receiver), "toString")
		if toString == nil || len(toString.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(toString.Declarations[0])) {
			return nil, true, l.notYet(node, "toLocaleString delegating to a user-defined toString")
		}
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if name == "valueOf" {
		if result, _ := l.representation(l.checker.GetTypeAtLocation(node)); result == ir.Union && of == ir.Object {
			if missing {
				return l.regexpGroupsMissingMethod(node, name, []ir.Expression{value}, func(reads []ir.Expression) ir.Expression { return fit(reads[0], result) })
			}
			return fit(value, result), true, nil
		} else if result != of {
			return nil, true, l.notYet(node, "valueOf whose library result type erases the "+typeName(of)+" representation to Object (keeping or returning that result needs a tagged object view)")
		}
		switch of {
		case ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map, ir.Closure:
			if missing {
				return l.regexpGroupsMissingMethod(node, name, []ir.Expression{value}, func(reads []ir.Expression) ir.Expression { return reads[0] })
			}
			return value, true, nil
		}
	}
	switch of {
	case ir.String:
		return value, true, nil
	case ir.Boolean:
		return ir.BooleanToString{Value: value}, true, nil
	case ir.Array:
		return l.arrayMethodArguments(node, receiver, "join", written)
	case ir.Object, ir.Map:
		tag := "[object Object]"
		if of == ir.Map {
			tag = "[object Map]"
			if l.isSet(receiver) {
				tag = "[object Set]"
			}
		}
		text := ir.StringConstant{Index: l.constant(tag)}
		// A receiver that may hold a library object with a tag of its own answers the tag the value has (ir.ObjectTag),
		// never one its type names: a type proves no tag (#eksbz4c, head T).
		answer := func(ir.Expression) ir.Expression { return text }
		if of == ir.Object && l.libraryTagAtRunTime(receiver) {
			answer = func(object ir.Expression) ir.Expression { return ir.ObjectTag{Value: object} }
		}
		if missing {
			return l.regexpGroupsMissingMethod(node, name, []ir.Expression{value}, func(reads []ir.Expression) ir.Expression { return answer(reads[0]) })
		}
		if _, constant := answer(value).(ir.StringConstant); !constant {
			return answer(value), true, nil
		}
		// Both arms are the same, but the test evaluates the receiver exactly once, preserving its
		// side effects and any narrowed-read check. No field is read to produce the constant tag.
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: text, WhenNot: text}, true, nil
	case ir.WeakCollection, ir.Promise:
		// A weak collection and a Promise each have an IR type of their own, so the value is one, and its kind says
		// which: a WeakMap's or a WeakSet's tag, or a Promise's. valueOf reaches here too, and isn't answered.
		if name != "valueOf" {
			return ir.ObjectTag{Value: value}, true, nil
		}
	}
	return nil, true, l.notYet(node, name+" on a "+typeName(of)+" (function source text and mixed value dispatch are not represented)")
}

// regexpGroupsAnswer is libraryObjectOwnAnswer's name for a RegExp match's groups, the one receiver here with no
// prototype, so a direct call of an Object.prototype method on one throws (regexpGroupsMissingMethod).
const regexpGroupsAnswer = "a RegExp match's groups"

// libraryObjectOwnAnswer names a receiver whose toString or toLocaleString answers other than a plain object's on
// Node, where lowering knows no answer of its own, or is empty. The constant tag Object.prototype's arm gives is
// [object Object], and #20e4t06 found it printed for a Map or Set iterator, whose tag is Map Iterator or Set
// Iterator, so the call stops at compile time for any named library interface (an array iterator, a generator, a
// Date, an Intl object) unless it's Object itself or an iterator result, which next makes as a plain object. A Map
// or Set, an array or a string is held as its own kind and keeps its own answer, and a program's own object or class
// keeps Object.prototype's. A library object with a tag of its own that a program can hold, one of
// ir.LibraryObjectTags (a Map's, a Set's, a string's or a matchAll iterator, a WeakMap, a WeakSet, a Promise), isn't
// named: its tag is read from the value at run time (libraryTagAtRunTime, #eksbz4c's head T). A RegExp match's
// groups, an anonymous library type declared inside a RegExp result interface, is named regexpGroupsAnswer only where
// no member stops, since it stops nothing: a groups object has no tag of its own, and a direct call that would find
// no method on one is tested for it at run time (regexpGroupsAdmitted, #eksbz4c's head G). Every member of the
// receiver's type is asked, failing closed.
func (l *lowering) libraryObjectOwnAnswer(receiver *ast.Node) string {
	answer := ""
	for _, member := range l.ownAnswerMembers(receiver) {
		own := l.memberOwnAnswer(member)
		if own != "" && own != regexpGroupsAnswer && !l.memberTagged(member) {
			return own
		}
		if own == regexpGroupsAnswer {
			answer = own
		}
	}
	return answer
}

// libraryObjectTagged reports whether any member of receiver's type is a library object whose tag is its own, one of
// ir.LibraryObjectTags, so a value it holds may answer that tag, or another the value has: a type proves no tag.
func (l *lowering) libraryObjectTagged(receiver *ast.Node) bool {
	for _, member := range l.ownAnswerMembers(receiver) {
		if l.memberTagged(member) {
			return true
		}
	}
	return false
}

// memberTagged reports whether member is a library interface ir.LibraryObjectTags holds, whose values answer a tag of
// their own (adamic_object_tag natively).
func (l *lowering) memberTagged(member *checker.Type) bool {
	if own := l.memberOwnAnswer(member); own == "" || own == regexpGroupsAnswer {
		return false
	}
	return ir.LibraryObjectTag(l.unbranded(member).Symbol().Name) != ""
}

// libraryTagAtRunTime reports whether Object.prototype.toString of receiver, an object, is read from the value at
// run time (ir.ObjectTag) rather than the constant [object Object]: where a member of its type is a library object
// with a tag of its own, or where it's a view some such value in the program relates to (#eksbz4c, head T), as
// v1_interface_map_tostring.a's Stepper holds a Map's iterator. A const made from a plain literal holds nothing
// else, and where the program's values aren't known the value is asked, which is never wrong.
func (l *lowering) libraryTagAtRunTime(receiver *ast.Node) bool {
	if l.libraryObjectTagged(receiver) {
		return true
	}
	if l.exactPlainObject(receiver) {
		return false
	}
	values, known := l.programOwnAnswerValues()
	if !known {
		return true
	}
	view := l.concrete(l.checker.GetTypeAtLocation(receiver))
	for _, value := range values {
		if value.tagged && l.checker.IsTypeAssignableTo(value.shape, view) {
			return true
		}
	}
	return false
}

// regexpGroupsAdmitted reports whether any member of receiver's type is a RegExp match's groups, so a value it holds
// may be a groups object, whose prototype is null, whatever its other members are.
func (l *lowering) regexpGroupsAdmitted(receiver *ast.Node) bool {
	for _, member := range l.ownAnswerMembers(receiver) {
		if l.memberOwnAnswer(member) == regexpGroupsAnswer {
			return true
		}
	}
	return false
}

// ownAnswerMembers are the object types libraryObjectOwnAnswer asks of receiver: the type proven at it, and a
// union's members and an intersection's parts, each asked, so MapIterator<T> & { tag: string } stops as its
// iterator part does (Optimizer's should).
func (l *lowering) ownAnswerMembers(receiver *ast.Node) []*checker.Type {
	proven := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)))
	members := []*checker.Type{proven}
	for index := 0; index < len(members); index++ {
		if members[index].Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			members = append(members, members[index].Types()...)
		}
	}
	return members
}

// memberOwnAnswer is libraryObjectOwnAnswer's name for one member, or empty for a member that answers as a plain
// object does.
func (l *lowering) memberOwnAnswer(member *checker.Type) string {
	if member.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	symbol := l.unbranded(member).Symbol()
	// A host's object is declared in a declaration file, lib, the prelude or @types/node alike (node:buffer's
	// Buffer lives in @types/node, #e0cnaa9), so any declaration file counts, never lib alone.
	if symbol == nil || len(symbol.Declarations) == 0 {
		return ""
	}
	if file := ast.GetSourceFileOfNode(symbol.Declarations[0]); !file.IsDeclarationFile && !load.IsLibrary(file) && !load.IsApple(file) {
		return ""
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindInterfaceDeclaration || declaration.Kind == ast.KindClassDeclaration {
		switch symbol.Name {
		case "Object", "IteratorYieldResult", "IteratorReturnResult", "Map", "ReadonlyMap", "Set", "ReadonlySet", "Array", "ReadonlyArray", "String", "Number", "Boolean":
			return ""
		}
		return "a library " + symbol.Name
	}
	for owner := declaration.Parent; owner != nil; owner = owner.Parent {
		if owner.Kind == ast.KindInterfaceDeclaration && strings.HasPrefix(owner.Name().Text(), "RegExp") {
			return regexpGroupsAnswer
		}
	}
	return ""
}

// ownAnswerValue is a type some expression in the program has that libraryObjectOwnAnswer names, and that name, or
// one that libraryObjectTagged holds, tagged, whose answer is read from the value at run time.
type ownAnswerValue struct {
	shape   *checker.Type
	library string
	tagged  bool
}

// programOwnAnswerValues is each distinct type a call, new, property or element read in the program has that
// libraryObjectOwnAnswer names, a match's groups aside, or that libraryObjectTagged holds, first seen first: every
// library object with an answer of its own that the program can hold, since each is made or reached by one of
// those. It's one walk per lowering, kept as entriesProgramHasRecords keeps its, so a toString through a view asks
// only the relation (#v414yhr). A type is named as the instance first asking it reads it, which a value's library
// type never depends on: a call gives MapIterator<T> as a MapIterator whatever T is. It's false when the program's
// modules aren't known.
func (l *lowering) programOwnAnswerValues() ([]ownAnswerValue, bool) {
	if l.ownAnswerValuesKnown {
		return l.ownAnswerValues, true
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil, false
	}
	l.ownAnswerValuesKnown = true
	seen := map[*checker.Type]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
			shape := l.checker.GetTypeAtLocation(node)
			if !seen[shape] {
				seen[shape] = true
				if library := l.libraryObjectOwnAnswer(node); library != "" && library != regexpGroupsAnswer {
					l.ownAnswerValues = append(l.ownAnswerValues, ownAnswerValue{shape: shape, library: library})
				} else if l.libraryObjectTagged(node) {
					l.ownAnswerValues = append(l.ownAnswerValues, ownAnswerValue{shape: shape, tagged: true})
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return l.ownAnswerValues, true
}

// regexpGroupsMissingMethod is a direct call of an Object.prototype method on a receiver whose type admits a RegExp
// match's groups, made of values, the receiver's first and then the arguments', each evaluated once in JavaScript's
// order. A groups object has no prototype, so Node finds nothing to call and throws a TypeError naming the callee as
// its source reads, "match.groups.toString is not a function", and anything else answers as answer makes of the
// values (#eksbz4c's head G). The test is of the value at run time, never the type (ir.IsRegExpGroups), and the
// TypeError is thrown as global.gc's is (library_node_globals.go), its text computed here, never the JavaScript
// backend's own, which would name the code it emitted. A callee whose text isn't spelled here stops.
func (l *lowering) regexpGroupsMissingMethod(node *ast.Node, name string, values []ir.Expression, answer func(reads []ir.Expression) ir.Expression) (ir.Expression, bool, error) {
	text, spelled := calleeSourceText(ast.SkipParentheses(node.AsCallExpression().Expression))
	if !spelled {
		return nil, true, l.notYet(node, name+" on "+regexpGroupsAnswer+" through a callee whose text Node's TypeError names isn't spelled")
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "regexp_groups_missing_method"}
	reads := []ir.Expression{}
	for _, value := range values {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: value.Type(), Function: index})
		function.Parameters = append(function.Parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	result := answer(reads)
	function.Returns = result.Type()
	failure := ir.MakeError{Message: ir.StringConstant{Index: l.constant(text + " is not a function")}, Name: ir.StringConstant{Index: l.constant("TypeError")}}
	function.Body = []ir.Statement{
		ir.If{Condition: ir.IsRegExpGroups{Value: reads[0]}, Then: []ir.Statement{ir.Throw{Value: failure}}},
		ir.Return{Value: result},
	}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: values, Returns: function.Returns}, true, nil
}

// calleeSourceText is a callee as V8 names it in a TypeError, from the parsed source, since V8 prints its own tree
// and never the text (measured on Node, J-cg4gdgf and #eksbz4c head G's probes): parentheses, a non-null assertion
// and a cast are gone once types are stripped, a property is .name and an optional one ?.name, an element access
// with a string key reads as that property and one with an index as [index], a string literal is written in double
// quotes, an array literal is its elements between brackets, joined by a comma alone, and a call is its callee and
// (...), with arguments or none. Anything else isn't spelled here (false), and its caller stops.
func calleeSourceText(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text(), true
	case ast.KindThisKeyword:
		return "this", true
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return calleeSourceText(node.Expression())
	case ast.KindStringLiteral:
		if strings.ContainsAny(node.Text(), "\"\\\n\r\u2028\u2029") {
			return "", false
		}
		return `"` + node.Text() + `"`, true
	case ast.KindArrayLiteralExpression:
		// A spread or a hole has no case here, so an array literal holding one isn't spelled.
		elements := []string{}
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			text, spelled := calleeSourceText(element)
			if !spelled {
				return "", false
			}
			elements = append(elements, text)
		}
		return "[" + strings.Join(elements, ",") + "]", true
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.QuestionDotToken != nil {
			return "", false
		}
		callee, spelled := calleeSourceText(call.Expression)
		return callee + "(...)", spelled
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access.Name().Kind != ast.KindIdentifier {
			return "", false
		}
		separator := "."
		if access.QuestionDotToken != nil {
			separator = "?."
		}
		object, spelled := calleeSourceText(access.Expression)
		return object + separator + access.Name().Text(), spelled
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access.QuestionDotToken != nil {
			return "", false
		}
		key := ast.SkipParentheses(access.ArgumentExpression)
		object, spelled := calleeSourceText(access.Expression)
		switch {
		case (key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral) && dottedPropertyName(key.Text()):
			return object + "." + key.Text(), spelled
		case key.Kind == ast.KindNumericLiteral && arrayIndexText(key.Text()):
			return object + "[" + key.Text() + "]", spelled
		}
		return "", false
	}
	return "", false
}

// dottedPropertyName reports whether text is a name V8 would print after a dot: ASCII letters, digits, _ and $, not
// led by a digit. Any other key isn't spelled.
func dottedPropertyName(text string) bool {
	for index, character := range text {
		switch {
		case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z', character == '_', character == '$':
		case character >= '0' && character <= '9' && index > 0:
		default:
			return false
		}
	}
	return text != ""
}

// arrayIndexText reports whether a numeric literal is written as V8 prints an index: decimal digits with no
// leading zero, the way [match.groups][0] reads in its TypeError.
func arrayIndexText(text string) bool {
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return text == "0" || (text != "" && text[0] != '0')
}

// Structural object views can hide an override, an Error, or an artificial absent-spread slot.
// Until prototype dispatch and presence descriptors exist, refuse any compatible runtime shape
// in the whole program with one of those hazards. This deliberately errs toward refusal.
// An empty name tests own descriptors without dispatching an overridable receiver method.
func (l *lowering) prototypeHazard(receiver *ast.Node, name string) string {
	return l.prototypeHazardWithErrorTable(receiver, name, false)
}

func (l *lowering) prototypeHazardWithErrorTable(receiver *ast.Node, name string, errorTable bool) string {
	view := l.concrete(l.checker.GetTypeAtLocation(receiver))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return "the program's runtime shapes are not known"
	}
	if !errorTable && name != "valueOf" && l.errorMayHide(receiver) {
		return "an Error may be hidden by the view"
	}
	// A library object whose own answer isn't Object.prototype's and that lowering doesn't read at run time, a Date,
	// a Stats, may be held by a program's own view, where its toString would print [object Object] and Node prints its
	// own answer, a Date's text (#v414yhr). A view naming a Date's method stops at the view today, but this stop
	// doesn't lean on that. Any expression whose type libraryObjectOwnAnswer names
	// may be such a value, so a view it relates to stops. Those types are found once per lowering
	// (programOwnAnswerValues), and only the relation is asked per view. A library object with a tag of its own, a Map
	// or Set iterator, stops no view: its tag is read from the value (libraryTagAtRunTime, #eksbz4c head T). A
	// receiver libraryObjectOwnAnswer names itself has its own stop, a match's groups their own road, and a const made
	// from a plain literal holds nothing else.
	if (name == "toString" || name == "toLocaleString") && l.libraryObjectOwnAnswer(receiver) == "" && !l.exactPlainObject(receiver) {
		values, known := l.programOwnAnswerValues()
		if !known {
			return "the program's runtime shapes are not known"
		}
		for _, value := range values {
			if !value.tagged && l.checker.IsTypeAssignableTo(value.shape, view) {
				return value.library + " may be hidden by the view"
			}
		}
	}
	reason := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if reason != "" {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindRegularExpressionLiteral || (node.Kind == ast.KindCallExpression && l.isLibraryType(l.checker.GetTypeAtLocation(node), "RegExp", "RegExpStringIterator")) || node.Kind == ast.KindNewExpression || node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindArrowFunction || node.Kind == ast.KindNumericLiteral || node.Kind == ast.KindStringLiteral || node.Kind == ast.KindTrueKeyword || node.Kind == ast.KindFalseKeyword {
			shape := l.checker.GetTypeAtLocation(node)
			// An object literal's own type is fresh, and the relation judges a fresh type for excess properties,
			// so { message: "x", hasOwnProperty: f } wouldn't relate to { message: string } and the walk would
			// pass over the very literal that hides the override (#jm2ba55). The view reaches the literal all the
			// same, so it's related by its regular widened type, freshness taken off by the checker's own rule.
			related := shape
			if node.Kind == ast.KindObjectLiteralExpression {
				related = l.checker.GetWidenedType(regularTypeOfObjectLiteral(l.checker, shape))
			}
			if l.checker.IsTypeAssignableTo(related, view) {
				// A matchAll iterator's toString and toLocaleString are Object.prototype's, which answer its own tag, read
				// from the value at run time (libraryTagAtRunTime, #eksbz4c head T), so only its other names stop here.
				tagRead := (name == "toString" || name == "toLocaleString") && l.isLibraryType(shape, "RegExpStringIterator")
				if l.isLibraryType(shape, "RegExp", "RegExpStringIterator") && !l.exactPlainObject(receiver) && !tagRead {
					// Regex objects have intrinsic prototype behavior and metadata slots. A structural
					// view cannot make those plain-object methods or own-property descriptors.
					reason = "a RegExp or its iterator may be hidden by the view"
					return true
				}
				if representation, known := l.representation(shape); known && representation != ir.Object && !l.exactPlainObject(receiver) {
					reason = "a value with a different native representation may be hidden by the view"
					return true
				}
				member := l.checker.GetPropertyOfType(shape, name)
				if name != "" && member != nil && !l.librarySymbol(member) {
					reason = "a user-defined " + name + " may be hidden by the view"
					return true
				}
				if name == "toLocaleString" {
					member = l.checker.GetPropertyOfType(shape, "toString")
					if member != nil && !l.librarySymbol(member) {
						reason = "a user-defined toString may be hidden by the view"
						return true
					}
				}
				if name == "" || name == "hasOwnProperty" || name == "propertyIsEnumerable" {
					for _, field := range l.checker.GetPropertiesOfType(shape) {
						if strings.ContainsRune(field.Name, 0) {
							reason = "a field name contains NUL, which native shape names cannot represent"
							return true
						}
					}
					if node.Kind == ast.KindObjectLiteralExpression {
						for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
							if property.Kind == ast.KindSpreadAssignment && l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
								reason = "a possibly absent spread has slots without own-property presence descriptors"
								return true
							}
						}
					}
					if node.Kind == ast.KindNewExpression {
						if declaration := l.classes[l.symbol(ast.SkipParentheses(node.AsNewExpression().Expression))]; declaration != nil {
							for _, member := range declaration.Members() {
								if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient) {
									reason = "declared class fields have no JavaScript own property"
									return true
								}
								if member.Name() != nil && member.Name().Kind == ast.KindPrivateIdentifier {
									reason = "private class slots are not JavaScript own properties"
									return true
								}
							}
						}
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return reason
}

// errorMayHide reports whether receiver may hold an Error, whose descriptors
// differ from plain object slots. The detached own-key test uses an Error table;
// other descriptor operations still refuse when an Error can reach their view.
// A compatible Error construction can reach the view, as can a runtime Error
// from a host call if the program catches an exception or a Promise rejection.
func (l *lowering) errorMayHide(receiver *ast.Node) bool {
	view := l.concrete(l.checker.GetTypeAtLocation(receiver))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	made, takes := false, false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if made {
			return true
		}
		if node.Kind == ast.KindNewExpression || node.Kind == ast.KindCallExpression {
			shape := l.checker.GetTypeAtLocation(node)
			if l.isLibraryType(shape, errorClasses...) && l.checker.IsTypeAssignableTo(shape, view) {
				made = true
				return true
			}
		}
		takes = takes || takesRejection(node)
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if made {
		return true
	}
	if !takes {
		return false
	}
	errorType := l.libraryErrorType()
	return errorType == nil || l.checker.IsTypeAssignableTo(errorType, view)
}

// takesRejection reports whether node hands the program a thrown value it didn't make: a catch that binds
// it, or a call that can pass a Promise's rejection reason to a handler, read by name alone so a receiver
// typed loosely still counts.
func takesRejection(node *ast.Node) bool {
	if node.Kind == ast.KindCatchClause {
		return node.AsCatchClause().VariableDeclaration != nil
	}
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	switch callee.Name().Text() {
	case "catch", "allSettled":
		return true
	case "then":
		return call.Arguments != nil && len(call.Arguments.Nodes) > 1
	}
	return false
}

// libraryErrorType is the library's Error interface, the type a caught Error the runtime made has, or nil
// if the library doesn't declare one.
func (l *lowering) libraryErrorType() *checker.Type {
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if !load.IsLibrary(file) {
			continue
		}
		for _, statement := range file.Statements.Nodes {
			if statement.Kind == ast.KindInterfaceDeclaration && statement.Name().Text() == "Error" {
				if symbol := l.checker.GetSymbolOfDeclaration(statement); symbol != nil {
					return l.checker.GetDeclaredTypeOfSymbol(symbol)
				}
			}
		}
	}
	return nil
}

// An unannotated const made directly from a plain literal has no boxed primitive hidden by a
// structural view. Its shape is fixed; spreads and annotated/widened names need the whole scan.
func (l *lowering) exactPlainObject(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.checker.GetSymbolAtLocation(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Type != nil || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
		node = ast.SkipParentheses(node)
	}
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind == ast.KindSpreadAssignment {
			return false
		}
	}
	return true
}
