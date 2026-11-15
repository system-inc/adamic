package lower

import (
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// Loose slots (#kaz4pk8). A slot typed unknown or object holds whatever enters it, and its type names
// nothing the cycle finder follows: reachLinks gives such a type no links (cycles.go), so a cycle through
// one was never a seed, never a graph region, and leaked (g13: a.self = a with self: unknown freed 0 of
// 1). Until the finder links these slots (the lift Runtime ruled), a cycle through one stops here, fail
// closed. It's judged in every program, since one whose only cycle runs through a loose slot forms no
// seed and so is never a graph program.
//
// A slot is loose when its type holds unknown or object (holdsLoose): itself, through a union, or
// through what a container holds, an array's, a tuple's, a Map's, a Set's or any reference's type
// arguments, a record's values and an index signature's.
//
// H is what can sit in a loose slot: the type of every value that enters one, read as the checker has it
// where it enters, after instantiation where lowering reads it. Where H is gathered, so a skipped site can
// be found: looseEntry, called on every node of every module from findCycles' walk, reads relationSite's
// sites (void_widening.go), a variable's, a parameter default's and a class field's initializer, an
// assignment (=, ??=, ||= and &&=: a field's, an element's, a record entry's and a variable's write), a
// return and an arrow's body, an argument by the resolved signature's parameter (push, unshift, splice,
// fill, Set's add, Map's set, a parameter property's constructor argument, and any argument bound to a
// parameter typed unknown or object), an object literal's field, an array literal's element, an
// assertion and the implied target; beside them a shorthand field ({ node }), a spread into an array
// literal or an argument list, an object spread, a throw's value (a catch binding typed unknown receives
// it), and a for-of or for-in over an existing variable. A value written where nothing names a slot, an
// untyped literal's field among them, keeps its own type there. noteLooseInit adds each literal field,
// element and entry lowering makes into a loose slot, read in the instantiation being lowered. A value
// that is itself unknown, object, any or a type parameter makes H every shape (looseSlots.every): its type says
// nothing of what it holds. So does a captured variable typed loose whose values don't all arrive by a
// site above: a parameter, a catch binding, a destructured or a loop binding (looseCellEvery).
//
// Where an any can still reach a store: a library declaration's any (JSON.parse's result, an any callback
// parameter, Function.prototype.call's arguments), and any any an adaptation left in tsc's sources. An
// any value written anywhere makes H every shape, but a slot typed any isn't judged here, and reaching
// still takes any for a template that holds nothing (cycles.go's deriveTemplate).
//
// The stop (looseSlotStop) asks the finder's own question with H linked in: on any path, unknown and
// object link to every type in H (looseLinks), and reaching is otherwise reachLinks', shape candidates and
// all, so a value seen through an upcast where it enters H still brings what it really is. Each value
// entering a loose slot is judged where it enters, by its type: a write (looseWrites, every IR node that
// writes a slot, by its Site), a literal's initialization (looseInits) and a captured variable's
// declaration or assignment. It stops when the value can reach the slot's holder back, or the cell
// itself. No write is excused as fresh: a value entering a loose slot is what H holds, and a cycle it
// closes later at a declared slot is one the finder can't see.
//
// Why a declared slot needs no judgment of its own: a cycle that crosses a loose slot crosses it at some
// value that entered it, and that entry was judged. The entered value's type reaches, through declared
// fields, shape candidates and every other loose slot's H, everything the value can come to hold
// afterwards, including the declared slot that closes the cycle, since that slot's type is on the path. So
// the entry already stopped, however late the declared write that closes the cycle runs (k18). The declared
// arm (looseDeclaredStop) judges such slots anyway for the first measured run, so the census can split what
// each arm walls; when the entry arm holds alone, the declared arm and its search per slot go.

// looseStopWords end every loose-slot stop, so its pins can name it.
const looseStopWords = "a cycle the cycle finder can't follow through unknown or object"

// looseSlots is H and what the stop keeps while it judges.
type looseSlots struct {
	// entries are H's types, each once, and every is whether H is every shape.
	entries []*checker.Type
	entered map[*checker.Type]bool
	every   bool
	// sources are the relation sites' values by the node lowering names a write's site at (looseReceiver)
	// or a variable's declared name (looseVariable).
	sources map[*ast.Node][]*checker.Type
	// targets caches holdsLoose.
	targets map[*checker.Type]bool
	// edges caches reachLinks for the stop's own search, valid while shapes holds edgeShapes types.
	edges      map[cycleNode][]cycleNode
	edgeShapes int
	// everything is every type a value can be, valid while shapes holds everythingShapes types.
	everything       []cycleNode
	everythingShapes int
	// touching is looseTouches' answer for every node the program's types reach.
	touching map[cycleNode]bool
}

// looseInit is a literal's field, element or entry lowering made: the literal's type (holder), the
// slot's, the value's, where it is, and the field's name, empty for an element or an entry, "..." for a
// spread and "[]" for a computed name.
type looseInit struct {
	holder *checker.Type
	slot   *checker.Type
	value  *checker.Type
	node   *ast.Node
	field  string
}

// looseEntry gathers H at one node (the comment above names the sites).
func (f *cycleFinder) looseEntry(node *ast.Node) {
	l := f.l
	switch node.Kind {
	case ast.KindShorthandPropertyAssignment:
		symbol := l.checker.GetShorthandAssignmentValueSymbol(node)
		if symbol == nil || node.Name() == nil {
			return
		}
		source := l.checker.GetTypeOfSymbol(symbol)
		object := l.slotType(node.Parent)
		if object == nil {
			object = l.checker.GetTypeAtLocation(node.Parent)
		}
		target := l.checker.GetTypeOfPropertyOfType(object, node.Name().Text())
		if target == nil {
			target = source
		}
		f.looseEnter(source, target)
		return
	case ast.KindSpreadElement:
		source := l.checker.GetTypeAtLocation(node.AsSpreadElement().Expression)
		parent := node.Parent
		switch parent.Kind {
		case ast.KindArrayLiteralExpression:
			target := l.slotType(parent)
			if target == nil {
				target = l.checker.GetTypeAtLocation(parent)
			}
			f.looseEnter(source, target)
		case ast.KindCallExpression, ast.KindNewExpression:
			// Which parameters a spread fills isn't known here, so it enters if any of them is loose.
			if signature := l.checker.GetResolvedSignature(parent); signature != nil {
				for _, parameter := range signature.Parameters() {
					if f.holdsLoose(l.checker.GetTypeOfSymbol(parameter)) {
						f.looseHold(source)
						f.looseKeep(looseReceiver(node), source)
						break
					}
				}
			}
		}
		return
	case ast.KindSpreadAssignment:
		object := l.slotType(node.Parent)
		if object == nil {
			object = l.checker.GetTypeAtLocation(node.Parent)
		}
		if f.looseFields(object) {
			f.looseHold(l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression))
		}
		return
	case ast.KindThrowStatement:
		if thrown := node.AsThrowStatement().Expression; thrown != nil {
			f.looseHold(l.checker.GetTypeAtLocation(thrown))
		}
		return
	case ast.KindForOfStatement, ast.KindForInStatement:
		loop := node.AsForInOrOfStatement()
		if loop.Initializer != nil && loop.Initializer.Kind != ast.KindVariableDeclarationList {
			f.looseEnter(l.checker.GetTypeAtLocation(loop.Expression), l.checker.GetTypeAtLocation(loop.Initializer))
		}
		return
	case ast.KindParameter:
		// A parameter property's store is the parameter's value, written at the class's name
		// (parameter_properties.go); its arguments entered H at their calls.
		if parameterProperty(node) && node.Parent != nil && node.Parent.Parent != nil && node.Parent.Parent.Name() != nil {
			if declared := l.checker.GetTypeAtLocation(node.Name()); f.holdsLoose(declared) {
				f.looseKeep(node.Parent.Parent.Name(), declared)
			}
		}
		return
	}
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
		return
	}
	_, source, target := l.relationSite(node)
	if source == nil {
		return
	}
	if target == nil {
		if !viewSite(node) {
			return
		}
		target = source
	}
	if !f.looseEnter(source, target) {
		return
	}
	f.looseKeep(looseReceiver(node), source)
	f.looseKeep(f.looseVariable(node), source)
}

// looseEnter adds source to H when target holds a loose slot, and says whether it did.
func (f *cycleFinder) looseEnter(source, target *checker.Type) bool {
	if source == nil || !f.holdsLoose(target) {
		return false
	}
	f.looseHold(source)
	return true
}

// looseHold adds source to H: every shape when it's loose itself.
func (f *cycleFinder) looseHold(source *checker.Type) {
	if source == nil {
		return
	}
	if looseSource(source) {
		f.loose.every = true
		return
	}
	if f.loose.entered[source] {
		return
	}
	if f.loose.entered == nil {
		f.loose.entered = map[*checker.Type]bool{}
	}
	f.loose.entered[source] = true
	f.loose.entries = append(f.loose.entries, source)
}

// looseKeep keeps source as what's written at key, a write's site node or a variable's name.
func (f *cycleFinder) looseKeep(key *ast.Node, source *checker.Type) {
	if key == nil || source == nil {
		return
	}
	if f.loose.sources == nil {
		f.loose.sources = map[*ast.Node][]*checker.Type{}
	}
	f.loose.sources[key] = append(f.loose.sources[key], source)
}

// looseSite is node climbed out of its parentheses, and the node it stands in.
func looseSite(node *ast.Node) (*ast.Node, *ast.Node) {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	return site, site.Parent
}

// looseReceiver is the node lowering names as a store's site (writeSite) when node is the value stored:
// the object of an assignment's field or element, the receiver of a method call's argument (push, add,
// set), or the class whose field node initializes. It's nil for any other site.
func looseReceiver(node *ast.Node) *ast.Node {
	site, parent := looseSite(node)
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.Right != site || (binary.OperatorToken.Kind != ast.KindEqualsToken && !logicalAssignment(binary.OperatorToken.Kind)) {
			return nil
		}
		return looseAccessed(binary.Left)
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if call.Expression == site {
			return nil
		}
		return looseAccessed(call.Expression)
	case ast.KindPropertyDeclaration:
		if parent.AsPropertyDeclaration().Initializer == site && parent.Parent != nil {
			return parent.Parent.Name()
		}
	}
	return nil
}

// looseAccessed is the object a field or element access reads from, or nil for any other node.
func looseAccessed(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// looseVariable is the declared name of the variable node is written into, a declaration's initializer,
// a parameter's default or an assignment's right, as lowering keeps a local's node (noteLocal), or nil.
func (f *cycleFinder) looseVariable(node *ast.Node) *ast.Node {
	site, parent := looseSite(node)
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		if parent.AsVariableDeclaration().Initializer == site {
			return parent.Name()
		}
	case ast.KindParameter:
		if parent.AsParameterDeclaration().Initializer == site {
			return parent.Name()
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		left := ast.SkipParentheses(binary.Left)
		if binary.Right != site || !ast.IsIdentifier(left) || (binary.OperatorToken.Kind != ast.KindEqualsToken && !logicalAssignment(binary.OperatorToken.Kind)) {
			return nil
		}
		if symbol := f.l.checker.GetSymbolAtLocation(left); symbol != nil && symbol.ValueDeclaration != nil {
			return symbol.ValueDeclaration.Name()
		}
	}
	return nil
}

// looseNode is whether a type on a reaching path is a loose slot's own, unknown or object, which
// reachLinks doesn't follow.
func looseNode(proven *checker.Type) bool {
	return proven.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0
}

// looseSource is whether a value of the type says nothing of what it holds: unknown, object, any or a
// type parameter, or a union or intersection with one among its members.
func looseSource(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if proven.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive|checker.TypeFlagsAny|checker.TypeFlagsInstantiableNonPrimitive) != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if looseSource(member) {
				return true
			}
		}
	}
	return false
}

// holdsLoose is whether a value seen as the type lands in a loose slot: unknown or object, through unions
// and intersections, and through what a container holds (the comment at the top of this file). An
// object's own fields are slots of their own, judged where they're written or initialized, so they
// aren't looked through.
func (f *cycleFinder) holdsLoose(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if value, ok := f.loose.targets[proven]; ok {
		return value
	}
	if f.loose.targets == nil {
		f.loose.targets = map[*checker.Type]bool{}
	}
	// A type met again while it's judged adds nothing on that road; another road decides.
	f.loose.targets[proven] = false
	value := f.deriveHoldsLoose(proven)
	f.loose.targets[proven] = value
	return value
}

func (f *cycleFinder) deriveHoldsLoose(proven *checker.Type) bool {
	flags := proven.Flags()
	if looseNode(proven) {
		return true
	}
	if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if f.holdsLoose(member) {
				return true
			}
		}
		return false
	}
	if flags&checker.TypeFlagsObject == 0 {
		return false
	}
	if element := f.l.recordElement(proven); element != nil && f.holdsLoose(element) {
		return true
	}
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			if f.holdsLoose(argument) {
				return true
			}
		}
	}
	for _, info := range f.l.checker.GetIndexInfosOfType(proven) {
		if f.holdsLoose(info.ValueType()) {
			return true
		}
	}
	return false
}

// looseFields is whether an object spread into a value seen as the type can land in a loose slot: the
// type holds one, or a field of it does.
func (f *cycleFinder) looseFields(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if f.holdsLoose(proven) {
		return true
	}
	members := []*checker.Type{proven}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject == 0 {
			continue
		}
		for _, property := range f.l.checker.GetPropertiesOfType(member) {
			if f.holdsLoose(f.l.checker.GetTypeOfSymbol(property)) {
				return true
			}
		}
	}
	return false
}

// looseCellEvery is whether a captured variable typed loose can hold values no site looseEntry reads put
// there: anything but a variable a declaration statement declares (a parameter's argument may come from a
// callback's caller, a catch binding's from a throw, a loop or destructured binding's from what it
// reads), so it holds every shape.
func (f *cycleFinder) looseCellEvery(local int) bool {
	proven := f.l.localTypes[local]
	if proven == nil || !f.holdsLoose(proven) {
		return false
	}
	node := f.l.localNodes[local]
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindVariableDeclaration {
		return true
	}
	list := node.Parent.Parent
	return list == nil || list.Kind != ast.KindVariableDeclarationList || list.Parent == nil || list.Parent.Kind != ast.KindVariableStatement
}

// looseLinks is what a loose type on a path links to: every type in H, or every shape when H is.
func (f *cycleFinder) looseLinks() []cycleNode {
	if f.loose.every {
		return f.looseAll()
	}
	links := []cycleNode{}
	for _, entered := range f.loose.entries {
		links = append(links, cycleNode{proven: entered})
	}
	return links
}

// looseAll is every type a value can be: each type the program uses, each shape, each function value's
// type (which links to the cells it captured) and each type in H.
func (f *cycleFinder) looseAll() []cycleNode {
	if f.loose.everything != nil && f.loose.everythingShapes == len(f.shapes) {
		return f.loose.everything
	}
	all := []cycleNode{}
	included := map[*checker.Type]bool{}
	add := func(proven *checker.Type) {
		if proven != nil && !included[proven] {
			included[proven] = true
			all = append(all, cycleNode{proven: proven})
		}
	}
	for _, proven := range f.seen {
		add(proven)
	}
	for _, shape := range f.shapes {
		add(shape)
	}
	for _, closure := range f.l.closureRecords {
		add(closure.proven)
	}
	for _, entered := range f.loose.entries {
		add(entered)
	}
	f.loose.everything, f.loose.everythingShapes = all, len(f.shapes)
	return all
}

// looseLinksOf is where reaching goes from a node with H linked in.
func (f *cycleFinder) looseLinksOf(node cycleNode) []cycleNode {
	if node.proven != nil && !f.weak(node.proven) && looseNode(node.proven) {
		return f.looseLinks()
	}
	if f.loose.edges == nil || f.loose.edgeShapes != len(f.shapes) {
		f.loose.edges = map[cycleNode][]cycleNode{}
		f.loose.edgeShapes = len(f.shapes)
	}
	links, ok := f.loose.edges[node]
	if !ok {
		links = f.reachLinks(node)
		f.loose.edges[node] = links
	}
	if node.cell != 0 && f.looseCellEvery(node.cell-1) {
		return append(append([]cycleNode{}, links...), f.looseAll()...)
	}
	return links
}

// looseArrived is whether reaching has come to target, as reaches judges it: the cell itself, or a type
// one value could be seen as both.
func (f *cycleFinder) looseArrived(node, target cycleNode) bool {
	if node.proven != nil {
		node.proven = f.literalType(node.proven)
	}
	return (node.cell != 0 && node == target) || (node.proven != nil && !f.weak(node.proven) && !f.template(node.proven) && node.proven.Flags()&checker.TypeFlagsObject != 0 && target.proven != nil && f.related(node.proven, target.proven))
}

// looseReaches reports whether a value of the type, every shape when it's loose itself, can reach target
// with H linked in.
func (f *cycleFinder) looseReaches(value *checker.Type, target cycleNode) bool {
	start := []cycleNode{{proven: value}}
	if looseSource(value) {
		// Every shape includes the holder itself, so a value that can be anything reaches it at once.
		if target.proven != nil {
			return true
		}
		start = f.looseAll()
	}
	// A start that is the target already, as a holder among every shape is, needs no search.
	for _, node := range start {
		if f.looseArrived(node, target) {
			return true
		}
	}
	visited := map[cycleNode]bool{}
	queue := append([]cycleNode{}, start...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if node.proven != nil {
			node.proven = f.literalType(node.proven)
		}
		if visited[node] {
			continue
		}
		visited[node] = true
		if f.looseArrived(node, target) {
			return true
		}
		queue = append(queue, f.looseLinksOf(node)...)
	}
	return false
}

// looseScalar is whether an IR value holds no reference whatever it's boxed as: a number, a boolean, a
// string, or undefined or null.
func looseScalar(value ir.Expression) bool {
	for {
		switch inner := value.(type) {
		case ir.Box:
			value = inner.Value
			continue
		case ir.Narrow:
			value = inner.Value
			continue
		case ir.Unwrap:
			value = inner.Value
			continue
		case ir.Defined:
			value = inner.Value
			continue
		case ir.CheckedCast:
			value = inner.Value
			continue
		case ir.Undefined, ir.Null:
			return true
		}
		break
	}
	switch value.Type() {
	case ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
		return true
	}
	return false
}

// noteLooseInit keeps each field, element or entry of a literal lowering just made, read in the
// instantiation being lowered, for the stop to judge once H is whole; the stop keeps only those that land
// in a loose slot. A literal is made as the slot it's written into, unless that slot is loose itself,
// where it keeps its own type.
func (l *lowering) noteLooseInit(node *ast.Node) {
	holderOf := func() (*checker.Type, *checker.Type) {
		own := l.concrete(l.checker.GetTypeAtLocation(node))
		if slot := l.slotType(node); slot != nil {
			if slot = l.concrete(l.checker.GetNonNullableType(slot)); !looseSource(slot) {
				return slot, own
			}
		}
		return own, own
	}
	note := func(holder, slot *checker.Type, value, at *ast.Node, field string) {
		l.looseInits = append(l.looseInits, looseInit{holder: holder, slot: slot, value: l.concrete(l.checker.GetTypeAtLocation(value)), node: at, field: field})
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		holder, own := holderOf()
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			var value *ast.Node
			switch property.Kind {
			case ast.KindPropertyAssignment:
				value = property.AsPropertyAssignment().Initializer
			case ast.KindShorthandPropertyAssignment:
				value = property.Name()
			case ast.KindSpreadAssignment:
				// A spread can fill any field, so it lands in a loose slot if the literal has one.
				note(holder, nil, property.AsSpreadAssignment().Expression, property, "...")
				continue
			default:
				continue
			}
			if value == nil {
				continue
			}
			if property.Name() == nil || property.Name().Kind == ast.KindComputedPropertyName {
				note(holder, nil, value, property, "[]")
				continue
			}
			key := property.Name().Text()
			slot := l.checker.GetTypeOfPropertyOfType(holder, key)
			if slot == nil {
				slot = l.checker.GetTypeOfPropertyOfType(own, key)
			}
			if slot == nil {
				for _, info := range l.checker.GetIndexInfosOfType(holder) {
					slot = info.ValueType()
				}
			}
			note(holder, slot, value, property, key)
		}
	case ast.KindArrayLiteralExpression:
		holder, _ := holderOf()
		arguments := []*checker.Type{}
		if holder.Flags()&checker.TypeFlagsObject != 0 && holder.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.isArrayType(holder) || checker.IsTupleType(holder)) {
			arguments = l.checker.GetTypeArguments(holder)
		}
		for index, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			var slot *checker.Type
			if len(arguments) > 0 {
				slot = arguments[min(index, len(arguments)-1)]
			}
			if element.Kind == ast.KindSpreadElement {
				element = element.AsSpreadElement().Expression
			}
			note(holder, slot, element, element, "")
		}
	case ast.KindNewExpression:
		holder := l.concrete(l.checker.GetTypeAtLocation(node))
		arguments := node.AsNewExpression().Arguments
		if arguments == nil || len(arguments.Nodes) == 0 || !l.isLibraryType(holder, "Map", "Set") {
			return
		}
		for _, argument := range l.checker.GetTypeArguments(holder) {
			note(holder, argument, arguments.Nodes[0], node, "")
		}
	}
}

// looseInitWhat names an initialization for its stop: a field by its name, a spread, a computed field,
// or a container's elements or entries.
func looseInitWhat(l *lowering, made looseInit) string {
	holder := l.checker.TypeToString(made.holder)
	switch made.field {
	case "":
		return "the elements or entries of " + holder
	case "...":
		return "a spread into " + holder
	case "[]":
		return "a computed field of " + holder
	}
	return holder + "." + made.field + ", a field"
}

// looseWrite is one value entering a loose slot: its holder, the slot's type, the IR value and the
// types the relation sites read for it, where it is and what to call it.
type looseWrite struct {
	holder  *checker.Type
	slot    *checker.Type
	value   ir.Expression
	sources []*checker.Type
	cell    int
	stop    func(string) error
	what    string
}

// looseSlotStop judges the program's loose slots (the comment at the top of this file): H is gathered first,
// from looseEntry's sites and every literal's loose initialization, then each arm judges against it.
func (f *cycleFinder) looseSlotStop() error {
	inits := []looseInit{}
	for _, made := range f.l.looseInits {
		if made.slot != nil && !f.holdsLoose(made.slot) {
			continue
		}
		if made.slot == nil && !f.looseFields(made.holder) && !f.holdsLoose(made.holder) {
			continue
		}
		f.looseHold(made.value)
		inits = append(inits, made)
	}
	// H is whole now. The entry arm judges each value entering a loose slot; the declared arm, kept for the
	// first measured run beside it, judges each declared slot whose reach back crosses a loose slot. When a
	// test observes the arms, both run in full, so a census can tell which one walls a program.
	observe := looseArmsObserved.Load()
	var first error
	for _, arm := range []struct {
		name  string
		judge func() error
	}{{"entry", func() error { return f.looseEntryStop(inits) }}, {"declared", f.looseDeclaredStop}} {
		err := arm.judge()
		if observe != nil {
			(*observe)(f.l.result.Source, arm.name, err)
		}
		if err != nil && first == nil {
			first = err
			if observe == nil {
				return err
			}
		}
	}
	return first
}

// looseArmsObserved, when a test sets it, hears each arm's first stop, or nil, for each program judged: the
// program's source, the arm ("entry" or "declared") and its stop. It's nil outside those tests
// (loose_slots_test.go).
var looseArmsObserved atomic.Pointer[func(source, arm string, err error)]

// looseEntryStop judges every value entering a loose slot against H: each write, each captured variable
// and each literal's initialization, and the first that can reach back stops.
func (f *cycleFinder) looseEntryStop(inits []looseInit) error {
	l := f.l
	writes := f.looseWrites()
	for _, write := range writes {
		if write.value != nil && looseScalar(write.value) {
			continue
		}
		sources := write.sources
		if len(sources) == 0 {
			// No site this walk read names the value (a library loop's own write): the slot's own type stands
			// in, which reaches whatever H holds.
			sources = []*checker.Type{write.slot}
		}
		target := cycleNode{proven: write.holder}
		if write.cell != 0 {
			target = cycleNode{cell: write.cell}
		}
		for _, source := range sources {
			if source == nil || viewPairScalar(source) || !f.looseReaches(source, target) {
				continue
			}
			if write.cell != 0 && l.closedFrameInput(write.cell-1) {
				// A closed frame input is proven not to close a cycle, as the seeds take it (graph_types.go).
				continue
			}
			return write.stop(write.what + ", written a value of type " + l.checker.TypeToString(source) + " that can reach back to " + looseBack(l, write) + " through what a value held as unknown or object can be: " + looseStopWords)
		}
	}
	for local, declared := range l.result.Locals {
		if !declared.Captured || declared.Global || !f.looseCellEvery(local) || !f.looseReaches(l.localTypes[local], cycleNode{cell: local + 1}) || l.closedFrameInput(local) {
			continue
		}
		what := "'" + declared.Name + "', a variable a function value captures, of type " + l.checker.TypeToString(l.localTypes[local]) + ", bound to whatever its callers or what it reads hand it, which can reach its own cell back through what a value held as unknown or object can be: " + looseStopWords
		if node := l.localNodes[local]; node != nil {
			return l.notYet(node, what)
		}
		return &NotYet{Where: l.result.Source, What: what}
	}
	for _, made := range inits {
		if made.value == nil || viewPairScalar(made.value) || !f.looseReaches(made.value, cycleNode{proven: made.holder}) {
			continue
		}
		slot := "a slot typed unknown or object"
		if made.slot != nil {
			slot = l.checker.TypeToString(made.slot)
		}
		return l.notYet(made.node, looseInitWhat(l, made)+", of type "+slot+", initialized with a value of type "+l.checker.TypeToString(made.value)+" that can reach back to the "+l.checker.TypeToString(made.holder)+" holding it through what a value held as unknown or object can be: "+looseStopWords)
	}
	return nil
}

// looseBack names what a write can reach back to: the holder, or the variable's own cell.
func looseBack(l *lowering, write looseWrite) string {
	if write.cell != 0 {
		return "its own cell"
	}
	return "the " + l.checker.TypeToString(write.holder) + " holding it"
}

// looseWrites is every value an IR node writes into a loose slot, by its Site, and every value a
// captured variable typed loose is declared or assigned: the same nodes fresh.go judges writes by, a
// field (SetProperty), an element (SetIndex, ArrayPush, ArrayFill, ArraySplice), a Map's key and value
// (MapSet), a Set's element (SetAdd) and a record's value (RecordCall's set and defineHidden,
// RecordCoalesce).
func (f *cycleFinder) looseWrites() []looseWrite {
	l := f.l
	writes := []looseWrite{}
	add := func(site int, value ir.Expression, slot func(*checker.Type) (*checker.Type, string)) {
		if site <= 0 || site > len(l.writeSites) {
			return
		}
		written := l.writeSites[site-1]
		holder := written.holder
		if holder == nil {
			// The three writes lowering made with no holder (libraryArraySyntheticWriteSite,
			// libraryIteratorArray, libraryOptionalNumbers) write into what their call makes.
			holder = l.concrete(l.checker.GetTypeAtLocation(written.node))
		}
		if holder == nil || f.template(holder) {
			return
		}
		proven, what := slot(holder)
		if proven == nil || f.weak(proven) || !f.holdsLoose(proven) {
			return
		}
		node := written.node
		writes = append(writes, looseWrite{holder: holder, slot: proven, value: value, sources: f.loose.sources[node], what: what, stop: func(what string) error {
			return l.notYet(node, what)
		}})
	}
	field := func(name string) func(*checker.Type) (*checker.Type, string) {
		return func(holder *checker.Type) (*checker.Type, string) {
			for _, declared := range f.fields(holder) {
				if l.cycleFieldMatches(holder, declared.Name, name) {
					return l.checker.GetTypeOfSymbol(declared), l.checker.TypeToString(holder) + "." + l.cycleFieldName(declared) + ", a field of type " + l.checker.TypeToString(l.checker.GetTypeOfSymbol(declared))
				}
			}
			if element := l.recordElement(holder); element != nil {
				return element, "the values of " + l.checker.TypeToString(holder)
			}
			return nil, ""
		}
	}
	argument := func(index int, kind string) func(*checker.Type) (*checker.Type, string) {
		return func(holder *checker.Type) (*checker.Type, string) {
			if element := l.recordElement(holder); element != nil {
				return element, "the values of " + l.checker.TypeToString(holder)
			}
			arguments := f.graphArguments(holder)
			if index < len(arguments) {
				// A tuple's positions are judged as one: whichever holds a loose slot.
				for _, each := range arguments[index:] {
					if f.holdsLoose(each) {
						return each, kind + " of " + l.checker.TypeToString(holder)
					}
					if kind != "the elements" {
						break
					}
				}
				return arguments[index], kind + " of " + l.checker.TypeToString(holder)
			}
			return nil, ""
		}
	}
	visit := func(node any) bool {
		switch value := node.(type) {
		case ir.SetProperty:
			add(value.Site, value.Value, field(value.Name))
		case ir.SetIndex:
			add(value.Site, value.Value, argument(0, "the elements"))
		case ir.ArrayPush:
			add(value.Site, value.Value, argument(0, "the elements"))
		case ir.ArrayFill:
			add(value.Site, value.Value, argument(0, "the elements"))
		case ir.ArraySplice:
			for _, item := range value.Items {
				add(value.Site, item, argument(0, "the elements"))
			}
		case ir.MapSet:
			add(value.Site, value.Key, argument(0, "the keys"))
			add(value.Site, value.Value, argument(1, "the values"))
		case ir.SetAdd:
			add(value.Site, value.Value, argument(0, "the elements"))
		case ir.RecordCall:
			if (value.Method == "set" || value.Method == "defineHidden") && len(value.Arguments) == 3 {
				add(value.Site, value.Arguments[2], argument(0, "the values"))
			}
		case ir.RecordCoalesce:
			add(value.Site, value.Value, argument(0, "the values"))
		case ir.Declare:
			if value.Value != nil {
				f.looseCellWrite(&writes, value.Local, value.Value)
			}
		case ir.Assign:
			f.looseCellWrite(&writes, value.Local, value.Value)
		}
		return true
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	return writes
}

// looseCellWrite adds a value written into a captured variable typed loose that a declaration statement
// declares; any other such variable is judged whole (looseCellEvery).
func (f *cycleFinder) looseCellWrite(writes *[]looseWrite, local int, value ir.Expression) {
	l := f.l
	if local < 0 || local >= len(l.result.Locals) {
		return
	}
	declared := l.result.Locals[local]
	proven := l.localTypes[local]
	if !declared.Captured || declared.Global || proven == nil || f.weak(proven) || !f.holdsLoose(proven) || f.looseCellEvery(local) {
		return
	}
	node := l.localNodes[local]
	*writes = append(*writes, looseWrite{slot: proven, value: value, sources: f.loose.sources[node], cell: local + 1, what: "'" + declared.Name + "', a variable a function value captures, of type " + l.checker.TypeToString(proven), stop: func(what string) error {
		if node != nil {
			return l.notYet(node, what)
		}
		return &NotYet{Where: l.result.Source, What: what}
	}})
}

// looseDeclaredStop is the declared arm: a declared slot, one that holds no loose slot itself, refused when
// a write fresh.go doesn't prove may close a cycle whose reach back to the holder crosses a loose slot.
// The entry arm should already stop every such cycle where its value entered the loose slot (the comment
// at the top of this file); this arm stays for the first measured run, so the census can say what each
// walls and whether the entry arm alone holds k17 and k18.
func (f *cycleFinder) looseDeclaredStop() error {
	for _, holder := range f.seen {
		if err := f.looseDeclaredSlots(holder); err != nil {
			return err
		}
	}
	return nil
}

// looseDeclaredSlots judges a holder's declared slots as slotsOf finds them: an array's or a tuple's
// elements unless it's readonly, a record's values, a Set's elements, a Map's keys and values, and an
// object's fields.
func (f *cycleFinder) looseDeclaredSlots(holder *checker.Type) error {
	l := f.l
	if holder.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(holder) {
		return nil
	}
	name := l.checker.TypeToString(holder)
	switch {
	case l.isArrayType(holder) || checker.IsTupleType(holder):
		if l.isLibraryType(holder, "ReadonlyArray") || (checker.IsTupleType(holder) && holder.TargetTupleType().IsReadonly()) {
			return nil
		}
		for _, element := range l.checker.GetTypeArguments(holder) {
			if err := f.looseDeclaredSlot(holder, element, fresh.WriteElement, "", "the elements of "+name); err != nil {
				return err
			}
		}
	case l.recordElement(holder) != nil:
		return f.looseDeclaredSlot(holder, l.recordElement(holder), fresh.WriteMapEntry, "", "the values of "+name)
	case l.isLibraryType(holder, "ReadonlyMap"), l.isLibraryType(holder, "ReadonlySet"):
	case l.isLibraryType(holder, "Set"):
		if arguments := l.checker.GetTypeArguments(holder); len(arguments) == 1 {
			return f.looseDeclaredSlot(holder, arguments[0], fresh.WriteSetElement, "", "the elements of "+name)
		}
	case l.isLibraryType(holder, "Map"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) != 2 {
			return nil
		}
		if err := f.looseDeclaredSlot(holder, arguments[0], fresh.WriteMapEntry, "", "the keys of "+name); err != nil {
			return err
		}
		return f.looseDeclaredSlot(holder, arguments[1], fresh.WriteMapEntry, "", "the values of "+name)
	default:
		for _, field := range f.fields(holder) {
			if err := f.looseDeclaredSlot(holder, l.checker.GetTypeOfSymbol(field), fresh.WriteField, field.Name, name+"."+l.cycleFieldName(field)+", a field"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *cycleFinder) looseDeclaredSlot(holder, slot *checker.Type, kind fresh.WriteKind, field, what string) error {
	if slot == nil || f.weak(slot) || f.holdsLoose(slot) || !f.looseTouches(cycleNode{proven: slot}) {
		return nil
	}
	write := f.cycleUnproven(kind, holder, field)
	if write == nil || !f.looseCrosses(slot, holder) {
		return nil
	}
	what += " of type " + f.l.checker.TypeToString(slot) + ", which can reach back to the " + f.l.checker.TypeToString(holder) + " holding it only through what a value held as unknown or object can be, and " + f.writtenAt(write) + " may close the cycle: " + looseStopWords
	if write.Site <= 0 || write.Site > len(f.l.writeSites) {
		return &NotYet{Where: f.l.result.Source, What: what}
	}
	return f.l.notYet(f.l.writeSites[write.Site-1].node, what)
}

// looseCrossing is whether reaching leaves a loose type at node: unknown or object, or a cell holding every
// shape.
func (f *cycleFinder) looseCrossing(node cycleNode) bool {
	if node.proven != nil {
		return !f.weak(node.proven) && looseNode(node.proven)
	}
	return node.cell != 0 && f.looseCellEvery(node.cell-1)
}

// looseStep is a node reached, and whether the path to it left a loose type.
type looseStep struct {
	node    cycleNode
	crossed bool
}

// looseCrosses reports whether a slot's type reaches the holder back along a path that leaves a loose type.
func (f *cycleFinder) looseCrosses(slot, holder *checker.Type) bool {
	target := cycleNode{proven: holder}
	visited := map[looseStep]bool{}
	queue := []looseStep{{node: cycleNode{proven: slot}}}
	for len(queue) != 0 {
		step := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if step.node.proven != nil {
			step.node.proven = f.literalType(step.node.proven)
		}
		// A path that left a loose type reaches everything one that didn't does.
		if visited[step] || (!step.crossed && visited[looseStep{node: step.node, crossed: true}]) {
			continue
		}
		visited[step] = true
		if step.crossed && f.looseArrived(step.node, target) {
			return true
		}
		crossed := step.crossed || f.looseCrossing(step.node)
		for _, link := range f.looseLinksOf(step.node) {
			queue = append(queue, looseStep{node: link, crossed: crossed})
		}
	}
	return false
}

// looseTouches is whether reaching from a node can come to a loose type at all, so a declared slot that
// can't skips the search. It's worked out once for every node the program's types reach; a node outside
// them reads as touching, failing closed.
func (f *cycleFinder) looseTouches(node cycleNode) bool {
	if node.proven != nil {
		node.proven = f.literalType(node.proven)
	}
	if f.loose.touching == nil {
		f.loose.touching = f.deriveLooseTouching()
	}
	touches, known := f.loose.touching[node]
	return touches || !known
}

func (f *cycleFinder) deriveLooseTouching() map[cycleNode]bool {
	touching := map[cycleNode]bool{}
	reverse := map[cycleNode][]cycleNode{}
	queue := []cycleNode{}
	push := func(node cycleNode) {
		if node.proven != nil {
			node.proven = f.literalType(node.proven)
		}
		if _, known := touching[node]; !known {
			touching[node] = false
			queue = append(queue, node)
		}
	}
	for _, proven := range f.seen {
		push(cycleNode{proven: proven})
	}
	for _, shape := range f.shapes {
		push(cycleNode{proven: shape})
	}
	for _, closure := range f.l.closureRecords {
		push(cycleNode{proven: closure.proven})
	}
	for local, declared := range f.l.result.Locals {
		if declared.Captured && !declared.Global {
			push(cycleNode{cell: local + 1})
		}
	}
	loose := []cycleNode{}
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if f.looseCrossing(node) {
			loose = append(loose, node)
			continue
		}
		for _, link := range f.looseLinksOf(node) {
			if link.proven != nil {
				link.proven = f.literalType(link.proven)
			}
			reverse[link] = append(reverse[link], node)
			push(link)
		}
	}
	for len(loose) != 0 {
		node := loose[len(loose)-1]
		loose = loose[:len(loose)-1]
		if touching[node] {
			continue
		}
		touching[node] = true
		loose = append(loose, reverse[node]...)
	}
	return touching
}
