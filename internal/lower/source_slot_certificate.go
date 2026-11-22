package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A write of an object field through a checked view, view.f = value, puts value in a slot other aliases
// of the same object read too, each at the type its own declaration gives f. One that declares f
// narrower than value would read past what the slot holds (TestViewObjectWritesNeedSourceCertificate:
// raw declares child as { ready, other }, the view writes { ready }, and raw.child.other runs past the
// new object). The source-slot certificate is Adamic's Rule C on #2msw6t8, asked of the reads that exist:
//
//  1. A plain write `x.f = value;` through a checked view is admitted, and the type it writes is
//     recorded for the name f (recordSourceSlotWrite). Any other write keeps view's stop, and so does
//     one to toString or valueOf, which ToPrimitive reads by name after the reads are judged. One to a
//     field that may hold undefined or null is NotYet, since the checked write can't change its slot's
//     representation, wherever its statement is lowered (checkSourceSlotPending). A field held other than as
//     an object keeps the plain checked write, and a write that can store undefined or null into one whose
//     type holds either is NotYet too (sourceSlotScalarWrite), held the same way.
//  2. What the slot can hold, W, is every type a view wrote to f and every type a declaration of f
//     gives it, undefined included where one is optional (sourceSlotDeclared): an absent slot or one an
//     initializer filled holds a value no view write records (Compiler's condition, binding).
//  3. Once lowering is done, every read of f is held to W (certifySourceSlots, sourceSlotHeld). A read
//     whose declared type holds all of W stays unchecked and costs nothing. Any other read is checked
//     under its receiver's key (CheckedFields), and the type it reads gets its view contract, so the
//     fields inside it are checked where they're read too, and a lie stops loudly at the read.
//
// Every road that reads a field by its name is a read site with its own declared type, or it keeps a
// stop (internal/oracle's TestSourceSlotCertificate pins one program per road):
//   - a property access, and a destructured field in a declaration, a parameter, a for-of, a nested
//     pattern or an assignment pattern: each an ir.Property that carries the field's declared type
//     (readObjectField, destructuredField's readers), held to W here; a catch binding is unknown or
//     any, so its destructuring needs a cast or a thrown object lowering stops first;
//   - a const computed key: view refuses a viewed name read by a literal key (refuseViewElementReads),
//     and finiteElementRead's ir.Property carries no declared type, so it keeps the write's stop here;
//   - a spread or an Object.assign copies the slot into a new object, whose reads are sites of their
//     own, and checkViewCopies stops a copy of a receiver type a check was added for; structuredClone
//     isn't declared by the library Adamic checks against;
//   - JSON.stringify and encodeJson read every field by the declared type's schema, so one whose type
//     holds f at a type that doesn't hold W keeps the write's stop (sourceSlotWholeReads);
//   - Object.values and Object.entries are admitted on a fixed object only for scalar values
//     (library_object.go), and a for-in yields keys, so neither reads an object field's value;
//   - Reflect.get, and a read of an unknown by name, give unknown, a value no type claims, which a
//     later narrowing tests as it runs;
//   - an accessor dispatcher's fallback is an unchecked load, and finishAccessors stops a checked read
//     that would go through it, so this runs before it;
//   - a read through a type parameter has no one declared type: the type its instantiation reads isn't
//     the one lowering records, so it keeps the write's stop (sourceSlotGenericReads), and so does a read
//     whose declared type is generic.
type sourceSlotWrite struct {
	target  *ast.Node
	name    string
	written *checker.Type
	stop    error
}

// sourceSlot is what a slot named by a written name can hold: the writes views made to it, and the type
// each declaration of the name gives it (sourceSlotDeclared).
type sourceSlot struct {
	writes   []sourceSlotWrite
	declared []*checker.Type
}

// sourceSlotPendingStop is a write to a checked field that may hold undefined or null, target its left side,
// statement the expression statement it's the whole of, and stop Rule C's NotYet there, made where the view
// walk met it (recordSourceSlotWrite, sourceSlotScalarWrite).
type sourceSlotPendingStop struct {
	target    *ast.Node
	statement *ast.Node
	stop      error
}

// sourceSlotNamedKinds are the IR kinds other than ir.Property that name a field, each judged without a
// check: none reads a fixed object's slot at a type's claim. DynamicProperty reads by the allocation's
// own runtime shape and gives a tagged union, as DynamicKeyedProperty and HasOwn do by a key; HasProperty,
// HasAccessor and OwnFunction test presence, or whether a function is there, reading no value;
// PhantomMember reads a primitive, and the regexp kinds a regexp's result; SetProperty is a write, judged
// at view; Labeled names a label. The kinds that read an object without naming a field are judged
// elsewhere: CheckedCast tests a tag at run time, ObjectCall's values and entries are checked or scalar
// only and its assign is a copy (checkViewCopies), JSONStringify and JSONEncode are
// sourceSlotWholeReads', and the record and map kinds read records and maps, which no fixed object is.
var sourceSlotNamedKinds = map[reflect.Type]bool{
	reflect.TypeOf(ir.DynamicProperty{}): true,
	reflect.TypeOf(ir.HasProperty{}):     true,
	reflect.TypeOf(ir.HasAccessor{}):     true,
	reflect.TypeOf(ir.OwnFunction{}):     true,
	reflect.TypeOf(ir.PhantomMember{}):   true,
	reflect.TypeOf(ir.RegExpProperty{}):  true,
	reflect.TypeOf(ir.RegExpGroup{}):     true,
	reflect.TypeOf(ir.SetProperty{}):     true,
	reflect.TypeOf(ir.Labeled{}):         true,
}

// recordSourceSlotWrite records target, an object field write through a checked view whose stop is
// stop, as written for certifySourceSlots, and gives back nil where it was, or the stop to give. Only the
// whole left side of a plain `=` statement is recorded (hostMemberAssignment): a compound, a logical or a
// destructuring write keeps its stop, and so does one found after certifySourceSlots has judged the
// reads. A field whose type holds undefined or null, an optional one among them, is NotYet: its slot can
// hold undefined, null or an object, and the checked write (object.c's adamic_object_view_write, and the
// JavaScript backend's adamicViewWrite) stops a write that changes the representation the slot holds, so
// a program Node runs would stop at it. That NotYet is held, not given, since the view walk judges every
// write in the program from types, lowered or not, and only a write lowering makes changes a slot
// (checkSourceSlotPending, Adamic's ruling on #rgrtt32). A fresh object or array literal is made as the slot
// it's written into (freshOrWidened), so it's recorded as the field's own type.
func (l *lowering) recordSourceSlotWrite(target *ast.Node, field *ast.Symbol, stop error) error {
	if l.sourceSlotsJudged {
		return stop
	}
	// finishToPrimitive reads an object's own toString and valueOf by name (plainToString) after
	// certifySourceSlots has judged every read, so no read of a slot by either name it makes would be held
	// to what a view wrote there: a write to one keeps its stop. Today String() of an object whose
	// toString or valueOf is an object is NotYet (ToPrimitive isn't lowered for it), so this guards the
	// day it is (Loom job J-shhzr5n).
	if name := target.Name().Text(); name == "toString" || name == "valueOf" {
		return stop
	}
	assignment := hostMemberAssignment(target)
	if assignment == nil {
		return stop
	}
	declared := l.checker.GetTypeOfSymbol(field)
	if field.Flags&ast.SymbolFlagsOptional != 0 || l.includesUndefined(declared) || l.includesNull(declared) {
		return l.holdSourceSlotStop(target, assignment.Parent, l.notYet(target, "writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds"))
	}
	for _, write := range l.sourceSlotWrites {
		if write.target == target {
			return nil
		}
	}
	value := assignment.AsBinaryExpression().Right
	written := l.checker.GetTypeAtLocation(value)
	switch ast.SkipParentheses(value).Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		written = declared
	}
	l.sourceSlotWrites = append(l.sourceSlotWrites, sourceSlotWrite{target: target, name: target.Name().Text(), written: written, stop: stop})
	return nil
}

// sourceSlotNullishWrite is the words of Rule C's scalar arm (sourceSlotScalarWrite).
const sourceSlotNullishWrite = "writing undefined or null into a checked field, whose checked write can't change the representation its slot holds"

// sourceSlotScalarWrite is Rule C's arm for a field held other than as an object, a write to which takes the
// plain checked write rather than recordSourceSlotWrite (#f95gkw2). A write that can store undefined or null
// into a field whose type holds either is NotYet, for two reasons that are one hazard. The checked write
// (object.c's adamic_object_view_write, the JavaScript backend's adamicViewWrite) stops a write that changes
// the representation the slot holds, so `n.note = undefined` into a slot a literal filled with a string, or
// `n.count = undefined` into a plain number's, stops a program Node runs. And a value that may be undefined
// but isn't the literal passes that check as the slot's own representation, so another alias that declares
// the field without undefined then reads a null string where it expects one, which native code follows into
// the null (a segfault where Node throws). What the slot holds is fixed by every write before this one, so no
// compile-time proof of it is asked; any written value whose type may hold undefined or null stops, failing
// closed on any, unknown and an open type parameter (mayHoldNull). A value that can't be either writes as
// before, and so does a write to a field whose type holds neither. Of a compound write only `??=`, `||=` and
// `&&=` can store their right side, and arithmetic or `++`/`--` stores a number or a string; a write whose
// value this can't name (a destructuring or a for-in or for-of target) stops where the field may hold either.
func (l *lowering) sourceSlotScalarWrite(target *ast.Node, field *ast.Symbol) error {
	// The indexed walk asks every keyed access (viewSiteAccesses), a read among them, and a read writes nothing.
	if !ast.IsAssignmentTarget(target) {
		return nil
	}
	declared := l.checker.GetTypeOfSymbol(field)
	if field.Flags&ast.SymbolFlagsOptional == 0 && !l.includesUndefined(declared) && !l.includesNull(declared) {
		return nil
	}
	written, named := l.sourceSlotWrittenType(target)
	if !named || l.includesUndefined(written) || l.mayHoldNull(written) {
		return l.holdSourceSlotStop(target, sourceSlotWriteStatement(target), l.notYet(target, sourceSlotNullishWrite))
	}
	return nil
}

// holdSourceSlotStop holds stop, a Rule C NotYet at target, for checkSourceSlotPending, keyed to statement, the
// statement whose lowering lowers the write, and gives back nil. A write with no such statement
// (sourceSlotWriteStatement's nil) has none to wait for, so its stop is given back at once, as the view walk gave
// it before #5qx6qv4. A target already held keeps its first stop.
func (l *lowering) holdSourceSlotStop(target, statement *ast.Node, stop error) error {
	if statement == nil {
		return stop
	}
	for _, pending := range l.sourceSlotPending {
		if pending.target == target {
			return nil
		}
	}
	l.sourceSlotPending = append(l.sourceSlotPending, sourceSlotPendingStop{target: target, statement: statement, stop: stop})
	return nil
}

// sourceSlotWriteStatement is the statement nearest target that statement lowers, the one whose lowering lowers
// the write: walking out from target, through any expression and through an arrow or a function expression,
// whose body is lowered only where the expression holding it is, to the first statement of a kind statement
// handles (sourceSlotStatementKinds). It's nil, so the stop is given at once, where the walk meets a node whose
// expressions lowering reaches by another road than its statement: a parameter's default, a class member, a
// decorator, an enum member, a JSX attribute or a declaration that is lowered by itself (#5qx6qv4, on the 322
// tsc units #rgrtt32's census stopped at utilities.ts:9006, `void (file.externalModuleIndicator = ...)` in an
// arrow's expression body).
func sourceSlotWriteStatement(target *ast.Node) *ast.Node {
	for node := target.Parent; node != nil; node = node.Parent {
		if sourceSlotStatementKinds[node.Kind] {
			return node
		}
		switch node.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			continue
		case ast.KindParameter, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindMethodDeclaration,
			ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindClassStaticBlockDeclaration,
			ast.KindDecorator, ast.KindEnumMember, ast.KindJsxAttribute, ast.KindFunctionDeclaration,
			ast.KindClassDeclaration, ast.KindClassExpression, ast.KindSourceFile, ast.KindModuleBlock, ast.KindBlock:
			return nil
		}
	}
	return nil
}

// sourceSlotStatementKinds are the statements statement lowers by itself, each marked in loweredStatements as it
// begins (statements.go). A Block is left out: a write is always inside a statement the Block holds, and a
// function's body Block is lowered by its statements, not by statement.
var sourceSlotStatementKinds = map[ast.Kind]bool{
	ast.KindExpressionStatement: true,
	ast.KindVariableStatement:   true,
	ast.KindReturnStatement:     true,
	ast.KindIfStatement:         true,
	ast.KindForStatement:        true,
	ast.KindWhileStatement:      true,
	ast.KindDoStatement:         true,
	ast.KindForInStatement:      true,
	ast.KindForOfStatement:      true,
	ast.KindSwitchStatement:     true,
	ast.KindLabeledStatement:    true,
	ast.KindThrowStatement:      true,
	ast.KindTryStatement:        true,
}

// checkSourceSlotPending gives back the first held write's NotYet, in the view walk's order, whose statement
// lowering lowered, once every pass that lowers a body has run: only IR the program emits can change a slot,
// so a write whose statement was never lowered guards nothing (Adamic's ruling on #rgrtt32, 163 tsc units at
// setValueDeclaration). A statement is marked where statement begins to lower it, before any road it takes,
// so every form a plain write lowers to keeps the stop, a road that records no write site (a Node member
// require makes doomed, a namespace member's assignment, an Apple property's setter, super's setter)
// included, and so does a lowering that stopped or was thrown away. A write whose statement can't be found
// keeps its stop too. Rule C's scalar arm (sourceSlotScalarWrite) holds its stop the same way.
func (l *lowering) checkSourceSlotPending() error {
	for _, pending := range l.sourceSlotPending {
		if pending.statement == nil || l.loweredStatements[pending.statement] {
			return pending.stop
		}
	}
	return nil
}

// sourceSlotTargetWrapper says whether an expression of kind holds an assignment target as the same slot.
func sourceSlotTargetWrapper(kind ast.Kind) bool {
	switch kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return true
	}
	return false
}

// sourceSlotWrittenType is the type of the value an assignment to target stores, and whether it could be named.
// An assignment target may sit inside parentheses, a non-null `!` (`n.count!++`), an `as`, a `satisfies` or
// an angle assertion, each of which leaves the target the same slot.
func (l *lowering) sourceSlotWrittenType(target *ast.Node) (*checker.Type, bool) {
	held := target
	for held.Parent != nil && sourceSlotTargetWrapper(held.Parent.Kind) {
		held = held.Parent
	}
	switch parent := held.Parent; {
	case parent == nil:
		return nil, false
	case parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == held:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return l.checker.GetTypeAtLocation(binary.Right), true
		}
		if ast.IsCompoundAssignment(binary.OperatorToken.Kind) {
			return l.checker.GetTypeAtLocation(parent), true
		}
	case parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression:
		return l.checker.GetNumberType(), true
	}
	return nil, false
}

// certifySourceSlots holds every read of a name recordSourceSlotWrite recorded to what its slot can hold,
// once lowering is done: a read whose declared type holds all of it stays as it is, and any other is
// checked under its receiver's key, its declared type given a view contract. A read lowering can't name
// a declared type for, or one that has no receiver key to check it under, gives back the first write's
// stop, view's own, unchanged.
func (l *lowering) certifySourceSlots() error {
	l.sourceSlotsJudged = true
	if len(l.sourceSlotWrites) == 0 {
		return nil
	}
	slots := map[string]*sourceSlot{}
	for _, write := range l.sourceSlotWrites {
		if slots[write.name] == nil {
			slots[write.name] = &sourceSlot{}
		}
		slots[write.name].writes = append(slots[write.name].writes, write)
	}
	if err := l.sourceSlotDeclared(slots); err != nil {
		return err
	}
	if err := l.sourceSlotGenericReads(slots); err != nil {
		return err
	}
	if err := l.sourceSlotWholeReads(slots); err != nil {
		return err
	}
	// Each declared type a checked read reads, in the order first met, with its slot's first write.
	var declaredTypes []*checker.Type
	failing := map[*checker.Type]sourceSlotWrite{}
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		read, isProperty := node.(ir.Property)
		if !isProperty {
			// Any other kind that names a written field is judged by sourceSlotNamedKinds, and one it
			// doesn't list keeps the stop: ir.BoundMethod, which resolves the slot as a closure with no
			// check of what it holds, and any kind added after this.
			if named := reflect.ValueOf(node).FieldByName("Name"); named.IsValid() && named.Kind() == reflect.String && slots[named.String()] != nil && !sourceSlotNamedKinds[reflect.TypeOf(node)] {
				refused = slots[named.String()].writes[0].stop
				return false
			}
			return true
		}
		if slots[read.Name] == nil {
			return true
		}
		slot := slots[read.Name]
		if read.Method {
			// A callee's read takes the object's own function value, or its class's method, and a
			// checked read of one isn't emitted the way a field's is (readiness keeps its view only for
			// a certified callable), so a written name read as a method keeps the stop.
			refused = slot.writes[0].stop
			return false
		}
		if read.ViewTypeID == 0 {
			// A read lowering made as a check of its own (a predicate's tag test, a speculation's field
			// check) carries its view. One with neither names no type to hold the slot to.
			if read.View == "" {
				refused = slot.writes[0].stop
			}
			return true
		}
		declared := l.viewTypes[read.ViewTypeID]
		if declared == nil || declared.Flags()&checker.TypeFlagsInstantiable != 0 {
			refused = slot.writes[0].stop
			return false
		}
		if l.sourceSlotHeld(declared, slot) {
			return true
		}
		if read.ViewReceiverTypeID == 0 {
			refused = slot.writes[0].stop
			return false
		}
		// Through markCheckedField, as every key lowering adds is, so a view made after this asks the
		// accesses of its name again (viewWalkIndexed).
		l.markCheckedField(ir.CheckedFieldKey(read.ViewReceiverTypeID, read.Name), read.Name, false)
		if _, seen := failing[declared]; !seen {
			failing[declared] = slot.writes[0]
			declaredTypes = append(declaredTypes, declared)
		}
		return true
	}
	// Every body lowering has made is in Main or Functions here: a class names its constructor,
	// methods and accessors by function index, and its field values are its layout's, walked too. The
	// passes after this one move statements (exceptions, NormalizeAsync) or clone a read judged here
	// (finishAccessors' fallbacks), and a checked read stays checked wherever it goes, since its key is
	// in CheckedFields.
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	for _, class := range l.result.Classes {
		walk(class.Fields, inspect)
		walk(class.PublicFields, inspect)
	}
	if refused != nil {
		return refused
	}
	// The read's own check is its field's presence and representation. What's inside it is read again
	// through the declared type, so that type's fields are checked under it too (viewSchema).
	for _, declared := range declaredTypes {
		if _, err := l.viewSchema(failing[declared].target, nil, declared); err != nil {
			return failing[declared].stop
		}
	}
	return nil
}

// sourceSlotHeld is the certificate's one question: whether a read of slot at type read holds every
// value the slot can, each type a view wrote to it and each type a declaration gives it. A value of type
// held is held where it's assignable to read, as the holder check relates a value to a slot
// (hostMemberHolder), and seeing it as read widens no mutable slot inside it (widened).
func (l *lowering) sourceSlotHeld(read *checker.Type, slot *sourceSlot) bool {
	holds := func(held *checker.Type) bool {
		return l.enumAssignable(held, read) && l.checker.IsTypeAssignableTo(held, read) && l.widened(held, read, map[[2]*checker.Type]bool{}) == nil
	}
	for _, write := range slot.writes {
		if !holds(write.written) {
			return false
		}
	}
	for _, declared := range slot.declared {
		if !holds(declared) {
			return false
		}
	}
	return true
}

// sourceSlotDeclared adds to each slot the type every declaration of its name gives it, in every file
// the checker sees: the program's modules, the declaration files it takes types from without importing
// a value (which moduleOrder doesn't list), and the libraries. A library declaration counts because a
// program object can be typed by a library interface whose optional member is the only declaration that
// says its slot may be absent (const descriptor: PropertyDescriptor = {}), and an object a library call
// makes holds what its declarations say.
//
// A declaration is an interface or type literal member, a class field (declare and parameter properties
// included), an object literal's property, shorthand or method, and a get or set accessor, which counts
// as the slot holding its type though an accessor holds no slot: counting it only makes more reads
// checked. An optional one adds undefined, and so does a class field nothing initializes, declared or
// definitely assigned with !, whose slot is absent until a store fills it. A computed name counts by its
// literal key. One whose key is a string the checker can't name could declare any name, so it gives back
// the first write's stop, view's own. An index signature isn't counted: a fixed object is never seen as
// a record (records.go's recordStorageView, #mez92e7), so none reads a fixed object's slot.
func (l *lowering) sourceSlotDeclared(slots map[string]*sourceSlot) error {
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		switch part.Kind {
		case ast.KindPropertySignature, ast.KindPropertyDeclaration, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodSignature, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindParameter:
			if part.Kind == ast.KindParameter && !parameterProperty(part) {
				break
			}
			name, named := l.sourceSlotDeclaredName(part)
			if !named {
				for _, slot := range slots {
					refused = slot.writes[0].stop
					return true
				}
			}
			slot := slots[name]
			if slot == nil {
				break
			}
			symbol := l.checker.GetSymbolOfDeclaration(part)
			if symbol == nil {
				refused = slot.writes[0].stop
				return true
			}
			slot.declared = append(slot.declared, l.checker.GetTypeOfSymbol(symbol))
			absent := symbol.Flags&ast.SymbolFlagsOptional != 0
			if part.Kind == ast.KindPropertyDeclaration && part.AsPropertyDeclaration().Initializer == nil && (ast.HasSyntacticModifier(part, ast.ModifierFlagsAmbient) || part.PostfixToken() != nil && part.PostfixToken().Kind == ast.KindExclamationToken) {
				absent = true
			}
			if absent {
				slot.declared = append(slot.declared, l.checker.GetUndefinedType())
			}
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		file.AsNode().ForEachChild(visit)
		if refused != nil {
			return refused
		}
	}
	return nil
}

// sourceSlotDeclaredName is the name a declaration gives a field, as a written name is spelled, and
// whether it could be named: "" for a private name, a symbol key or a numeric one, which no x.f write
// names, since f is always an identifier there (hostMemberAssignment). A computed key the checker gives a
// string literal type is that string, and one whose type is any other string isn't named.
func (l *lowering) sourceSlotDeclaredName(declaration *ast.Node) (string, bool) {
	name := declaration.Name()
	if name == nil {
		return "", true
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return name.Text(), true
	case ast.KindComputedPropertyName:
		key := l.checker.GetTypeAtLocation(name.AsComputedPropertyName().Expression)
		if key.Flags()&checker.TypeFlagsStringLiteral != 0 {
			text, isText := key.AsLiteralType().Value().(string)
			return text, isText
		}
		if key.Flags()&(checker.TypeFlagsESSymbolLike|checker.TypeFlagsNumberLike) != 0 {
			return "", true
		}
		return "", false
	}
	return "", true
}

// sourceSlotGenericReads gives back a write's stop where the program reads its name through a generic
// receiver: a type parameter other than this, or an indexed access, conditional or substitution type.
// Its read is lowered with the declared type of the constraint's field, while the instantiation it's
// called with may read the field narrower than that, as T["f"] for a T of Required<...>. A read by a
// property access, a literal key or an object binding pattern is weighed, and a plain write isn't.
func (l *lowering) sourceSlotGenericReads(slots map[string]*sourceSlot) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		var name string
		var receiver *checker.Type
		switch part.Kind {
		case ast.KindPropertyAccessExpression:
			if hostMemberAssignment(part) == nil && ast.IsIdentifier(part.Name()) {
				name = part.Name().Text()
				receiver = l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression)
			}
		case ast.KindElementAccessExpression:
			access := part.AsElementAccessExpression()
			if key := l.checker.GetTypeAtLocation(access.ArgumentExpression); key.Flags()&checker.TypeFlagsStringLiteral != 0 {
				name, _ = key.AsLiteralType().Value().(string)
				receiver = l.checker.GetTypeAtLocation(access.Expression)
			}
		case ast.KindBindingElement:
			binding := part.AsBindingElement()
			if part.Parent != nil && part.Parent.Kind == ast.KindObjectBindingPattern {
				key := binding.PropertyName
				if key == nil {
					key = part.Name()
				}
				if key != nil && (ast.IsIdentifier(key) || key.Kind == ast.KindStringLiteral) {
					name = key.Text()
					receiver = l.checker.GetTypeAtLocation(part.Parent)
				}
			}
		}
		if slot := slots[name]; slot != nil && receiver != nil && l.sourceSlotGeneric(receiver) {
			refused = slot.writes[0].stop
			return true
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return refused
}

// sourceSlotGeneric is whether a receiver, or a member of it, is a type parameter other than this, or an
// indexed access, conditional or substitution type.
func (l *lowering) sourceSlotGeneric(receiver *checker.Type) bool {
	receiver = l.checker.GetNonNullableType(receiver)
	members := []*checker.Type{receiver}
	if receiver.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		members = receiver.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsTypeParameter != 0 && !member.AsTypeParameter().IsThisType() {
			return true
		}
		if member.Flags()&(checker.TypeFlagsIndexedAccess|checker.TypeFlagsConditional|checker.TypeFlagsSubstitution) != 0 {
			return true
		}
	}
	return false
}

// sourceSlotWholeReads gives back a write's stop where a reader that reads every field by its value's
// declared type, JSON.stringify's schema or encodeJson's (wholeObjectCopies' readers), reads a value
// whose type holds the name, at any depth, at a type that doesn't hold what the slot can.
func (l *lowering) sourceSlotWholeReads(slots map[string]*sourceSlot) error {
	seen := map[*checker.Type]bool{}
	for _, copied := range l.wholeObjectCopies {
		if copied.reader == "" {
			// A copy, whose fields are read again at sites of their own.
			continue
		}
		if write := l.sourceSlotHeldNarrower(copied.copied, slots, seen); write != nil {
			return write.stop
		}
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		if part.Kind == ast.KindCallExpression {
			call := part.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "stringify" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") && len(call.Arguments.Nodes) != 0 {
				if write := l.sourceSlotHeldNarrower(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), slots, seen); write != nil {
					refused = write.stop
					return true
				}
			}
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return refused
}

// sourceSlotHeldNarrower is the first write to a slot whose name holder, or anything it holds through
// its fields, elements and index signatures, declares at a type that doesn't hold what the slot can, or
// nil. A generic holder can't be read, so it's the first write recorded. Seen holds the types already
// weighed, each answered nil the second time, since the first answered for it.
func (l *lowering) sourceSlotHeldNarrower(holder *checker.Type, slots map[string]*sourceSlot, seen map[*checker.Type]bool) *sourceSlotWrite {
	holder = l.checker.GetNonNullableType(l.concrete(holder))
	if seen[holder] {
		return nil
	}
	seen[holder] = true
	if holder.Flags()&checker.TypeFlagsInstantiable != 0 {
		return &l.sourceSlotWrites[0]
	}
	if holder.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range holder.Types() {
			if write := l.sourceSlotHeldNarrower(member, slots, seen); write != nil {
				return write
			}
		}
		return nil
	}
	if holder.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(holder) {
		declared := l.checker.GetTypeOfSymbol(property)
		if slot := slots[property.Name]; slot != nil && !l.sourceSlotHeld(declared, slot) {
			return &slot.writes[0]
		}
		if write := l.sourceSlotHeldNarrower(declared, slots, seen); write != nil {
			return write
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(holder) {
		if write := l.sourceSlotHeldNarrower(info.ValueType(), slots, seen); write != nil {
			return write
		}
	}
	return nil
}
