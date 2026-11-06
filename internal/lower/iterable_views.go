package lower

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A slot typed Iterable<T> or IterableIterator<T> is held as an object (as a union for
// Iterable<string>, which a string fits too), and nothing in the type says which object: an array, a
// generator, a string, a Map itself or a program's own iterator all fit it, and each steps
// differently. A union-held view reads its slot through a checked unbox, the loud backstop should
// this proof ever be wrong. tsc's mapIterator and mapDefinedIterator take one and are handed a
// collection's iterator (`jsxOptionMap.entries()`, `fileInfos.keys()`), so a for-of over such a slot
// lowers as that iterator's own loop (libraryForOfIterator), and is proven after lowering, when every
// call that can reach the slot is in, to be handed nothing else (#qk8rztp S2b-1b, Lower's ruling on q1).
//
// The proof asks the allocation graph the view passes use (ReachingAllocations) which values reach the
// iterated expression, with an entry hook that stops the flow at the first value that isn't a road
// (a read, a call's result, a conditional's arm, a box, a narrowing, a cast). A collection's iterator,
// ir.CollectionIterator, is the one listed kind. Anything else the flow ends at is an origin the proof
// can't name, and the view is NotYet: a field, a literal, the parameter of a function used as a value
// or of an object literal's method (whose callers the graph doesn't see), a call with no known target.
// The hook stops at a field rather than projecting through it, so no projection depth bound is ever
// reached. Arrays and Maps never get this far: going where a structural type is held, they stop at the
// call (shape_fit.go).
//
// Every origin must hold its step's value as the view reads it, decided here and never converted in
// the step: number for a Map's values() read as Iterable<number>, and for an entries() pair, its key
// and its value as the tuple's two elements are held.

// iterableView is one for-of forOf lowered as a view: where it stops if unproven, the value it steps,
// how it reads each step, the function its loop is in, and the local its loop holds the iterator in
// (viewLoop), -1 when it can't be named. A hand view is one hand-driven `x.next()` instead
// (handNextView), which has no loop: its origins are proven the same way, and nothing rewrites it.
type iterableView struct {
	where    *ast.Node
	iterated ir.Expression
	held     iterableElement
	function int
	holder   int
	hand     bool
}

// iterableElement is how a step's value is held: element, and when it is a [key, value] pair, how the
// pair holds each half. A view's pair says how its T, a two-element tuple, holds them. An origin's
// says an entries() iterator makes the pair each step, holding the collection's key and value as the
// collection does.
type iterableElement struct {
	element    ir.Type
	pair       bool
	key, value ir.Type
}

// iterableOrigin is a value the proof's flow stopped at: a collection's iterator, held as it says, a
// generator's call handed in through IterableOf (generator is its function plus one), held as it
// yields, or anything else, by its kind.
type iterableOrigin struct {
	name       string
	collection bool
	generator  int
	held       iterableElement
}

// viewLoop notes, once forOf has lowered a view's loop, the function it is in and the local that
// holds its iterator, which the loop's block declares first (libraryForOfIterator), so that
// checkIterableViews can rewrite a loop a generator feeds into its dispatch.
func (l *lowering) viewLoop(view int, statements []ir.Statement, err error) ([]ir.Statement, error) {
	if err != nil {
		return nil, err
	}
	l.iterableViews[view].function = l.functionIndex
	l.iterableViews[view].holder = -1
	if len(statements) == 1 {
		if block, isBlock := statements[0].(ir.Block); isBlock && len(block.Body) == 2 {
			if declared, isDeclare := block.Body[0].(ir.Declare); isDeclare {
				l.iterableViews[view].holder = declared.Local
			}
		}
	}
	return statements, nil
}

// handDrivenNext reports whether a call is a hand-driven `x.next()` over a view: the direct call of
// lib's next, with no `?.` and no parentheses around the callee, on a bare name typed the
// library's Iterable<T> or IterableIterator<T> (or either `| undefined`, narrowed), held as an object,
// that names a parameter (Adamic's (a) on #qk8rztp S2c-2). A detached next (`const n = x.next`,
// `x.next.call`, a passed `x.next`) is no call of it, so the library-member guard still refuses it;
// a call with an argument is one, which callClosure stops not yet; and that the parameter is the function's own and its every origin a
// collection's iterator is proven after lowering (handNextView, checkIterableViews).
func (l *lowering) handDrivenNext(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || call.Expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	callee := call.Expression
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || callee.Flags&ast.NodeFlagsOptionalChain != 0 || callee.Name().Text() != "next" || !ast.IsIdentifier(access.Expression) {
		return false
	}
	if !l.inheritedLibraryMember(callee) {
		return false
	}
	if _, isView := l.iterableViewElement(access.Expression); !isView {
		return false
	}
	if held, known := l.representation(l.checker.GetTypeAtLocation(access.Expression)); !known || held != ir.Object {
		return false
	}
	symbol := l.symbol(access.Expression)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindParameter {
			return false
		}
	}
	return len(symbol.Declarations) == 1
}

// handNextView records a hand-driven next's receiver as a view, once callClosure has lowered the
// callee: a read, through the narrowing's ir.Defined or not (iterableViewRead), of one of the
// function's own parameters. A parameter of an enclosing function, read from a closure, is no view,
// and is refused as main refuses any library next.
func (l *lowering) handNextView(node *ast.Node, closure ir.Expression) error {
	callee := node.AsCallExpression().Expression
	property, isProperty := closure.(ir.Property)
	if !isProperty {
		return l.prototypeRead(callee, "next")
	}
	read, isRead := iterableViewRead(property.Object)
	if !isRead || l.function == nil || !slices.Contains(l.function.Parameters, read.Local) {
		return l.prototypeRead(callee, "next")
	}
	held, _ := l.iterableViewElement(callee.AsPropertyAccessExpression().Expression)
	l.iterableViews = append(l.iterableViews, iterableView{where: node, iterated: property.Object, held: held, function: l.functionIndex, holder: -1, hand: true})
	return nil
}

// generatorOrigin is how a generator's call handed in through IterableOf holds what it steps: what
// its function yields, and for a [key, value] tuple, how the tuple holds each half
// (generatorPairTypes, read from what the call yields, never the declared iterator type).
func (l *lowering) generatorOrigin(of ir.IterableOf) iterableOrigin {
	call, isCall := of.Value.(ir.Call)
	if !isCall || call.Returns != ir.Generator || call.Virtual != 0 || of.Site < 1 || of.Site > len(l.generatorArguments) {
		return iterableOrigin{name: "a generator seen as an object other than by its own call"}
	}
	// A direct call has the one target it names.
	targets := l.result.CallTargets(call)
	if len(targets) != 1 {
		return iterableOrigin{name: "a generator seen as an object other than by its own call"}
	}
	generator := targets[0]
	argument := l.generatorArguments[of.Site-1]
	held := iterableElement{element: l.result.Functions[generator].Yields}
	if argument.pair {
		held.pair, held.key, held.value = true, argument.key, argument.value
	}
	return iterableOrigin{name: "a generator from " + l.result.Functions[generator].Name + "()", generator: generator + 1, held: held}
}

// iterableViewElement says whether a for-of's iterated value is typed the library's Iterable<T> or
// IterableIterator<T>, with T held in one boxed slot, and how T is held. TReturn and TNext don't
// matter to a for-of: it drops the return value and sends nothing.
func (l *lowering) iterableViewElement(iterated *ast.Node) (iterableElement, bool) {
	proven := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(iterated)))
	if !l.isLibraryType(proven, "Iterable", "IterableIterator") {
		return iterableElement{}, false
	}
	arguments := l.typeArguments(proven)
	if len(arguments) == 0 {
		return iterableElement{}, false
	}
	yielded := l.concrete(arguments[0])
	element, known := l.representation(yielded)
	if !known || !boxedSlot(element) {
		return iterableElement{}, false
	}
	held := iterableElement{element: element}
	if checker.IsTupleType(yielded) {
		if halves := l.typeArguments(yielded); len(halves) == 2 {
			key, keyKnown := l.representation(halves[0])
			value, valueKnown := l.representation(halves[1])
			if keyKnown && valueKnown {
				held.pair, held.key, held.value = true, key, value
			}
		}
	}
	return held, true
}

// collectionOrigin is how a collection's iterator holds what it steps: runtime/map_set.c's
// collection_next gives a key for keys() and for any Set, a value for a Map's values(), and for
// entries() a pair it makes of the key and the value.
func collectionOrigin(iterator ir.CollectionIterator) iterableOrigin {
	collection := "a Map's "
	if iterator.Set {
		collection = "a Set's "
	}
	origin := iterableOrigin{name: collection + iterator.Part + "()", collection: true}
	switch {
	case iterator.Part == "entries":
		origin.held = iterableElement{element: ir.Object, pair: true, key: iterator.Key, value: iterator.Value}
	case iterator.Part == "keys" || iterator.Set:
		origin.held = iterableElement{element: iterator.Key}
	default:
		origin.held = iterableElement{element: iterator.Value}
	}
	return origin
}

// readsAs says whether a view reading its steps as view reads an origin's steps, held as held, as
// they are: the same element, and for a pair the origin makes, the same key and value the view's
// tuple reads.
func (held iterableElement) readsAs(view iterableElement) bool {
	if held.element != view.element {
		return false
	}
	return !held.pair || view.pair && held.key == view.key && held.value == view.value
}

// checkIterableViews proves every view forOf lowered, or stops at the first it can't.
func (l *lowering) checkIterableViews() error {
	if len(l.iterableViews) == 0 && len(l.generatorArguments) == 0 {
		return nil
	}
	// A parameter stepped by hand is read only by its no-argument nexts and undefined tests (Adamic's
	// ruling on #qk8rztp S2c-2a, Compiler's must): a for-of beside them would step the same iterator,
	// and may close it, in an interleaving nothing has held to Node, so it stops not yet.
	hands := map[int]bool{}
	for _, view := range l.iterableViews {
		if read, isRead := iterableViewRead(view.iterated); isRead && view.hand {
			hands[read.Local] = true
		}
	}
	for _, view := range l.iterableViews {
		if read, isRead := iterableViewRead(view.iterated); isRead && !view.hand && hands[read.Local] {
			return l.notYet(view.where, "a for-of beside a hand-driven next() over the same parameter (#qk8rztp S2c-2a)")
		}
	}
	if err := l.checkGeneratorArguments(); err != nil {
		return err
	}
	var origins []iterableOrigin
	numbered := map[iterableOrigin]int{}
	// The sites the hook gives are positive, so no allocation's (which are negative) is one of them.
	site := func(origin iterableOrigin) int {
		if number, seen := numbered[origin]; seen {
			return number
		}
		origins = append(origins, origin)
		numbered[origin] = len(origins)
		return len(origins)
	}
	graph := newAllocationFlowGraph(l.result)
	graph.entry = func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.CollectionIterator:
			return site(collectionOrigin(value))
		case ir.IterableOf:
			return site(l.generatorOrigin(value))
		case ir.Read, ir.Call, ir.CallClosure, ir.Conditional, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast:
			// A road: the graph follows it to what it holds, or says why it can't.
			return 0
		}
		return site(iterableOrigin{name: "a value made by " + strings.TrimPrefix(fmt.Sprintf("%T", value), "ir.")})
	}
	for index, view := range l.iterableViews {
		// A view read where the checker narrowed undefined out is ir.Defined over the read
		// (narrowed.go's defined), which passes the value through or panics, so its origins are the
		// read's (#qk8rztp S2c-3). view.iterated itself stays the Defined: the loop's holder is
		// declared with it, keeping the narrowing's `NULL` check, so an undefined this proof missed
		// panics loudly in Defined's words and is never stepped.
		proven := view.iterated
		if viewed, isViewed := iterableViewRead(view.iterated); isViewed {
			proven = viewed
		}
		reaches := graph.ReachingAllocations(proven)
		unproven := slices.Clone(reaches.Reasons)
		var differently []string
		var generators []iterableOrigin
		for _, number := range reaches.Sites {
			if number < 1 || number > len(origins) {
				unproven = append(unproven, "an allocation")
				continue
			}
			origin := origins[number-1]
			switch {
			case origin.generator != 0:
				generators = append(generators, origin)
			case origin.collection:
				// The dispatch's collection arm steps it, or the loop itself does when no generator
				// feeds the view.
			default:
				unproven = append(unproven, origin.name)
				continue
			}
			if !origin.held.readsAs(view.held) {
				differently = append(differently, origin.name+" held as "+iterableHeldName(origin.held))
			}
		}
		if reaches.Unknown || len(unproven) != 0 {
			return l.notYet(view.where, "a view over an iterable of unproven origin ("+strings.Join(unproven, "; ")+")")
		}
		if len(differently) != 0 {
			return l.notYet(view.where, "a view over an iterable whose origins yield values held differently ("+strings.Join(differently, "; ")+", read as "+iterableHeldName(view.held)+")")
		}
		if view.hand && len(generators) != 0 {
			// A generator's step returns its value, not a result object (#qk8rztp S2c-2b).
			return l.notYet(view.where, "a hand-driven next() through a view a generator feeds (#qk8rztp S2c-2b)")
		}
		if len(generators) != 0 {
			if err := l.generatorView(index, generators); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkGeneratorArguments holds every generator's call a call hands into an Iterable<T> parameter
// to the one-read rule, the rest of Adamic's ruling on #qk8rztp S2c-1 (condition 3) and Lower's (a):
// the parameter belongs to the one body the call calls, which doesn't suspend or read arguments, no
// closure captures it, nothing assigns it, and it is read exactly once in the lowered program, as a
// view's for-of over it. The reads are counted in the IR, so a default that names the parameter, a
// capture or a second for-of counts as surely as a read written out; a type position reads nothing.
// A parameter typed `Iterable<T> | undefined` is admitted too, held as a nullable object, and its
// undefined tests (undefinedTestOf) are tests, not reads, since they never step or hand on what the
// slot holds (#qk8rztp S2c-3, tsc's firstOrUndefinedIterator). With one read and a direct call,
// nothing but that for-of can ever name the generator, so the for-of's exits closing it and the
// caller letting go of it is the whole of its life.
func (l *lowering) checkGeneratorArguments() error {
	views := map[int]int{}
	for index, view := range l.iterableViews {
		if read, isRead := iterableViewRead(view.iterated); isRead && !view.hand {
			views[read.Local] = index
		}
	}
	var failed error
	inspect := func(node any) bool {
		if failed != nil {
			return false
		}
		if call, isCall := node.(ir.Call); isCall {
			for position, argument := range call.Arguments {
				if of, isIterable := argument.(ir.IterableOf); isIterable {
					if failed = l.generatorParameter(call, position, of, views); failed != nil {
						return false
					}
				}
			}
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	return failed
}

// iterableViewRead is the read a view's for-of iterates: a bare ir.Read, or one ir.Defined, one level
// deep, whose value is a bare ir.Read, what a read of a slot the checker narrowed undefined out of
// lowers to, as a nullable parameter's is inside its undefined test (#qk8rztp S2c-3). Any other
// wrapper is no read of a view: the parameter isn't mapped to it, and the proof doesn't see through
// it.
func iterableViewRead(iterated ir.Expression) (ir.Read, bool) {
	if defined, isDefined := iterated.(ir.Defined); isDefined {
		iterated = defined.Value
	}
	read, isRead := iterated.(ir.Read)
	return read, isRead
}

// generatorParameter is checkGeneratorArguments for one generator handed in at one call. A nullable
// parameter must be held as an object, whose undefined is a `NULL` pointer the for-of's test has
// ruled out; `Iterable<string> | undefined` is held as a union, whose undefined is a member the
// dispatch's plain cast wasn't built to meet, so it stays not yet built (#qk8rztp S2c-3).
func (l *lowering) generatorParameter(call ir.Call, position int, of ir.IterableOf, views map[int]int) error {
	if of.Site < 1 || of.Site > len(l.generatorArguments) {
		return fmt.Errorf("lower: compiler bug: a generator seen as an object with no site")
	}
	where := l.generatorArguments[of.Site-1].node
	if call.Virtual != 0 || len(call.Spread) != 0 {
		return l.generatorNotYet(where, "passed into a parameter of no one body")
	}
	targets := l.result.CallTargets(call)
	if len(targets) != 1 || position >= len(l.result.Functions[targets[0]].Parameters) {
		return l.generatorNotYet(where, "passed into a parameter of no one body")
	}
	target := targets[0]
	callee := l.result.Functions[target]
	parameter := callee.Parameters[position]
	name := l.result.Locals[parameter].Name
	for _, view := range l.iterableViews {
		if read, isRead := iterableViewRead(view.iterated); view.hand && isRead && read.Local == parameter {
			return l.generatorNotYet(where, "passed into the parameter "+name+", which a hand-driven next() reads (#qk8rztp S2c-2b)")
		}
	}
	if index, isView := views[parameter]; !isView || l.iterableViews[index].function != target {
		return l.generatorNotYet(where, "passed into the parameter "+name+", whose one use isn't a for-of over it")
	}
	switch {
	case callee.Suspends():
		return l.generatorNotYet(where, "passed into a function that suspends")
	case callee.ReadsArguments || callee.ArgumentsCount != 0:
		return l.generatorNotYet(where, "passed into a function that reads arguments")
	case l.result.Locals[parameter].Captured:
		return l.generatorNotYet(where, "passed into the parameter "+name+", which a closure captures")
	case l.generatorArguments[of.Site-1].nullable && l.result.Locals[parameter].Type != ir.Object:
		return l.generatorNotYet(where, "passed into the parameter "+name+", which may be undefined and is held as a union")
	}
	// Every field that names a local counts, not only Read and Assign: a for-of head, a pattern's
	// binding and a catch write one too (Lower's must), and anything else that names it is a use. An
	// undefined test of the parameter is the one node not gone into, so its read isn't counted, and
	// nothing else is exempt (#qk8rztp S2c-3).
	var uses localUseCount
	count := func(node any) bool {
		if undefinedTestOf(node, parameter) {
			return false
		}
		uses.node(node, parameter)
		return true
	}
	walk(l.result.Main, count)
	for _, function := range l.result.Functions {
		walk(function.Body, count)
	}
	if uses.writes != 0 {
		return l.generatorNotYet(where, "passed into the parameter "+name+", which the function writes (an assignment, a for-of head, a pattern or a catch)")
	}
	if uses.named != 0 {
		return l.generatorNotYet(where, "passed into the parameter "+name+", which the lowered function names other than by a read")
	}
	if uses.reads != 1 {
		return l.generatorNotYet(where, fmt.Sprintf("passed into the parameter %s, read %d times where one for-of over it is its only use", name, uses.reads))
	}
	return nil
}

// generatorView admits a view a generator feeds, a collection's iterator feeding it too or not, held
// as an object or as a union (Iterable<string>), and rewrites its loop into the dispatch, or stops at
// what S2c-1 doesn't build: a view fed by more than one generator function. The step itself resumes
// through the frame's own code (adamic_generator_next
// calls generator->code->resume, and JavaScript's next() is the generator's own), so several
// functions could share it; what can't is GeneratorNext's one Function, whose ResumeMayThrow says
// whether the step's caller tests for a throw (Lower's seam).
func (l *lowering) generatorView(index int, generators []iterableOrigin) error {
	view := l.iterableViews[index]
	switch {
	case view.holder < 0 || view.function < 0:
		return fmt.Errorf("lower: compiler bug: the iterable view at %s has no loop to dispatch", l.program.Where(view.where))
	}
	for _, other := range generators[1:] {
		if other.generator != generators[0].generator {
			return l.notYet(view.where, "a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1)")
		}
	}
	return l.dispatchView(view, generators[0].generator-1)
}

// dispatchView rewrites a view's loop, lowered as a collection iterator's (libraryIteratorLoop), into
// its per-step dispatch on what the object is at run time, never its static type (Compiler's
// condition, Lower's (b)). The body is the one forOf lowered, kept as it is, and the step's value is
// assigned to the loop's item, declared at the top of each pass, by whichever arm steps:
//
//   - a generator, by GeneratorNext and GeneratorDone, as forOfGenerator's loop does;
//   - a collection's iterator, by libraryIteratorStep, the collection loop's own step, never a copy
//     (Lower's condition 4);
//   - anything else is a compiler bug said out loud, naming the site, since the proof listed every
//     origin (Lower's condition 5).
//
// The block closes what it holds at every way out (ir.Block's Closes), so a break, a return, a
// throw or a jump to an outer label finishes a generator, as Node's return() does, and a later pass
// over it finds it done (Lower's must); a collection's iterator has no return(), and the close
// leaves it as it is. The holder's declaration keeps the value forOf gave it, for a nullable
// parameter an ir.Defined over its read, so the narrowing's `NULL` check still runs before any arm
// (#qk8rztp S2c-3).
func (l *lowering) dispatchView(view iterableView, function int) error {
	yields := l.result.Functions[function].Yields
	held := ir.Read{Local: view.holder, Of: ir.Object}
	step := l.libraryLocal(view.function, "generator_step", yields)
	message := "compiler bug: the iterable view at " + l.program.Where(view.where) + " holds a value no arm of its dispatch steps"
	rewritten := false
	rewrite := func(node any) any {
		block, isBlock := node.(ir.Block)
		if !isBlock || len(block.Body) != 2 {
			return node
		}
		declared, isDeclare := block.Body[0].(ir.Declare)
		loop, isLoop := block.Body[1].(ir.Loop)
		if !isDeclare || declared.Local != view.holder || !isLoop || len(loop.Body) < 3 {
			return node
		}
		libraryStep, isStep := loop.Body[0].(ir.Declare)
		item, isItem := loop.Body[2].(ir.Declare)
		if !isStep || !isItem {
			return node
		}
		collectionArm, value := libraryIteratorStep(held, ir.Read{Local: libraryStep.Local, Of: ir.Object}, l.result.Locals[item.Local].Type)
		collectionArm = append(collectionArm, ir.Assign{Local: item.Local, Value: value})
		generatorArm := []ir.Statement{
			// The step is tested for a throw before done is read, as forOfGenerator's is.
			ir.Declare{Local: step.Local, Value: ir.GeneratorNext{Generator: held, Function: function, Of: yields}},
			ir.If{Condition: ir.GeneratorDone{Generator: held}, Then: []ir.Statement{ir.Break{}}},
			ir.Assign{Local: item.Local, Value: step},
		}
		dispatch := ir.If{
			Condition: ir.IsGenerator{Value: held},
			Then:      generatorArm,
			Else: []ir.Statement{ir.If{
				Condition: ir.IsCollectionIterator{Value: held},
				Then:      collectionArm,
				Else:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}},
			}},
		}
		loop.Body = append([]ir.Statement{ir.Declare{Local: item.Local}, dispatch}, loop.Body[3:]...)
		if view.iterated.Type() == ir.Union {
			// An Iterable<string> slot is a union, which S2b-1b's arm unboxes through Narrow's checked
			// Object arm, whose tag test would stop on a generator. Here the dispatch is that test: the
			// unbox is a plain pointer cast of the union's reference, whose heap kind IsGenerator and
			// IsCollectionIterator read, and a string or anything else reaches the loud default
			// (Lower's (e) on #qk8rztp). Narrow's Object arm stays as it is for every other caller.
			declared.Value = ir.Narrow{Value: view.iterated, To: ir.Object}
		}
		rewritten = true
		return ir.Block{Body: []ir.Statement{declared, loop}, Closes: view.holder + 1}
	}
	l.result.Functions[view.function].Body = rewriteShapeStatements(l.result.Functions[view.function].Body, rewrite)
	if !rewritten {
		return fmt.Errorf("lower: compiler bug: the iterable view at %s has no loop to dispatch", l.program.Where(view.where))
	}
	return nil
}

// iterableHeldName names how a step's value is held, for a stop.
func iterableHeldName(held iterableElement) string {
	if held.pair {
		return "a pair of " + iterableTypeName(held.key) + " and " + iterableTypeName(held.value)
	}
	return iterableTypeName(held.element)
}

func iterableTypeName(of ir.Type) string {
	switch of {
	case ir.Object:
		return "an object"
	case ir.Array:
		return "an array"
	case ir.Map:
		return "a map"
	case ir.Closure:
		return "a function"
	case ir.Record:
		return "a record"
	case ir.Weak:
		return "a weak reference"
	}
	return typeName(of)
}
