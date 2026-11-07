// Package fresh proves writes can't close a cycle, for the cycle finder's relaxation for fresh
// writes (docs/memory.md): over each function's control-flow graph (internal/flow), which objects a
// value can reach, and which the function made and nobody else holds.
package fresh

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// ProveWrites judges every write into a slot that can hold a reference: whether it's proven not to
// close a cycle (docs/memory.md, "Relaxing the finder for fresh writes"). The cycle finder refuses
// every mutable slot whose type can reach back to its holder; it lets one stand when every write into
// it is proven here.
//
// A cycle needs a last write, the one that closes it: before it, nothing reached itself. So it's
// enough that at every write, the value written can't reach the object written into. Then no write
// ever closes a cycle, and by induction over the program's writes none forms.
//
// # How it's proven
//
// Each function is interpreted over abstract objects, flow-sensitively over its graph, exception
// edges included. An abstract object is an allocation site in the function (a literal, new Map(), a
// result a summarized call made fresh, below), split by recency into its newest instance and all its
// older ones, or outside: anything this function didn't make, or made and let escape. The state at
// each point is what each variable may hold, what each object's fields and elements may hold, and
// which objects have escaped: been handed to a call, stored in a global or a captured variable, thrown,
// or stored into anything outside or escaped. A call can only reach what's escaped, so an object
// that hasn't escaped (confined) is held only by this function's variables and by other confined
// objects, and the function sees everything done to it.
//
// A write is proven when no object the written value can reach strongly may be the holder: the
// confined objects the value reaches aren't the holder, and if the value can reach anything outside
// or escaped, the holder is confined. That's both of the design's proofs at once. A fresh value
// reaches only confined objects, none of them the holder; a fresh holder is confined, so nothing
// outside reaches it, and only the function's own stores could have put it inside the value. A Weak
// keeps nothing, so a value reaching the holder only weakly doesn't close a cycle; what a Weak points
// at still escapes with the Weak, since whoever holds the Weak can read it.
//
// # Summaries
//
// A function returns fresh values when what it returns is confined at its exit: then only its caller
// holds it. Its summary is the confined objects its result reaches, by where they were first made,
// with outside for anything else. A call is the summary, made anew at the call site, so a value from
// new Node() or a helper that builds one is as fresh as a literal. Arguments still escape: a callee
// may keep them. Summaries are found to a fixed point over the call graph from "returns nothing",
// recursion included, and fall back to "returns something outside" if that takes too long.
func ProveWrites(program *ir.Program) []Write {
	return prepare(program, false).writeResults()
}

func prepare(program *ir.Program, ownership bool) *freshness {
	proof := &freshness{
		program:   program,
		summaries: map[int]*summary{},
		origins:   map[originKey]int{},
		direct:    directlyCalled(program),
		top:       program.HasAsync(),
		ownership: ownership,
	}
	functions := []int{}
	for index := range program.Functions {
		functions = append(functions, index)
	}
	functions = append(functions, -1)
	graphs := map[int]*flow.Function{}
	for _, function := range functions {
		graphs[function] = flow.Build(program, function)
	}
	for round := 0; ; round++ {
		proof.writes = map[writeKey]*Write{}
		proof.queries = nil
		changed := false
		for _, function := range functions {
			returned := proof.analyze(function, graphs[function])
			if function < 0 || proof.top {
				continue
			}
			if !returned.equal(proof.summaries[function]) {
				proof.summaries[function] = returned
				changed = true
			}
		}
		if !changed {
			break
		}
		if round == summaryRounds {
			// A fixed point this slow to reach isn't worth waiting for: every call's result is outside,
			// which is what a call was before summaries, and sound.
			proof.top = true
		}
	}
	return proof
}

func (proof *freshness) writeResults() []Write {
	keys := make([]writeKey, 0, len(proof.writes))
	for key := range proof.writes {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(left, right writeKey) int {
		if left.function != right.function {
			return left.function - right.function
		}
		if left.instruction != right.instruction {
			return int(left.instruction) - int(right.instruction)
		}
		return left.ordinal - right.ordinal
	})
	writes := []Write{}
	for _, key := range keys {
		writes = append(writes, *proof.writes[key])
	}
	return writes
}

// directlyCalled is every function whose every call the graph sees as an ir.Call with its arguments:
// not a closure (called through a function value), not a sort comparator (called by the runtime with
// the array's elements), and with no reference parameter a closure captures (whose cell the summary
// couldn't follow).
func directlyCalled(program *ir.Program) map[int]bool {
	direct := map[int]bool{}
	for index, function := range program.Functions {
		if function.Closure || function.RestElement != 0 {
			continue
		}
		direct[index] = true
		for _, local := range function.Parameters {
			if declared := program.Locals[local]; declared.Captured && mutable(declared.Type) {
				delete(direct, index)
			}
		}
	}
	var walk func(node reflect.Value)
	walk = func(node reflect.Value) {
		switch node.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !node.IsNil() {
				walk(node.Elem())
			}
		case reflect.Struct:
			if sorted, ok := node.Interface().(ir.ArraySort); ok {
				targets := program.ClosureTargets(sorted)
				if targets.Unknown {
					clear(direct)
				}
				for _, target := range targets.Functions {
					delete(direct, target)
				}
			}
			for index := 0; index < node.NumField(); index++ {
				walk(node.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < node.Len(); index++ {
				walk(node.Index(index))
			}
		}
	}
	for _, class := range program.Classes {
		for _, method := range class.Methods {
			delete(direct, method)
		}
	}
	walk(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		walk(reflect.ValueOf(function.Body))
	}
	return direct
}

// summaryRounds bounds the fixed point over summaries.
const summaryRounds = 32

// WriteKind is what a write writes into.
type WriteKind uint8

const (
	// WriteField is object.name = value.
	WriteField WriteKind = iota + 1
	// WriteElement is an array's element: push, an index, splice's items, fill.
	WriteElement
	// WriteMapEntry is map.set: a key and a value.
	WriteMapEntry
	// WriteSetElement is set.add.
	WriteSetElement
	// WriteUnknown is an IR node this pass doesn't know, which may write anything anywhere: it keeps
	// every slot refused.
	WriteUnknown
)

// Write is one write in the program, and whether it's proven not to close a cycle.
type Write struct {
	Kind WriteKind

	// Site is the IR node's Site: which of lowering's writes it is, to find the type written into.
	Site int

	// Name is a field's name, for WriteField.
	Name string

	// Function is the IR function the write is in, -1 for the top level.
	Function int

	// Proven is set when, everywhere the write runs, the value can't reach what it's written into.
	Proven bool

	// Why says, when it isn't proven, what the value and the holder were found to be.
	Why string
}

// object is an abstract object: outside, an allocation site's newest instance or older ones, or,
// below zero, a placeholder for what a parameter is or reaches (see analysis.placeholder).
type object int32

const outside object = 0

func newest(site int) object   { return object(2*site + 1) }
func older(site int) object    { return object(2*site + 2) }
func siteOf(object object) int { return int(object-1) / 2 }

// objects is a set of abstract objects.
type objects map[object]bool

func (objects objects) add(object object) bool {
	if objects[object] {
		return false
	}
	objects[object] = true
	return true
}

func (objects objects) sorted() []object {
	list := make([]object, 0, len(objects))
	for key := range objects {
		list = append(list, key)
	}
	slices.Sort(list)
	return list
}

// value is what an expression or a slot may hold: objects held strongly, and objects held through a
// Weak, which keeps nothing.
type value struct {
	strong, weak objects
}

func (held *value) addStrong(object object) bool {
	if held.strong == nil {
		held.strong = objects{}
	}
	return held.strong.add(object)
}

func (held *value) addWeak(object object) bool {
	if held.weak == nil {
		held.weak = objects{}
	}
	return held.weak.add(object)
}

// merge adds other's objects, and says whether anything was new.
func (held *value) merge(other value) bool {
	changed := false
	for key := range other.strong {
		changed = held.addStrong(key) || changed
	}
	for key := range other.weak {
		changed = held.addWeak(key) || changed
	}
	return changed
}

func (held value) empty() bool { return len(held.strong) == 0 && len(held.weak) == 0 }

func (held value) copy() value {
	var made value
	made.merge(held)
	return made
}

func outsideValue() value {
	return value{strong: objects{outside: true}, weak: objects{outside: true}}
}

// anyField is the key for fields a spread copied in, whatever their names: a read of any field
// reads it too. An array's elements and a map's keys and values are under elementKey.
const (
	anyField   = ""
	elementKey = "[]"
)

// state is what's known at one point of a function.
type state struct {
	// locals is what each tracked variable may hold, by IR local index; a pseudo-local past the
	// program's locals holds a value an instruction keeps for a later one (a for...of's iterable,
	// what the function returns).
	locals map[int]value

	// heap is what each confined object's fields, elements and map entries may hold.
	heap map[object]map[string]value

	// escaped is what whoever is outside this call may hold: placeholders always are. leaked is what
	// this function let escape itself, its caller's arguments among them, which its summary says.
	escaped objects
	leaked  objects

	// clobbered is set once something this function can't see into has run (a call that isn't
	// summarized), which may have written anything into its parameters.
	clobbered bool
}

func newState() *state {
	return &state{locals: map[int]value{}, heap: map[object]map[string]value{}, escaped: objects{}, leaked: objects{}}
}

func (state *state) copy() *state {
	made := newState()
	for local, held := range state.locals {
		made.locals[local] = held.copy()
	}
	for key, fields := range state.heap {
		copied := map[string]value{}
		for field, held := range fields {
			copied[field] = held.copy()
		}
		made.heap[key] = copied
	}
	for key := range state.escaped {
		made.escaped[key] = true
	}
	for key := range state.leaked {
		made.leaked[key] = true
	}
	made.clobbered = state.clobbered
	return made
}

// join adds other's facts to s, and says whether anything was new.
func (state *state) join(other *state) bool {
	changed := false
	for local, held := range other.locals {
		mine := state.locals[local]
		if mine.merge(held) {
			changed = true
		}
		state.locals[local] = mine
	}
	for key, fields := range other.heap {
		for field, held := range fields {
			if state.store(key, field, held) {
				changed = true
			}
		}
	}
	for key := range other.escaped {
		changed = state.escaped.add(key) || changed
	}
	for key := range other.leaked {
		changed = state.leaked.add(key) || changed
	}
	if other.clobbered && !state.clobbered {
		state.clobbered, changed = true, true
	}
	if changed {
		state.closeEscapes()
	}
	return changed
}

// store adds held to what `object`'s field may hold, and says whether anything was new.
func (state *state) store(object object, field string, held value) bool {
	fields := state.heap[object]
	if fields == nil {
		fields = map[string]value{}
		state.heap[object] = fields
	}
	mine := fields[field]
	changed := mine.merge(held)
	fields[field] = mine
	return changed
}

// escape marks every object a value holds, strongly or weakly, as escaped and leaked, and
// everything they hold: this function let it go where it can't see.
func (state *state) escape(held value) {
	for key := range held.strong {
		state.escapeObject(key, true)
	}
	for key := range held.weak {
		state.escapeObject(key, true)
	}
}

// expose marks a value as escaped but not leaked: stored into what a parameter reaches, which the
// caller holds, and its summary says where.
func (state *state) expose(held value) {
	for key := range held.strong {
		state.escapeObject(key, false)
	}
	for key := range held.weak {
		state.escapeObject(key, false)
	}
}

func (state *state) escapeObject(o object, leak bool) {
	queue := []object{o}
	for len(queue) > 0 {
		next := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if next == outside {
			continue
		}
		added := state.escaped.add(next)
		if leak && state.leaked.add(next) {
			added = true
		}
		if !added {
			continue
		}
		for _, held := range state.heap[next] {
			queue = append(queue, held.strong.sorted()...)
			queue = append(queue, held.weak.sorted()...)
		}
	}
}

// closeEscapes keeps what escaped and leaked objects hold escaped and leaked: a join or a merge by
// recency can give one something that hadn't been yet.
func (state *state) closeEscapes() {
	exposed := state.escaped.sorted()
	for key := range state.heap {
		if key < 0 && !state.escaped[key] {
			exposed = append(exposed, key)
		}
	}
	for _, object := range exposed {
		for _, held := range state.heap[object] {
			if state.leaked[object] {
				state.escape(held)
			} else {
				state.expose(held)
			}
		}
	}
}

// exposed reports whether whoever is outside may hold `object`.
func (state *state) exposed(object object) bool {
	return object == outside || object < 0 || state.escaped[object]
}

// storeInto is a store into `object` that lets what's stored escape as `object` has: leaked into what's leaked,
// exposed into what a caller holds.
func (state *state) storeInto(object object, field string, held value) bool {
	if object == outside {
		state.escape(held)
		return false
	}
	changed := state.store(object, field, held)
	switch {
	case state.leaked[object]:
		state.escape(held)
	case state.exposed(object):
		state.expose(held)
	}
	return changed
}

// reach is every object a value reaches through strong references alone.
func (state *state) reach(from value) objects {
	reached := objects{}
	queue := from.strong.sorted()
	for len(queue) > 0 {
		object := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if !reached.add(object) {
			continue
		}
		for _, held := range state.heap[object] {
			queue = append(queue, held.strong.sorted()...)
		}
	}
	return reached
}

// closes reports whether writing held into holder may close a cycle: whether the value may reach,
// strongly, an object the holder may be. Why says what was found when it may.
func (state *state) closes(holder value, held value) (bool, string) {
	return state.closesWith(holder, held, state.exposed)
}

// closesWith is closes, with what the value may reach outside judged by exposed: a write a call
// replays is judged after it, and what the write itself let escape doesn't count against it.
func (state *state) closesWith(holder value, held value, exposed func(object) bool) (bool, string) {
	reached := state.reach(held)
	reachesOutside := false
	for key := range reached {
		if exposed(key) {
			reachesOutside = true
		}
	}
	for _, object := range holder.strong.sorted() {
		if reachesOutside && state.exposed(object) {
			if object == outside {
				return true, "the value written reaches something this function didn't make or let escape, and what it's written into wasn't made here"
			}
			return true, "the value written reaches something this function didn't make or let escape, and what it's written into has escaped by then"
		}
		if reached[object] {
			return true, "the value written may reach what it's written into"
		}
	}
	return false, ""
}

// recent makes a site's newest instance one of its older ones, before the site makes another.
func (state *state) recent(site int) {
	from, to := newest(site), older(site)
	rename := func(held *value) {
		if held.strong[from] {
			delete(held.strong, from)
			held.addStrong(to)
		}
		if held.weak[from] {
			delete(held.weak, from)
			held.addWeak(to)
		}
	}
	for local, held := range state.locals {
		rename(&held)
		state.locals[local] = held
	}
	for _, fields := range state.heap {
		for field, held := range fields {
			rename(&held)
			fields[field] = held
		}
	}
	if fields, ok := state.heap[from]; ok {
		delete(state.heap, from)
		for field, held := range fields {
			state.store(to, field, held)
		}
	}
	if state.escaped[from] {
		delete(state.escaped, from)
		state.escaped.add(to)
	}
	if state.leaked[from] {
		delete(state.leaked, from)
		state.leaked.add(to)
	}
	state.closeEscapes()
}

// freshness is the whole program's proof.
type freshness struct {
	ownership bool
	queries   []TransferAnswer
	program   *ir.Program
	summaries map[int]*summary
	origins   map[originKey]int
	writes    map[writeKey]*Write

	// top is set once summaries are given up on: every call's result is outside, every argument
	// escapes, and nothing is deferred.
	top bool

	// direct is each function only ever called by ir.Call, with no parameter a closure captures:
	// its parameters are placeholders, and what it can't prove is judged where it's called.
	direct map[int]bool
}

// originKey is where an object was first made: a function, an instruction there, and which of the
// instruction's allocations it is.
type originKey struct {
	function    int
	instruction flow.InstructionId
	ordinal     int
}

type writeKey struct {
	function    int
	instruction flow.InstructionId
	ordinal     int
}

// siteKey is an allocation site in one function: an instruction, and the origin of what it makes
// there (its own allocation, or an object a summarized call makes).
type siteKey struct {
	instruction flow.InstructionId
	origin      int
}

// analysis is one function's interpretation.
type analysis struct {
	proof    *freshness
	function int
	graph    *flow.Function

	graphSites      map[int]bool
	classifiedSites map[int]bool
	sites           map[siteKey]int
	siteOrigin      []int

	// sitesOf is every site each instruction has made, so each is made recent before the instruction
	// runs, while nothing in it holds the old instance's name.
	sitesOf map[flow.InstructionId][]int

	// While an instruction is interpreted: the state, the instruction, how many allocations and
	// writes it has made so far, and whether writes are being judged (the last pass, at the fixed
	// point).
	state       *state
	instruction flow.InstructionId
	allocations int
	writeCount  int
	judging     bool

	// returnedLocal is the pseudo-local past the program's locals that holds what the function
	// returns, and iterables those after it, each the iterable of a for...of.
	returnedLocal int
	iterables     map[*ir.Statement]int

	// placeholders is set for a function in proof.direct: each parameter is a placeholder, and so is
	// what each field read from one reaches. terms names each, by its object (below zero).
	placeholders bool
	terms        []term
	termObjects  map[term]object

	// deferred is each write this function couldn't prove that its callers will judge: the write, and
	// the holder and value it may have, joined over every time it runs.
	deferred map[writeKey]*deferral
}

// term is a placeholder: parameter param itself (level 0), what it holds in field (level 1), or
// everything below that (level 2).
type term struct {
	param int
	field string
	level int
}

// deferral is a write left to the callers to judge, with what it writes into and what it writes.
type deferral struct {
	write          Write
	holder, stored value
}

// placeholder is the object for a term, made the first time it's met.
func (analysis *analysis) placeholder(term term) object {
	if object, ok := analysis.termObjects[term]; ok {
		return object
	}
	analysis.terms = append(analysis.terms, term)
	object := object(-len(analysis.terms))
	analysis.termObjects[term] = object
	return object
}

// below is the placeholder for what a placeholder holds in a field: a parameter's field, then
// everything below that.
func (analysis *analysis) below(object object, field string) object {
	t := analysis.termOf(object)
	switch t.level {
	case 0:
		return analysis.placeholder(term{param: t.param, field: field, level: 1})
	case 1:
		return analysis.placeholder(term{param: t.param, field: t.field, level: 2})
	}
	return object
}

// termOf is a placeholder's term.
func (analysis *analysis) termOf(object object) term {
	return analysis.terms[-int(object)-1]
}

// load is what a field of the objects a value holds strongly may hold.
func (analysis *analysis) load(from value, field string) value {
	var loaded value
	for key := range from.strong {
		fields := analysis.state.heap[key]
		loaded.merge(fields[field])
		loaded.merge(fields[anyField])
		switch {
		case key < 0:
			// What the argument holds there, or anything below it.
			reached := analysis.below(key, field)
			loaded.addStrong(reached)
			loaded.addWeak(reached)
			if analysis.state.clobbered {
				loaded.merge(outsideValue())
			}
		case analysis.state.exposed(key):
			// Whoever is outside may have written anything outside there.
			loaded.merge(outsideValue())
		}
	}
	return loaded
}

// everything is what any field, element or entry of the objects a value holds strongly may hold.
func (analysis *analysis) everything(from value) value {
	var loaded value
	for key := range from.strong {
		for _, held := range analysis.state.heap[key] {
			loaded.merge(held)
		}
		switch {
		case key < 0:
			reached := analysis.below(key, anyField)
			loaded.addStrong(reached)
			loaded.addWeak(reached)
			if analysis.state.clobbered {
				loaded.merge(outsideValue())
			}
		case analysis.state.exposed(key):
			loaded.merge(outsideValue())
		}
	}
	return loaded
}

// analyze interprets a function to its fixed point, judges its writes, and returns its summary.
func (proof *freshness) analyze(function int, graph *flow.Function) *summary {
	analysis := &analysis{
		proof:       proof,
		function:    function,
		graph:       graph,
		sites:       map[siteKey]int{},
		sitesOf:     map[flow.InstructionId][]int{},
		iterables:   map[*ir.Statement]int{},
		termObjects: map[term]object{},
		deferred:    map[writeKey]*deferral{},
	}
	if proof.ownership {
		analysis.graphSites = map[int]bool{}
		analysis.classifiedSites = map[int]bool{}
	}
	analysis.returnedLocal = len(proof.program.Locals)
	analysis.placeholders = function >= 0 && proof.direct[function] && !proof.top
	entry := newState()
	if function >= 0 {
		for index, local := range proof.program.Functions[function].Parameters {
			declared := proof.program.Locals[local]
			if !mutable(declared.Type) || declared.Captured {
				continue
			}
			if analysis.placeholders {
				parameter := analysis.placeholder(term{param: index})
				entry.locals[local] = value{strong: objects{parameter: true}, weak: objects{parameter: true}}
			} else {
				entry.locals[local] = outsideValue()
			}
		}
	}
	in := map[flow.BlockId]*state{graph.Entry: entry}
	position := map[flow.BlockId]int{}
	for index, block := range graph.Blocks {
		position[block.Id] = index
	}
	pending := map[flow.BlockId]bool{graph.Entry: true}
	for len(pending) > 0 {
		// The earliest pending block in reverse postorder, so a loop's body is walked before its exit.
		var block *flow.BasicBlock
		for id := range pending {
			candidate, _ := graph.Block(id)
			if block == nil || position[candidate.Id] < position[block.Id] {
				block = candidate
			}
		}
		delete(pending, block.Id)
		before, after := analysis.block(block, in[block.Id])
		analysis.flowOut(block, before, after, func(successor flow.BlockId, out *state) {
			if in[successor] == nil {
				in[successor] = out.copy()
				pending[successor] = true
			} else if in[successor].join(out) {
				pending[successor] = true
			}
		})
	}
	// At the fixed point, once more, judging every write, and joining the states the function
	// leaves by.
	analysis.judging = true
	exit := newState()
	exited := false
	for _, block := range graph.Blocks {
		if in[block.Id] == nil {
			continue
		}
		_, after := analysis.block(block, in[block.Id])
		switch block.Terminal.(type) {
		case *flow.Return, *flow.Throw:
			// Leaving, either way: what it did on the way is its callers' to know.
			exit.join(after)
			exited = true
		}
	}
	if function < 0 || !exited {
		// Never leaving, its callers never judge what it deferred.
		for key, deferred := range analysis.deferred {
			analysis.proof.unproven(
				key,
				deferred.write,
				deferred.write.Why+", in a function that never returns or throws",
			)
		}
		return nil
	}
	return analysis.summarize(exit)
}

// flowOut hands a block's state to each block after it: a handler takes either the state before
// the instruction that threw or after it, since a call that throws never gives its variable a value.
func (analysis *analysis) flowOut(
	block *flow.BasicBlock,
	before *state,
	after *state,
	visit func(flow.BlockId, *state),
) {
	if throws, ok := block.Terminal.(*flow.MayThrow); ok {
		visit(throws.Next, after)
		handler := before.copy()
		handler.join(after)
		visit(throws.Handler, handler)
		return
	}
	flow.EachSuccessor(block.Terminal, func(successor flow.BlockId) { visit(successor, after) })
}

// block interprets a block from its state on entry: the state before its last instruction, and after.
func (analysis *analysis) block(block *flow.BasicBlock, entry *state) (*state, *state) {
	analysis.state = entry.copy()
	before := analysis.state
	for index, id := range block.Instructions {
		if index == len(block.Instructions)-1 {
			before = analysis.state.copy()
		}
		analysis.run(id)
	}
	if flow.Suspends(block.Terminal) {
		// The environment owns these values, but confinement across queued work is unproved.
		for _, held := range analysis.state.locals {
			analysis.state.escape(held)
		}
		analysis.state.clobbered = true
	}
	return before, analysis.state
}

// run interprets one instruction.
func (analysis *analysis) run(id flow.InstructionId) {
	instruction := analysis.graph.Instructions[id]
	analysis.instruction, analysis.allocations, analysis.writeCount = id, 0, 0
	for _, site := range analysis.sitesOf[id] {
		analysis.state.recent(site)
	}
	switch statement := (*instruction.At).(type) {
	case ir.SetProperty:
		holder := analysis.value(statement.Object)
		held := analysis.value(statement.Value)
		analysis.write(WriteField, statement.Site, statement.Name, holder, held, statement.Name)
	case ir.SetIndex:
		holder := analysis.value(statement.Array)
		analysis.value(statement.Index)
		held := analysis.value(statement.Value)
		analysis.write(WriteElement, statement.Site, "", holder, held, elementKey)
	case ir.AllocateEnvironment:
		// Empty captured storage allocates no user value and changes no existing slot.
	case ir.Declare:
		analysis.define(statement.Local, analysis.value(statement.Value))
	case ir.Assign:
		analysis.define(statement.Local, analysis.value(statement.Value))
	case ir.Return:
		returned := analysis.value(statement.Value)
		held := analysis.state.locals[analysis.returnedLocal]
		held.merge(returned)
		analysis.state.locals[analysis.returnedLocal] = held
	case ir.Throw:
		analysis.state.escape(analysis.value(statement.Value))
	case ir.ForOf:
		iterable := analysis.iterable(instruction.At)
		if instruction.Part == 0 {
			analysis.state.locals[iterable] = analysis.value(statement.Iterable)
			break
		}
		elements := analysis.load(analysis.state.locals[iterable], elementKey)
		if statement.Pattern == nil {
			analysis.define(statement.Local, elements)
			break
		}
		for _, binding := range statement.Pattern {
			// A pair a map gives, or a tuple of the array's: what the element holds there.
			bound := elements.copy()
			bound.merge(analysis.load(elements, binding.Field))
			analysis.define(binding.Local, bound)
		}
	case ir.Try:
		// The catch begins by taking the error, which came from wherever it was thrown.
		if statement.CatchLocal >= 0 {
			analysis.define(statement.CatchLocal, outsideValue())
		}
	case ir.WriteLine, ir.Evaluate, ir.Panic, ir.If, ir.Loop, ir.Switch:
		// What they evaluate: a value written or discarded, a panic's message, a condition, a switch's
		// value or a case's test.
		analysis.value(instruction.Expression)
	default:
		// A statement added to the IR after this was written: it may write anything anywhere.
		analysis.unknown(statement)
		analysis.value(instruction.Expression)
	}
}

// iterable is the pseudo-local holding what a for...of iterates, from its first part to its second.
func (analysis *analysis) iterable(at *ir.Statement) int {
	local, ok := analysis.iterables[at]
	if !ok {
		local = analysis.returnedLocal + 1 + len(analysis.iterables)
		analysis.iterables[at] = local
	}
	return local
}

// define gives a variable a value: a tracked one holds it from here on, and one a call can reach (a
// global, a captured variable) lets it escape.
func (analysis *analysis) define(local int, held value) {
	declared := analysis.proof.program.Locals[local]
	if !mutable(declared.Type) {
		return
	}
	if declared.Global || declared.Captured {
		analysis.state.escape(held)
		return
	}
	analysis.state.locals[local] = held
}

// write judges a write of held into the field (or elements) of holder, and makes it.
func (analysis *analysis) write(kind WriteKind, site int, name string, holder value, held value, field string) {
	key := writeKey{function: analysis.function, instruction: analysis.instruction, ordinal: analysis.writeCount}
	analysis.writeCount++
	if analysis.judging {
		write := Write{Kind: kind, Site: site, Name: name, Function: analysis.function, Proven: true}
		analysis.proof.record(key, write)
		analysis.judge(key, write, holder, held, "", analysis.state.exposed)
	}
	for key := range holder.strong {
		analysis.state.storeInto(key, field, held)
	}
}

// judge decides a write here, or defers it to this function's callers: one this function can't
// prove alone may be proven where it's called, with what the caller knows of its arguments.
func (analysis *analysis) judge(
	key writeKey,
	write Write,
	holder value,
	held value,
	where string,
	exposed func(object) bool,
) {
	closes, why := analysis.state.closesWith(holder, held, exposed)
	if !closes {
		return
	}
	if analysis.placeholders {
		deferred := analysis.deferred[key]
		if deferred == nil {
			write.Why = why + where
			deferred = &deferral{write: write}
			analysis.deferred[key] = deferred
		}
		deferred.holder.merge(holder)
		deferred.stored.merge(held)
		return
	}
	analysis.proof.unproven(key, write, why+where)
}

// record notes a write the first time it's met in a round, proven until something says otherwise.
func (proof *freshness) record(key writeKey, write Write) *Write {
	recorded := proof.writes[key]
	if recorded == nil {
		recorded = &write
		recorded.Proven, recorded.Why = true, ""
		proof.writes[key] = recorded
	}
	return recorded
}

// unproven marks a write as one that may close a cycle, keeping the first reason found.
func (proof *freshness) unproven(key writeKey, write Write, why string) {
	recorded := proof.record(key, write)
	if recorded.Proven {
		recorded.Proven, recorded.Why = false, why
	}
}

// unknown records a write this pass can't judge: an IR node it doesn't know.
func (analysis *analysis) unknown(node any) {
	key := writeKey{function: analysis.function, instruction: analysis.instruction, ordinal: analysis.writeCount}
	analysis.writeCount++
	if analysis.judging {
		analysis.proof.writes[key] = &Write{
			Kind:     WriteUnknown,
			Function: analysis.function,
			Why:      fmt.Sprintf("%T is a node the cycle finder doesn't know", node),
		}
	}
}

// allocate is a new object made by this instruction: its next allocation's site, newest instance.
func (analysis *analysis) allocate() object {
	origin := analysis.proof.origin(
		originKey{function: analysis.function, instruction: analysis.instruction, ordinal: analysis.allocations},
	)
	analysis.allocations++
	return newest(analysis.site(origin))
}

// site is this instruction's site for an origin, made the first time it's met.
func (analysis *analysis) site(origin int) int {
	key := siteKey{instruction: analysis.instruction, origin: origin}
	site, ok := analysis.sites[key]
	if !ok {
		site = len(analysis.siteOrigin)
		analysis.sites[key] = site
		analysis.siteOrigin = append(analysis.siteOrigin, origin)
		analysis.sitesOf[analysis.instruction] = append(analysis.sitesOf[analysis.instruction], site)
	}
	return site
}

func (proof *freshness) origin(key originKey) int {
	origin, ok := proof.origins[key]
	if !ok {
		origin = len(proof.origins)
		proof.origins[key] = origin
	}
	return origin
}

// fresh is a new object holding contents under field.
func (analysis *analysis) fresh(field string, contents value) value {
	made := analysis.allocate()
	if !contents.empty() {
		analysis.state.store(made, field, contents)
	}
	return value{strong: objects{made: true}}
}

// call is what a call does to what it's handed: everything escapes, and its result is outside.
func (analysis *analysis) call(operands []value, returns ir.Type) value {
	for _, operand := range operands {
		analysis.state.escape(operand)
	}
	// It may write anything it can reach into the parameters.
	analysis.state.clobbered = true
	if !mutable(returns) {
		return value{}
	}
	return outsideValue()
}

// value interprets an expression: its effects, and what its value may hold.
func (analysis *analysis) value(expression ir.Expression) value {
	if expression == nil {
		return value{}
	}
	if held, known := analysis.libraryLanguage(expression); known {
		return held
	}
	switch expression := expression.(type) {
	case ir.Await:
		return analysis.call([]value{analysis.value(expression.Value)}, expression.Of)
	case ir.Yield:
		// The consumer runs before the body goes on, and may do anything with what was handed over.
		return analysis.call([]value{analysis.value(expression.Value)}, 0)
	case ir.GeneratorNext:
		// The body runs, holding whatever the frame holds, and may hand back any of it.
		return analysis.call([]value{analysis.value(expression.Generator)}, expression.Of)
	case ir.GeneratorDone:
		analysis.value(expression.Generator)
		return value{}
	case ir.IterableOf:
		// The generator itself, never a fresh object: no reuse or region plan may take its storage
		// as an object's (Memory, on #qk8rztp S2c-1).
		return analysis.value(expression.Value)
	case ir.IsGenerator:
		analysis.value(expression.Value)
		return value{}
	case ir.IsCollectionIterator:
		analysis.value(expression.Value)
		return value{}
	case ir.PromiseValue:
		return analysis.call([]value{analysis.value(expression.Value)}, ir.Promise)
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.JSONNull, ir.Null:
		return value{}
	case ir.Read:
		declared := analysis.proof.program.Locals[expression.Local]
		if !mutable(declared.Type) {
			return value{}
		}
		if declared.Global || declared.Captured {
			return outsideValue()
		}
		held, ok := analysis.state.locals[expression.Local]
		if !ok {
			// Never given a value on any path here: a read the checker let through that this pass can't
			// place, so it's taken as the most it could be.
			return outsideValue()
		}
		return held.copy()
	case ir.Void:
		analysis.value(expression.Value)
		return value{}
	case ir.Comma:
		analysis.value(expression.Left)
		return analysis.value(expression.Right)
	case ir.Effects:
		for _, statement := range expression.Body {
			analysis.expressionEffect(statement)
		}
		return analysis.value(expression.Result)
	case ir.LogicalAssignment:
		result := analysis.value(expression.Read)
		analysis.value(expression.Key)
		before := analysis.state.copy()
		switch write := expression.Write.(type) {
		case ir.Assign:
			held := analysis.value(write.Value)
			result.merge(held)
			analysis.define(write.Local, held)
		case ir.SetProperty:
			holder := analysis.value(write.Object)
			held := analysis.value(write.Value)
			result.merge(held)
			if expression.Key != nil {
				for _, name := range expression.KeyNames {
					analysis.write(WriteField, write.Site, name, holder, held, name)
				}
			} else {
				analysis.write(WriteField, write.Site, write.Name, holder, held, write.Name)
			}
		case ir.SetIndex:
			holder := analysis.value(write.Array)
			analysis.value(write.Index)
			held := analysis.value(write.Value)
			result.merge(held)
			analysis.write(WriteElement, write.Site, "", holder, held, elementKey)
		}
		analysis.state.join(before)
		return result
	case ir.Truthy:
		analysis.value(expression.Value)
		return value{}
	case ir.Logical:
		result := analysis.value(expression.Left)
		result.merge(analysis.value(expression.Right))
		return result
	case ir.Unary:
		analysis.value(expression.Operand)
		return value{}
	case ir.Binary:
		analysis.value(expression.Left)
		analysis.value(expression.Right)
		return value{}
	case ir.IsNull:
		analysis.value(expression.Value)
		return value{}
	case ir.RegExpNew:
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		// Patterns are compiled constants; the runtime object holds only immutable strings.
		return analysis.fresh(anyField, value{})
	case ir.RegExpCall:
		return analysis.regexCall(expression)
	case ir.RegExpProperty:
		analysis.value(expression.Array)
		if mutable(expression.Type()) {
			return outsideValue()
		}
		return value{}
	case ir.RegExpGroup:
		analysis.value(expression.Object)
		if mutable(expression.Type()) {
			return outsideValue()
		}
		return value{}
	case ir.JSONEncode:
		analysis.value(expression.Value)
		return value{}
	case ir.JSONStringify:
		// Lowering excludes toJSON and replacer callbacks; serialization only reads values.
		analysis.value(expression.Value)
		analysis.value(expression.Replacer)
		analysis.value(expression.Space)
		return value{}
	case ir.NumberToString:
		analysis.value(expression.Value)
		return value{}
	case ir.BooleanToString:
		analysis.value(expression.Value)
		return value{}
	case ir.Concat:
		for _, part := range expression.Parts {
			analysis.value(part)
		}
		return value{}
	case ir.Conditional:
		analysis.value(expression.Condition)
		result := analysis.value(expression.WhenTrue)
		result.merge(analysis.value(expression.WhenNot))
		return result
	case ir.Coalesce:
		result := analysis.value(expression.Value)
		result.merge(analysis.value(expression.Fallback))
		analysis.value(expression.Panic)
		return result
	case ir.Box:
		return analysis.value(expression.Value)
	case ir.Narrow:
		return analysis.value(expression.Value)
	case ir.PlaceholderUse:
		// A checked use forwards the value; it cannot create a reference or a write.
		return analysis.value(expression.Value)
	case ir.Unwrap:
		return analysis.value(expression.Value)
	case ir.Defined:
		return analysis.value(expression.Value)
	case ir.MaybeOf:
		return analysis.value(expression.Value)
	case ir.CheckedCast:
		result := analysis.value(expression.Value)
		for _, allowed := range expression.Allowed {
			analysis.value(allowed)
		}
		return result
	case ir.WeakOf:
		// A Weak keeps nothing: what it points at is held weakly.
		held := analysis.value(expression.Value)
		var weak value
		for key := range held.strong {
			weak.addWeak(key)
		}
		for key := range held.weak {
			weak.addWeak(key)
		}
		return weak
	case ir.WeakTarget:
		held := analysis.value(expression.Value)
		var target value
		for key := range held.weak {
			target.addStrong(key)
		}
		for key := range held.strong {
			target.addStrong(key)
		}
		return target
	case ir.Length:
		analysis.value(expression.Array)
		return value{}
	case ir.MathCall:
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		return value{}
	case ir.RecordCoalesce:
		holder := analysis.value(expression.Record)
		analysis.value(expression.Key)
		result := analysis.value(expression.Value)
		analysis.write(WriteMapEntry, expression.Site, "", holder, result, elementKey)
		result.merge(analysis.load(holder, elementKey))
		return result
	case ir.RecordCall:
		return analysis.recordCall(expression)
	case ir.RecordLiteral:
		var entries value
		if expression.Spread != nil {
			entries.merge(analysis.load(analysis.value(expression.Spread), elementKey))
		}
		for _, entry := range expression.Entries {
			analysis.value(entry.Key)
			entries.merge(analysis.value(entry.Value))
		}
		return analysis.fresh(elementKey, entries)
	case ir.ObjectCall:
		return analysis.objectCall(expression)
	case ir.NumberCall:
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		return value{}
	case ir.ToFixed:
		analysis.value(expression.Value)
		analysis.value(expression.Digits)
		return value{}
	case ir.NumberFormat:
		analysis.value(expression.Value)
		analysis.value(expression.Argument)
		return value{}
	case ir.IsUndefined:
		analysis.value(expression.Value)
		return value{}
	case ir.MaybeToString:
		analysis.value(expression.Value)
		return value{}
	case ir.ArrayIsArray:
		analysis.value(expression.Value)
		return value{}
	case ir.RecordStorage:
		analysis.value(expression.Value)
		return value{}
	case ir.TypeOf:
		analysis.value(expression.Value)
		return value{}
	case ir.UnionToString:
		analysis.value(expression.Value)
		return value{}
	case ir.UnknownKind:
		analysis.value(expression.Value)
		return value{}
	case ir.OwnFunction:
		analysis.value(expression.Object)
		return value{}
	case ir.ErrorToString:
		analysis.value(expression.Value)
		return value{}
	case ir.ArrayOfScalars:
		analysis.value(expression.Array)
		return value{}
	case ir.JoinEnter:
		analysis.value(expression.Array)
		return value{}
	case ir.JoinLeave:
		return value{}
	case ir.StringLength:
		analysis.value(expression.Value)
		return value{}
	case ir.CharCodeAt:
		analysis.value(expression.Value)
		analysis.value(expression.Index)
		return value{}
	case ir.Utf8Length:
		analysis.value(expression.Text)
		return value{}
	case ir.Utf8At:
		analysis.value(expression.Text)
		analysis.value(expression.Index)
		return value{}
	case ir.StringFromCodes:
		// A new string, from numbers: nothing it reaches can hold a holder.
		for _, code := range expression.Codes {
			analysis.value(code)
		}
		if expression.Spread != nil {
			analysis.value(expression.Spread)
		}
		return value{}
	case ir.Trim:
		analysis.value(expression.Value)
		return value{}
	case ir.StringIndex:
		analysis.value(expression.Value)
		analysis.value(expression.Index)
		return value{}
	case ir.StringCall:
		analysis.value(expression.Value)
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		if expression.Type() == ir.Array {
			// split: a new array of strings.
			return analysis.fresh(elementKey, value{})
		}
		return value{}
	case ir.CodePoints:
		analysis.value(expression.Value)
		return analysis.fresh(elementKey, value{})
	case ir.ProgramArguments:
		return analysis.fresh(elementKey, value{})
	case ir.JSONDecode:
		analysis.value(expression.Text)
		// The decoded graph is fresh, but nesting is conservatively treated as outside.
		return outsideValue()
	case ir.ReadTextFile:
		analysis.value(expression.Path)
		return analysis.fresh(anyField, value{})
	case ir.WriteTextFile:
		analysis.value(expression.Path)
		analysis.value(expression.Text)
		return analysis.fresh(anyField, value{})
	case ir.MakeError:
		analysis.value(expression.Message)
		if expression.Name != nil {
			analysis.value(expression.Name)
		}
		return analysis.fresh(anyField, value{})
	case ir.ObjectLiteral:
		var copied value
		if expression.Spread != nil {
			source := analysis.value(expression.Spread)
			if expression.NoReuse {
				analysis.state.escape(source)
			}
			copied = analysis.everything(source)
		}
		all := append(append([]ir.Field{}, expression.Fields...), expression.Empty...)
		fields := make([]value, len(all))
		for index, field := range all {
			fields[index] = analysis.value(field.Value)
		}
		made := analysis.fresh(anyField, copied)
		analysis.tagOwnershipGraph(made, expression.GraphTypes, expression.ProgramRegion)
		for index, field := range all {
			if !fields[index].empty() {
				analysis.state.store(newestOf(made), field.Name, fields[index])
			}
		}
		return made
	case ir.ArrayLiteral:
		var elements value
		for index, element := range expression.Elements {
			held := analysis.value(element)
			if index < len(expression.Spread) && expression.Spread[index] {
				held = analysis.load(held, elementKey)
			}
			elements.merge(held)
		}
		made := analysis.fresh(elementKey, elements)
		analysis.tagOwnershipGraph(made, expression.GraphTypes, expression.ProgramRegion)
		return made
	case ir.StringIterator:
		return analysis.fresh(anyField, analysis.value(expression.Value))
	case ir.CollectionIterator:
		return analysis.fresh(anyField, analysis.value(expression.Collection))
	case ir.MapNew:
		var entries value
		for _, entry := range expression.Entries {
			entries.merge(analysis.value(entry[0]))
			entries.merge(analysis.value(entry[1]))
		}
		if expression.Pairs != nil {
			// Each pair's key and value: a tuple's fields, or another map's entries.
			pairs := analysis.load(analysis.value(expression.Pairs), elementKey)
			entries.merge(pairs)
			entries.merge(analysis.everything(pairs))
		}
		return analysis.fresh(elementKey, entries)
	case ir.MapKeys:
		return analysis.fresh(elementKey, analysis.load(analysis.value(expression.Map), elementKey))
	case ir.MapValues:
		return analysis.fresh(elementKey, analysis.load(analysis.value(expression.Map), elementKey))
	case ir.MapClear:
		analysis.value(expression.Map)
		return value{}
	case ir.MapForEach:
		return analysis.call([]value{analysis.value(expression.Map), analysis.value(expression.Callback)}, 0)
	case ir.ReadDirectory:
		analysis.value(expression.Path)
		return analysis.fresh(anyField, analysis.fresh(elementKey, value{}))
	case ir.ProcessCall:
		return analysis.call(analysis.operands(expression), expression.Type())
	case ir.NodeHostCall:
		return analysis.call(analysis.operands(expression), expression.Type())
	case ir.NodeFSFile:
		return analysis.nodeFSFile(expression)
	case ir.FileStatus:
		analysis.value(expression.Path)
		return analysis.fresh(anyField, value{})
	case ir.SetNew:
		return analysis.fresh(elementKey, analysis.load(analysis.value(expression.Values), elementKey))
	case ir.SetValues:
		return analysis.fresh(elementKey, analysis.load(analysis.value(expression.Set), elementKey))
	case ir.MapEntries:
		contents := analysis.load(analysis.value(expression.Map), elementKey)
		pairs := analysis.fresh(anyField, contents)
		return analysis.fresh(elementKey, pairs)
	case ir.ArraySlice:
		elements := analysis.load(analysis.value(expression.Array), elementKey)
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		return analysis.fresh(elementKey, elements)
	case ir.ArrayConcat:
		elements := analysis.load(analysis.value(expression.Array), elementKey)
		for _, other := range expression.Others {
			elements.merge(analysis.load(analysis.value(other), elementKey))
		}
		return analysis.fresh(elementKey, elements)
	case ir.HasProperty:
		analysis.value(expression.Object)
		return value{}
	case ir.DynamicProperty:
		return analysis.load(analysis.value(expression.Object), expression.Name)
	case ir.DynamicKeyedProperty:
		// It may call a class getter by whatever name the key holds, so it's taken as a call, which
		// lets its operands escape and its result be anything outside.
		return analysis.call(analysis.operands(expression), expression.Type())
	case ir.HasAccessor:
		analysis.value(expression.Object)
		return value{}
	case ir.Property:
		return analysis.load(analysis.value(expression.Object), expression.Name)
	case ir.ArrayIndex:
		array := analysis.value(expression.Array)
		analysis.value(expression.Index)
		return analysis.load(array, elementKey)
	case ir.ArrayPop:
		return analysis.load(analysis.value(expression.Array), elementKey)
	case ir.ArraySearch:
		analysis.value(expression.Array)
		analysis.value(expression.Value)
		if expression.From != nil {
			analysis.value(expression.From)
		}
		return value{}
	case ir.ArrayJoin:
		analysis.value(expression.Array)
		analysis.value(expression.Separator)
		return value{}
	case ir.ArrayReverse:
		return analysis.value(expression.Array)
	case ir.ArrayPush:
		holder := analysis.value(expression.Array)
		held := analysis.value(expression.Value)
		analysis.write(WriteElement, expression.Site, "", holder, held, elementKey)
		return value{}
	case ir.ArrayFill:
		if expression.Array == nil {
			analysis.value(expression.Length)
			held := analysis.value(expression.Value)
			analysis.value(expression.Start)
			analysis.value(expression.End)
			return analysis.fresh(elementKey, held)
		}
		holder := analysis.value(expression.Array)
		held := analysis.value(expression.Value)
		analysis.value(expression.Start)
		analysis.value(expression.End)
		analysis.write(WriteElement, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.ArraySplice:
		holder := analysis.value(expression.Array)
		analysis.value(expression.Start)
		analysis.value(expression.Count)
		var items value
		for _, item := range expression.Items {
			items.merge(analysis.value(item))
		}
		removed := analysis.load(holder, elementKey)
		if len(expression.Items) > 0 {
			analysis.write(WriteElement, expression.Site, "", holder, items, elementKey)
		}
		return analysis.fresh(elementKey, removed)
	case ir.MapGet:
		object := analysis.value(expression.Map)
		analysis.value(expression.Key)
		return analysis.load(object, elementKey)
	case ir.MapSet:
		holder := analysis.value(expression.Map)
		held := analysis.value(expression.Key)
		held.merge(analysis.value(expression.Value))
		analysis.write(WriteMapEntry, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.MapHas:
		analysis.value(expression.Map)
		analysis.value(expression.Key)
		return value{}
	case ir.MapDelete:
		analysis.value(expression.Map)
		analysis.value(expression.Key)
		return value{}
	case ir.MapSize:
		analysis.value(expression.Map)
		return value{}
	case ir.HasOwn:
		analysis.value(expression.Object)
		analysis.value(expression.Key)
		return value{}
	case ir.SetAdd:
		holder := analysis.value(expression.Set)
		held := analysis.value(expression.Value)
		analysis.write(WriteSetElement, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.MakeClosure:
		// What a function value holds is its cells, and what's in a cell has escaped.
		return outsideValue()
	case ir.InstanceOf:
		analysis.value(expression.Value)
		return value{}
	case ir.Call:
		operands := []value{}
		for _, argument := range expression.Arguments {
			operands = append(operands, analysis.value(argument))
		}
		if analysis.proof.top ||
			analysis.proof.program.CallExpandsArguments(expression) ||
			expression.Returns == ir.Generator {
			// A generator's call keeps every argument in its frame for steps to come, which its
			// summary, made for a call that returns, can't say.
			return analysis.call(operands, expression.Returns)
		}
		before := analysis.state.copy()
		joined := before.copy()
		var result value
		for _, target := range analysis.proof.program.CallTargets(expression) {
			analysis.state = before.copy()
			if !analysis.proof.direct[target] {
				// Its parameters are outside to it, so it may keep them anywhere.
				analysis.call(operands, 0)
			}
			result.merge(analysis.made(target, analysis.proof.summaries[target], operands))
			joined.join(analysis.state)
		}
		analysis.state = joined
		if !mutable(expression.Returns) {
			return value{}
		}
		return result
	case ir.CallClosure:
		// Cells and structural receivers are outside this summary's model. The
		// union for any bounded target set, and for Unknown, is an arbitrary call.
		return analysis.call(analysis.operands(expression), expression.Returns)
	case ir.ParallelMap:
		items := analysis.value(expression.Items)
		work := analysis.value(expression.Work)
		if analysis.judging && analysis.proof.ownership && expression.Moved {
			analysis.proof.queries = append(analysis.proof.queries, analysis.transferAnswer(expression, items))
		}
		// Tasks have a separate effect proof; results may alias shared inputs.
		return analysis.call([]value{items, work}, ir.Array)
	case ir.ArrayMap:
		return analysis.call([]value{analysis.value(expression.Array), analysis.value(expression.Callback)}, ir.Array)
	case ir.ArrayVisit:
		return analysis.call(
			[]value{analysis.value(expression.Array), analysis.value(expression.Callback)},
			expression.Type(),
		)
	case ir.ArrayReduce:
		return analysis.call([]value{
			analysis.value(expression.Array),
			analysis.value(expression.Callback),
			analysis.value(expression.Initial),
		}, expression.Result)
	case ir.ArrayFrom:
		return analysis.call([]value{analysis.value(expression.Length), analysis.value(expression.Callback)}, ir.Array)
	case ir.ArraySort:
		// The comparator, a function or a function value, is handed the elements.
		array := analysis.value(expression.Array)
		analysis.call([]value{array, analysis.value(expression.Callback)}, 0)
		return array
	}
	// A node added to the IR after this was written: it may write anything anywhere, so it keeps every
	// slot refused, and it's taken as a call.
	analysis.unknown(expression)
	return analysis.call(analysis.operands(expression), expression.Type())
}

// operands interprets every expression an IR node holds.
func (analysis *analysis) operands(expression ir.Expression) []value {
	var operands []value
	for _, operand := range expressionsIn(expression) {
		operands = append(operands, analysis.value(operand))
	}
	return operands
}

// expressionsIn is every expression an IR node holds directly, in the order its fields are declared.
func expressionsIn(expression ir.Expression) []ir.Expression {
	var found []ir.Expression
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if inner, ok := value.Interface().(ir.Expression); ok {
				found = append(found, inner)
				return
			}
			walk(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	walk(reflect.ValueOf(expression))
	return found
}

func newestOf(made value) object {
	for key := range made.strong {
		return key
	}
	return outside
}

// made is a summarized call, replayed here: its fresh objects made anew, each the newest instance of
// this instruction's site for where it was first made; what it stored into its parameters stored
// into the arguments; what it let escape of them let escape; and the writes it left to its callers
// judged with what's known here. Its result is what it returns.
func (analysis *analysis) made(function int, callee *summary, arguments []value) value {
	if callee == nil {
		// Returns nothing yet, on the way to the fixed point.
		return value{}
	}
	instances := map[int]object{}
	for _, origin := range callee.origins() {
		instances[origin] = newest(analysis.site(origin))
	}
	// What had escaped before the call: while the callee runs, only code it can't see into could
	// tie what it was given to what had, and then anything escaped counts.
	before := objects{}
	for key := range analysis.state.escaped {
		before.add(key)
	}
	translate := func(held summaryValue) value {
		var made value
		for _, origin := range held.strong {
			if origin < 0 {
				made.addStrong(outside)
			} else {
				made.addStrong(instances[origin])
			}
		}
		for _, origin := range held.weak {
			if origin < 0 {
				made.addWeak(outside)
			} else {
				made.addWeak(instances[origin])
			}
		}
		for _, strongTerm := range held.strongTerms {
			argument := analysis.argument(strongTerm, arguments)
			made.merge(argument)
		}
		for _, weakTerm := range held.weakTerms {
			argument := analysis.argument(weakTerm, arguments)
			for key := range argument.strong {
				made.addWeak(key)
			}
			for key := range argument.weak {
				made.addWeak(key)
			}
		}
		return made
	}
	// Until nothing more is added: a store through one argument can change what another reaches,
	// when the caller passed the same object for both.
	for {
		escaped, leaked := len(analysis.state.escaped), len(analysis.state.leaked)
		changed := false
		for _, origin := range callee.origins() {
			for field, held := range callee.fields[origin] {
				changed = analysis.state.storeInto(instances[origin], field, translate(held)) || changed
			}
		}
		for _, t := range callee.edgeTerms() {
			holders := translate(summaryValue{strongTerms: []term{t}})
			for field, held := range callee.edges[t] {
				stored := translate(held)
				for _, object := range holders.strong.sorted() {
					changed = analysis.state.storeInto(object, field, stored) || changed
				}
			}
		}
		for _, leak := range callee.leaks {
			analysis.state.escape(translate(summaryValue{strongTerms: []term{leak}}))
		}
		if !changed && escaped == len(analysis.state.escaped) && leaked == len(analysis.state.leaked) {
			break
		}
	}
	if callee.clobbers {
		analysis.state.clobbered = true
	}
	if analysis.judging {
		where := ", where " +
			analysis.proof.functionName(function) +
			" is called from " +
			analysis.proof.functionName(analysis.function)
		for _, key := range callee.deferredKeys() {
			deferred := callee.deferred[key]
			exposed := func(object object) bool {
				return object == outside ||
					object < 0 ||
					before[object] ||
					(callee.clobbers && analysis.state.exposed(object))
			}
			analysis.judge(key, deferred.write, translate(deferred.holder), translate(deferred.stored), where, exposed)
		}
	}
	return translate(callee.result)
}

// argument is what a placeholder of a callee's stands for here: the argument, or everything the
// argument reaches through the field, strongly or weakly.
func (analysis *analysis) argument(term term, arguments []value) value {
	if term.param >= len(arguments) {
		return outsideValue()
	}
	if term.level == 0 {
		return arguments[term.param].copy()
	}
	var start value
	if term.field == anyField {
		start = analysis.everything(arguments[term.param])
	} else {
		start = analysis.load(arguments[term.param], term.field)
	}
	if term.level == 1 {
		return start
	}
	start = analysis.everything(start)
	var reached value
	queue := append(start.strong.sorted(), start.weak.sorted()...)
	visited := objects{}
	for len(queue) > 0 {
		object := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if !visited.add(object) {
			continue
		}
		reached.addStrong(object)
		reached.addWeak(object)
		switch {
		case object < 0:
			// One of this function's own placeholders: everything below it too.
			queue = append(queue, analysis.below(analysis.below(object, anyField), anyField))
		case analysis.state.exposed(object):
			queue = append(queue, outside)
		}
		for _, held := range analysis.state.heap[object] {
			queue = append(queue, held.strong.sorted()...)
			queue = append(queue, held.weak.sorted()...)
		}
	}
	return reached
}

func (proof *freshness) functionName(function int) string {
	if function < 0 {
		return "the top level"
	}
	return proof.program.Functions[function].Name
}

// summary is what a call does, from its caller's side: what it returns, and the fresh objects that
// reaches, by origin (-1 is outside); what it stored into what its parameters reach (edges); which of
// those it let escape (leaks); whether it ran anything it can't see into (clobbers); and the writes
// it couldn't prove, left to its callers.
type summary struct {
	result   summaryValue
	fields   map[int]map[string]summaryValue
	edges    map[term]map[string]summaryValue
	leaks    []term
	clobbers bool
	deferred map[writeKey]summaryDeferral
}

type summaryValue struct {
	strong, weak           []int
	strongTerms, weakTerms []term
}

type summaryDeferral struct {
	write          Write
	holder, stored summaryValue
}

func (summary *summary) origins() []int {
	seen := map[int]bool{}
	note := func(held summaryValue) {
		for _, origin := range append(append([]int{}, held.strong...), held.weak...) {
			if origin >= 0 {
				seen[origin] = true
			}
		}
	}
	note(summary.result)
	for origin, fields := range summary.fields {
		seen[origin] = true
		for _, held := range fields {
			note(held)
		}
	}
	for _, fields := range summary.edges {
		for _, held := range fields {
			note(held)
		}
	}
	for _, deferred := range summary.deferred {
		note(deferred.holder)
		note(deferred.stored)
	}
	list := []int{}
	for origin := range seen {
		list = append(list, origin)
	}
	sort.Ints(list)
	return list
}

func (summary *summary) edgeTerms() []term {
	list := []term{}
	for key := range summary.edges {
		list = append(list, key)
	}
	sortTerms(list)
	return list
}

func (summary *summary) deferredKeys() []writeKey {
	list := []writeKey{}
	for key := range summary.deferred {
		list = append(list, key)
	}
	sortWriteKeys(list)
	return list
}

func sortTerms(list []term) {
	slices.SortFunc(list, func(left, right term) int {
		if left.param != right.param {
			return left.param - right.param
		}
		if left.level != right.level {
			return left.level - right.level
		}
		return strings.Compare(left.field, right.field)
	})
}

func sortWriteKeys(keys []writeKey) {
	slices.SortFunc(keys, func(left, right writeKey) int {
		if left.function != right.function {
			return left.function - right.function
		}
		if left.instruction != right.instruction {
			return int(left.instruction) - int(right.instruction)
		}
		return left.ordinal - right.ordinal
	})
}

func (summary *summary) String() string {
	if summary == nil {
		return "nothing"
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%v clobbers=%v leaks=%v", summary.result, summary.clobbers, summary.leaks)
	fieldsOf := func(prefix string, fields map[string]summaryValue) {
		names := []string{}
		for name := range fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(&text, " %s.%q=%v", prefix, name, fields[name])
		}
	}
	for _, origin := range summary.origins() {
		fieldsOf(fmt.Sprint(origin), summary.fields[origin])
	}
	for _, term := range summary.edgeTerms() {
		fieldsOf(fmt.Sprintf("%v", term), summary.edges[term])
	}
	for _, key := range summary.deferredKeys() {
		fmt.Fprintf(&text, " deferred %v=%v,%v", key, summary.deferred[key].holder, summary.deferred[key].stored)
	}
	return text.String()
}

func (summary *summary) equal(other *summary) bool {
	if (summary == nil) != (other == nil) {
		return false
	}
	return summary.String() == other.String()
}

// summarize is the function's summary from the state it leaves in: each site's instances taken
// together by where they were first made, leaked if either has, and each placeholder as its term.
func (analysis *analysis) summarize(exit *state) *summary {
	leaked := func(object object) bool {
		if object == outside {
			return true
		}
		if object < 0 {
			return false
		}
		site := siteOf(object)
		return exit.leaked[newest(site)] || exit.leaked[older(site)]
	}
	collapse := func(held value) summaryValue {
		strong, weak := map[int]bool{}, map[int]bool{}
		made := summaryValue{}
		for key := range held.strong {
			switch {
			case key < 0:
				made.strongTerms = append(made.strongTerms, analysis.termOf(key))
			case leaked(key):
				strong[-1] = true
			default:
				strong[analysis.siteOrigin[siteOf(key)]] = true
			}
		}
		for key := range held.weak {
			switch {
			case key < 0:
				made.weakTerms = append(made.weakTerms, analysis.termOf(key))
			case leaked(key):
				weak[-1] = true
			default:
				weak[analysis.siteOrigin[siteOf(key)]] = true
			}
		}
		made.strong, made.weak = sortedKeys(strong), sortedKeys(weak)
		sortTerms(made.strongTerms)
		sortTerms(made.weakTerms)
		return made
	}
	made := &summary{
		fields:   map[int]map[string]summaryValue{},
		edges:    map[term]map[string]summaryValue{},
		deferred: map[writeKey]summaryDeferral{},
		clobbers: exit.clobbered,
	}
	returned := exit.locals[analysis.returnedLocal]
	made.result = collapse(returned)
	roots := []value{returned}
	for key, fields := range exit.heap {
		if key >= 0 {
			continue
		}
		term := analysis.termOf(key)
		made.edges[term] = map[string]summaryValue{}
		for field, held := range fields {
			made.edges[term][field] = collapse(held)
			roots = append(roots, held)
		}
	}
	for key := range exit.leaked {
		if key < 0 {
			made.leaks = append(made.leaks, analysis.termOf(key))
		}
	}
	sortTerms(made.leaks)
	for key, deferred := range analysis.deferred {
		made.deferred[key] = summaryDeferral{
			write:  deferred.write,
			holder: collapse(deferred.holder),
			stored: collapse(deferred.stored),
		}
		roots = append(roots, deferred.holder, deferred.stored)
	}
	// The fresh objects any of that reaches, and what each holds.
	gathered := map[int]map[string]value{}
	var queue []object
	for _, root := range roots {
		queue = append(queue, root.strong.sorted()...)
		queue = append(queue, root.weak.sorted()...)
	}
	visited := objects{}
	for len(queue) > 0 {
		object := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if object < 0 || leaked(object) || !visited.add(object) {
			continue
		}
		origin := analysis.siteOrigin[siteOf(object)]
		if gathered[origin] == nil {
			gathered[origin] = map[string]value{}
		}
		for field, held := range exit.heap[object] {
			mine := gathered[origin][field]
			mine.merge(held)
			gathered[origin][field] = mine
			queue = append(queue, held.strong.sorted()...)
			queue = append(queue, held.weak.sorted()...)
		}
	}
	for origin, fields := range gathered {
		made.fields[origin] = map[string]summaryValue{}
		for field, held := range fields {
			made.fields[origin][field] = collapse(held)
		}
	}
	return made
}

func sortedKeys(set map[int]bool) []int {
	list := []int{}
	for key := range set {
		list = append(list, key)
	}
	sort.Ints(list)
	return list
}

// mutable reports whether a value of a type can hold another: an object, an array, a map, a closure
// (whose cells can), a union that may be one, or a Weak. Numbers, booleans and strings can't.
func mutable(valueType ir.Type) bool {
	return valueType.IsReference() && valueType != ir.String
}

// expressionEffect interprets only effects that lowering admits inside expressions.
func (analysis *analysis) expressionEffect(statement ir.Statement) {
	switch statement := statement.(type) {
	case ir.Assign:
		analysis.define(statement.Local, analysis.value(statement.Value))
	case ir.Declare:
		analysis.define(statement.Local, analysis.value(statement.Value))
	case ir.Evaluate:
		analysis.value(statement.Value)
	case ir.WriteLine:
		analysis.value(statement.Value)
	case ir.Panic:
		analysis.value(statement.Message)
	case ir.SetProperty:
		holder := analysis.value(statement.Object)
		held := analysis.value(statement.Value)
		analysis.write(WriteField, statement.Site, statement.Name, holder, held, statement.Name)
	case ir.SetIndex:
		holder := analysis.value(statement.Array)
		analysis.value(statement.Index)
		held := analysis.value(statement.Value)
		analysis.write(WriteElement, statement.Site, "", holder, held, elementKey)
	case ir.Block:
		for _, inner := range statement.Body {
			analysis.expressionEffect(inner)
		}
	default:
		analysis.unknown(statement)
	}
}
