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
//     representation.
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
// a program Node runs would stop at it. A fresh object or array literal is made as the slot it's written
// into (freshOrWidened), so it's recorded as the field's own type.
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
		return l.notYet(target, "writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds")
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
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
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
		l.result.CheckedFields[ir.CheckedFieldKey(read.ViewReceiverTypeID, read.Name)] = true
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
		if _, err := l.viewSchema(failing[declared].target, declared); err != nil {
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
