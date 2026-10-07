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
	proof := &freshness{program: program, summaries: map[int]*summary{}, origins: map[originKey]int{}, direct: directlyCalled(program)}
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
	keys := make([]writeKey, 0, len(proof.writes))
	for key := range proof.writes {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b writeKey) int {
		if a.function != b.function {
			return a.function - b.function
		}
		if a.instruction != b.instruction {
			return int(a.instruction) - int(b.instruction)
		}
		return a.ordinal - b.ordinal
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
		if function.Closure {
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
	// Unaliased proves the holder has never been stored, captured or passed.
	Unaliased bool
	Kind      WriteKind

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

func newest(site int) object { return object(2*site + 1) }
func older(site int) object  { return object(2*site + 2) }
func siteOf(o object) int    { return int(o-1) / 2 }

// objects is a set of abstract objects.
type objects map[object]bool

func (s objects) add(o object) bool {
	if s[o] {
		return false
	}
	s[o] = true
	return true
}

func (s objects) sorted() []object {
	list := make([]object, 0, len(s))
	for o := range s {
		list = append(list, o)
	}
	slices.Sort(list)
	return list
}

// value is what an expression or a slot may hold: objects held strongly, and objects held through a
// Weak, which keeps nothing.
type value struct {
	strong, weak objects
}

func (v *value) addStrong(o object) bool {
	if v.strong == nil {
		v.strong = objects{}
	}
	return v.strong.add(o)
}

func (v *value) addWeak(o object) bool {
	if v.weak == nil {
		v.weak = objects{}
	}
	return v.weak.add(o)
}

// merge adds other's objects, and says whether anything was new.
func (v *value) merge(other value) bool {
	changed := false
	for o := range other.strong {
		changed = v.addStrong(o) || changed
	}
	for o := range other.weak {
		changed = v.addWeak(o) || changed
	}
	return changed
}

func (v value) empty() bool { return len(v.strong) == 0 && len(v.weak) == 0 }

func (v value) copy() value {
	var made value
	made.merge(v)
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

func (s *state) copy() *state {
	made := newState()
	for local, held := range s.locals {
		made.locals[local] = held.copy()
	}
	for o, fields := range s.heap {
		copied := map[string]value{}
		for field, held := range fields {
			copied[field] = held.copy()
		}
		made.heap[o] = copied
	}
	for o := range s.escaped {
		made.escaped[o] = true
	}
	for o := range s.leaked {
		made.leaked[o] = true
	}
	made.clobbered = s.clobbered
	return made
}

// join adds other's facts to s, and says whether anything was new.
func (s *state) join(other *state) bool {
	changed := false
	for local, held := range other.locals {
		mine := s.locals[local]
		if mine.merge(held) {
			changed = true
		}
		s.locals[local] = mine
	}
	for o, fields := range other.heap {
		for field, held := range fields {
			if s.store(o, field, held) {
				changed = true
			}
		}
	}
	for o := range other.escaped {
		changed = s.escaped.add(o) || changed
	}
	for o := range other.leaked {
		changed = s.leaked.add(o) || changed
	}
	if other.clobbered && !s.clobbered {
		s.clobbered, changed = true, true
	}
	if changed {
		s.closeEscapes()
	}
	return changed
}

// store adds held to what o's field may hold, and says whether anything was new.
func (s *state) store(o object, field string, held value) bool {
	fields := s.heap[o]
	if fields == nil {
		fields = map[string]value{}
		s.heap[o] = fields
	}
	mine := fields[field]
	changed := mine.merge(held)
	fields[field] = mine
	return changed
}

// escape marks every object a value holds, strongly or weakly, as escaped and leaked, and
// everything they hold: this function let it go where it can't see.
func (s *state) escape(held value) {
	for o := range held.strong {
		s.escapeObject(o, true)
	}
	for o := range held.weak {
		s.escapeObject(o, true)
	}
}

// expose marks a value as escaped but not leaked: stored into what a parameter reaches, which the
// caller holds, and its summary says where.
func (s *state) expose(held value) {
	for o := range held.strong {
		s.escapeObject(o, false)
	}
	for o := range held.weak {
		s.escapeObject(o, false)
	}
}

func (s *state) escapeObject(o object, leak bool) {
	queue := []object{o}
	for len(queue) > 0 {
		next := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if next == outside {
			continue
		}
		added := s.escaped.add(next)
		if leak && s.leaked.add(next) {
			added = true
		}
		if !added {
			continue
		}
		for _, held := range s.heap[next] {
			queue = append(queue, held.strong.sorted()...)
			queue = append(queue, held.weak.sorted()...)
		}
	}
}

// closeEscapes keeps what escaped and leaked objects hold escaped and leaked: a join or a merge by
// recency can give one something that hadn't been yet.
func (s *state) closeEscapes() {
	exposed := s.escaped.sorted()
	for o := range s.heap {
		if o < 0 && !s.escaped[o] {
			exposed = append(exposed, o)
		}
	}
	for _, o := range exposed {
		for _, held := range s.heap[o] {
			if s.leaked[o] {
				s.escape(held)
			} else {
				s.expose(held)
			}
		}
	}
}

// exposed reports whether whoever is outside may hold o.
func (s *state) exposed(o object) bool {
	return o == outside || o < 0 || s.escaped[o]
}

// storeInto is a store into o that lets what's stored escape as o has: leaked into what's leaked,
// exposed into what a caller holds.
func (s *state) storeInto(o object, field string, held value) bool {
	if o == outside {
		s.escape(held)
		return false
	}
	changed := s.store(o, field, held)
	switch {
	case s.leaked[o]:
		s.escape(held)
	case s.exposed(o):
		s.expose(held)
	}
	return changed
}

// reach is every object a value reaches through strong references alone.
func (s *state) reach(from value) objects {
	reached := objects{}
	queue := from.strong.sorted()
	for len(queue) > 0 {
		o := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if !reached.add(o) {
			continue
		}
		for _, held := range s.heap[o] {
			queue = append(queue, held.strong.sorted()...)
		}
	}
	return reached
}

// closes reports whether writing held into holder may close a cycle: whether the value may reach,
// strongly, an object the holder may be. Why says what was found when it may.
func (s *state) closes(holder value, held value) (bool, string) {
	return s.closesWith(holder, held, s.exposed)
}

// closesWith is closes, with what the value may reach outside judged by exposed: a write a call
// replays is judged after it, and what the write itself let escape doesn't count against it.
func (s *state) closesWith(holder value, held value, exposed func(object) bool) (bool, string) {
	reached := s.reach(held)
	reachesOutside := false
	for o := range reached {
		if exposed(o) {
			reachesOutside = true
		}
	}
	for _, o := range holder.strong.sorted() {
		if reachesOutside && s.exposed(o) {
			if o == outside {
				return true, "the value written reaches something this function didn't make or let escape, and what it's written into wasn't made here"
			}
			return true, "the value written reaches something this function didn't make or let escape, and what it's written into has escaped by then"
		}
		if reached[o] {
			return true, "the value written may reach what it's written into"
		}
	}
	return false, ""
}

// recent makes a site's newest instance one of its older ones, before the site makes another.
func (s *state) recent(site int) {
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
	for local, held := range s.locals {
		rename(&held)
		s.locals[local] = held
	}
	for _, fields := range s.heap {
		for field, held := range fields {
			rename(&held)
			fields[field] = held
		}
	}
	if fields, ok := s.heap[from]; ok {
		delete(s.heap, from)
		for field, held := range fields {
			s.store(to, field, held)
		}
	}
	if s.escaped[from] {
		delete(s.escaped, from)
		s.escaped.add(to)
	}
	if s.leaked[from] {
		delete(s.leaked, from)
		s.leaked.add(to)
	}
	s.closeEscapes()
}

// freshness is the whole program's proof.
type freshness struct {
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

	sites      map[siteKey]int
	siteOrigin []int

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
func (a *analysis) placeholder(t term) object {
	if o, ok := a.termObjects[t]; ok {
		return o
	}
	a.terms = append(a.terms, t)
	o := object(-len(a.terms))
	a.termObjects[t] = o
	return o
}

// below is the placeholder for what a placeholder holds in a field: a parameter's field, then
// everything below that.
func (a *analysis) below(o object, field string) object {
	t := a.termOf(o)
	switch t.level {
	case 0:
		return a.placeholder(term{param: t.param, field: field, level: 1})
	case 1:
		return a.placeholder(term{param: t.param, field: t.field, level: 2})
	}
	return o
}

// termOf is a placeholder's term.
func (a *analysis) termOf(o object) term {
	return a.terms[-int(o)-1]
}

// load is what a field of the objects a value holds strongly may hold.
func (a *analysis) load(from value, field string) value {
	var loaded value
	for o := range from.strong {
		fields := a.state.heap[o]
		loaded.merge(fields[field])
		loaded.merge(fields[anyField])
		switch {
		case o < 0:
			// What the argument holds there, or anything below it.
			reached := a.below(o, field)
			loaded.addStrong(reached)
			loaded.addWeak(reached)
			if a.state.clobbered {
				loaded.merge(outsideValue())
			}
		case a.state.exposed(o):
			// Whoever is outside may have written anything outside there.
			loaded.merge(outsideValue())
		}
	}
	return loaded
}

// everything is what any field, element or entry of the objects a value holds strongly may hold.
func (a *analysis) everything(from value) value {
	var loaded value
	for o := range from.strong {
		for _, held := range a.state.heap[o] {
			loaded.merge(held)
		}
		switch {
		case o < 0:
			reached := a.below(o, anyField)
			loaded.addStrong(reached)
			loaded.addWeak(reached)
			if a.state.clobbered {
				loaded.merge(outsideValue())
			}
		case a.state.exposed(o):
			loaded.merge(outsideValue())
		}
	}
	return loaded
}

// analyze interprets a function to its fixed point, judges its writes, and returns its summary.
func (proof *freshness) analyze(function int, graph *flow.Function) *summary {
	a := &analysis{proof: proof, function: function, graph: graph, sites: map[siteKey]int{}, sitesOf: map[flow.InstructionId][]int{}, iterables: map[*ir.Statement]int{}, termObjects: map[term]object{}, deferred: map[writeKey]*deferral{}}
	a.returnedLocal = len(proof.program.Locals)
	a.placeholders = function >= 0 && proof.direct[function] && !proof.top
	entry := newState()
	if function >= 0 {
		for index, local := range proof.program.Functions[function].Parameters {
			declared := proof.program.Locals[local]
			if !mutable(declared.Type) || declared.Captured {
				continue
			}
			if a.placeholders {
				parameter := a.placeholder(term{param: index})
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
		before, after := a.block(block, in[block.Id])
		a.flowOut(block, before, after, func(successor flow.BlockId, out *state) {
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
	a.judging = true
	exit := newState()
	exited := false
	for _, block := range graph.Blocks {
		if in[block.Id] == nil {
			continue
		}
		_, after := a.block(block, in[block.Id])
		switch block.Terminal.(type) {
		case *flow.Return, *flow.Throw:
			// Leaving, either way: what it did on the way is its callers' to know.
			exit.join(after)
			exited = true
		}
	}
	if function < 0 || !exited {
		// Never leaving, its callers never judge what it deferred.
		for key, deferred := range a.deferred {
			a.proof.unproven(key, deferred.write, deferred.write.Why+", in a function that never returns or throws")
		}
		return nil
	}
	return a.summarize(exit)
}

// flowOut hands a block's state to each block after it: a handler takes either the state before
// the instruction that threw or after it, since a call that throws never gives its variable a value.
func (a *analysis) flowOut(block *flow.BasicBlock, before *state, after *state, visit func(flow.BlockId, *state)) {
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
func (a *analysis) block(block *flow.BasicBlock, entry *state) (*state, *state) {
	a.state = entry.copy()
	before := a.state
	for index, id := range block.Instructions {
		if index == len(block.Instructions)-1 {
			before = a.state.copy()
		}
		a.run(id)
	}
	return before, a.state
}

// run interprets one instruction.
func (a *analysis) run(id flow.InstructionId) {
	instruction := a.graph.Instructions[id]
	a.instruction, a.allocations, a.writeCount = id, 0, 0
	for _, site := range a.sitesOf[id] {
		a.state.recent(site)
	}
	switch statement := (*instruction.At).(type) {
	case ir.SetProperty:
		holder := a.value(statement.Object)
		held := a.value(statement.Value)
		a.write(WriteField, statement.Site, statement.Name, holder, held, statement.Name)
	case ir.SetIndex:
		holder := a.value(statement.Array)
		a.value(statement.Index)
		held := a.value(statement.Value)
		a.write(WriteElement, statement.Site, "", holder, held, elementKey)
	case ir.Declare:
		a.define(statement.Local, a.value(statement.Value))
	case ir.Assign:
		a.define(statement.Local, a.value(statement.Value))
	case ir.Return:
		returned := a.value(statement.Value)
		held := a.state.locals[a.returnedLocal]
		held.merge(returned)
		a.state.locals[a.returnedLocal] = held
	case ir.Throw:
		a.state.escape(a.value(statement.Value))
	case ir.ForOf:
		iterable := a.iterable(instruction.At)
		if instruction.Part == 0 {
			a.state.locals[iterable] = a.value(statement.Iterable)
			break
		}
		elements := a.load(a.state.locals[iterable], elementKey)
		if statement.Pattern == nil {
			a.define(statement.Local, elements)
			break
		}
		for _, binding := range statement.Pattern {
			// A pair a map gives, or a tuple of the array's: what the element holds there.
			bound := elements.copy()
			bound.merge(a.load(elements, binding.Field))
			a.define(binding.Local, bound)
		}
	case ir.Try:
		// The catch begins by taking the error, which came from wherever it was thrown.
		if statement.CatchLocal >= 0 {
			a.define(statement.CatchLocal, outsideValue())
		}
	case ir.WriteLine, ir.Evaluate, ir.Panic, ir.If, ir.Loop, ir.Switch:
		// What they evaluate: a value written or discarded, a panic's message, a condition, a switch's
		// value or a case's test.
		a.value(instruction.Expression)
	default:
		// A statement added to the IR after this was written: it may write anything anywhere.
		a.unknown(statement)
		a.value(instruction.Expression)
	}
}

// iterable is the pseudo-local holding what a for...of iterates, from its first part to its second.
func (a *analysis) iterable(at *ir.Statement) int {
	local, ok := a.iterables[at]
	if !ok {
		local = a.returnedLocal + 1 + len(a.iterables)
		a.iterables[at] = local
	}
	return local
}

// define gives a variable a value: a tracked one holds it from here on, and one a call can reach (a
// global, a captured variable) lets it escape.
func (a *analysis) define(local int, held value) {
	declared := a.proof.program.Locals[local]
	if !mutable(declared.Type) {
		return
	}
	if declared.Global || declared.Captured {
		a.state.escape(held)
		return
	}
	a.state.locals[local] = held
}

// write judges a write of held into the field (or elements) of holder, and makes it.
func (a *analysis) write(kind WriteKind, site int, name string, holder value, held value, field string) {
	key := writeKey{function: a.function, instruction: a.instruction, ordinal: a.writeCount}
	a.writeCount++
	if a.judging {
		write := Write{Kind: kind, Site: site, Name: name, Function: a.function, Proven: true}
		recorded := a.proof.record(key, write)
		recorded.Unaliased = recorded.Unaliased && a.unaliased(holder)
		a.judge(key, write, holder, held, "", a.state.exposed)
	}
	for o := range holder.strong {
		a.state.storeInto(o, field, held)
	}
}

// judge decides a write here, or defers it to this function's callers: one this function can't
// prove alone may be proven where it's called, with what the caller knows of its arguments.
func (a *analysis) judge(key writeKey, write Write, holder value, held value, where string, exposed func(object) bool) {
	closes, why := a.state.closesWith(holder, held, exposed)
	if !closes {
		return
	}
	if a.placeholders {
		deferred := a.deferred[key]
		if deferred == nil {
			write.Why = why + where
			deferred = &deferral{write: write}
			a.deferred[key] = deferred
		}
		deferred.holder.merge(holder)
		deferred.stored.merge(held)
		return
	}
	a.proof.unproven(key, write, why+where)
}

// record notes a write the first time it's met in a round, proven until something says otherwise.
func (proof *freshness) record(key writeKey, write Write) *Write {
	recorded := proof.writes[key]
	if recorded == nil {
		recorded = &write
		recorded.Proven, recorded.Why = true, ""
		recorded.Unaliased = true
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
func (a *analysis) unknown(node any) {
	key := writeKey{function: a.function, instruction: a.instruction, ordinal: a.writeCount}
	a.writeCount++
	if a.judging {
		a.proof.writes[key] = &Write{Kind: WriteUnknown, Function: a.function, Why: fmt.Sprintf("%T is a node the cycle finder doesn't know", node)}
	}
}

// allocate is a new object made by this instruction: its next allocation's site, newest instance.
func (a *analysis) allocate() object {
	origin := a.proof.origin(originKey{function: a.function, instruction: a.instruction, ordinal: a.allocations})
	a.allocations++
	return newest(a.site(origin))
}

// site is this instruction's site for an origin, made the first time it's met.
func (a *analysis) site(origin int) int {
	key := siteKey{instruction: a.instruction, origin: origin}
	site, ok := a.sites[key]
	if !ok {
		site = len(a.siteOrigin)
		a.sites[key] = site
		a.siteOrigin = append(a.siteOrigin, origin)
		a.sitesOf[a.instruction] = append(a.sitesOf[a.instruction], site)
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
func (a *analysis) fresh(field string, contents value) value {
	made := a.allocate()
	if !contents.empty() {
		a.state.store(made, field, contents)
	}
	return value{strong: objects{made: true}}
}

// call is what a call does to what it's handed: everything escapes, and its result is outside.
func (a *analysis) call(operands []value, returns ir.Type) value {
	for _, operand := range operands {
		a.state.escape(operand)
	}
	// It may write anything it can reach into the parameters.
	a.state.clobbered = true
	if !mutable(returns) {
		return value{}
	}
	return outsideValue()
}

// value interprets an expression: its effects, and what its value may hold.
func (a *analysis) value(expression ir.Expression) value {
	if expression == nil {
		return value{}
	}
	if held, known := a.libraryLanguage(expression); known {
		return held
	}
	switch expression := expression.(type) {
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.JSONNull, ir.Null:
		return value{}
	case ir.Read:
		declared := a.proof.program.Locals[expression.Local]
		if !mutable(declared.Type) {
			return value{}
		}
		if declared.Global || declared.Captured {
			return outsideValue()
		}
		held, ok := a.state.locals[expression.Local]
		if !ok {
			// Never given a value on any path here: a read the checker let through that this pass can't
			// place, so it's taken as the most it could be.
			return outsideValue()
		}
		return held.copy()
	case ir.Unary:
		a.value(expression.Operand)
		return value{}
	case ir.Binary:
		a.value(expression.Left)
		a.value(expression.Right)
		return value{}
	case ir.IsNull:
		a.value(expression.Value)
		return value{}
	case ir.RegExpNew:
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		// Patterns are compiled constants; the runtime object holds only immutable strings.
		return a.fresh(anyField, value{})
	case ir.NodeHostCall:
		return a.call(a.operands(expression), expression.Type())
	case ir.RegExpCall:
		return a.regexCall(expression)
	case ir.RegExpProperty:
		a.value(expression.Array)
		if mutable(expression.Type()) {
			return outsideValue()
		}
		return value{}
	case ir.RegExpGroup:
		a.value(expression.Object)
		if mutable(expression.Type()) {
			return outsideValue()
		}
		return value{}
	case ir.JSONStringify:
		// Lowering excludes toJSON and replacer callbacks; serialization only reads values.
		a.value(expression.Value)
		a.value(expression.Replacer)
		a.value(expression.Space)
		return value{}
	case ir.NumberToString:
		a.value(expression.Value)
		return value{}
	case ir.BooleanToString:
		a.value(expression.Value)
		return value{}
	case ir.Concat:
		for _, part := range expression.Parts {
			a.value(part)
		}
		return value{}
	case ir.Conditional:
		a.value(expression.Condition)
		result := a.value(expression.WhenTrue)
		result.merge(a.value(expression.WhenNot))
		return result
	case ir.Coalesce:
		result := a.value(expression.Value)
		result.merge(a.value(expression.Fallback))
		a.value(expression.Panic)
		return result
	case ir.Box:
		return a.value(expression.Value)
	case ir.Narrow:
		return a.value(expression.Value)
	case ir.Unwrap:
		return a.value(expression.Value)
	case ir.Defined:
		return a.value(expression.Value)
	case ir.MaybeOf:
		return a.value(expression.Value)
	case ir.CheckedCast:
		result := a.value(expression.Value)
		for _, allowed := range expression.Allowed {
			a.value(allowed)
		}
		return result
	case ir.WeakOf:
		// A Weak keeps nothing: what it points at is held weakly.
		held := a.value(expression.Value)
		var weak value
		for o := range held.strong {
			weak.addWeak(o)
		}
		for o := range held.weak {
			weak.addWeak(o)
		}
		return weak
	case ir.WeakTarget:
		held := a.value(expression.Value)
		var target value
		for o := range held.weak {
			target.addStrong(o)
		}
		for o := range held.strong {
			target.addStrong(o)
		}
		return target
	case ir.Length:
		a.value(expression.Array)
		return value{}
	case ir.MathCall:
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		return value{}
	case ir.ObjectCall:
		return a.objectCall(expression)
	case ir.NumberCall:
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		return value{}
	case ir.ToFixed:
		a.value(expression.Value)
		a.value(expression.Digits)
		return value{}
	case ir.NumberFormat:
		a.value(expression.Value)
		a.value(expression.Argument)
		return value{}
	case ir.IsUndefined:
		a.value(expression.Value)
		return value{}
	case ir.MaybeToString:
		a.value(expression.Value)
		return value{}
	case ir.TypeOf:
		a.value(expression.Value)
		return value{}
	case ir.UnionToString:
		a.value(expression.Value)
		return value{}
	case ir.StringLength:
		a.value(expression.Value)
		return value{}
	case ir.CharCodeAt:
		a.value(expression.Value)
		a.value(expression.Index)
		return value{}
	case ir.Utf8Length:
		a.value(expression.Text)
		return value{}
	case ir.Utf8At:
		a.value(expression.Text)
		a.value(expression.Index)
		return value{}
	case ir.StringFromCodes:
		// A new string, from numbers: nothing it reaches can hold a holder.
		for _, code := range expression.Codes {
			a.value(code)
		}
		if expression.Spread != nil {
			a.value(expression.Spread)
		}
		return value{}
	case ir.Trim:
		a.value(expression.Value)
		return value{}
	case ir.StringIndex:
		a.value(expression.Value)
		a.value(expression.Index)
		return value{}
	case ir.StringCall:
		a.value(expression.Value)
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		if expression.Type() == ir.Array {
			// split: a new array of strings.
			return a.fresh(elementKey, value{})
		}
		return value{}
	case ir.CodePoints:
		a.value(expression.Value)
		return a.fresh(elementKey, value{})
	case ir.ProgramArguments:
		return a.fresh(elementKey, value{})
	case ir.ReadTextFile:
		a.value(expression.Path)
		return a.fresh(anyField, value{})
	case ir.WriteTextFile:
		a.value(expression.Path)
		a.value(expression.Text)
		return a.fresh(anyField, value{})
	case ir.MakeError:
		a.value(expression.Message)
		if expression.Name != nil {
			a.value(expression.Name)
		}
		return a.fresh(anyField, value{})
	case ir.ObjectLiteral:
		var copied value
		if expression.Spread != nil {
			source := a.value(expression.Spread)
			if expression.NoReuse {
				a.state.escape(source)
			}
			copied = a.everything(source)
		}
		all := append(append([]ir.Field{}, expression.Fields...), expression.Empty...)
		fields := make([]value, len(all))
		for index, field := range all {
			fields[index] = a.value(field.Value)
		}
		made := a.fresh(anyField, copied)
		for index, field := range all {
			if !fields[index].empty() {
				a.state.store(newestOf(made), field.Name, fields[index])
			}
		}
		return made
	case ir.ArrayLiteral:
		var elements value
		for index, element := range expression.Elements {
			held := a.value(element)
			if index < len(expression.Spread) && expression.Spread[index] {
				held = a.load(held, elementKey)
			}
			elements.merge(held)
		}
		return a.fresh(elementKey, elements)
	case ir.CollectionIterator:
		return a.fresh(anyField, a.value(expression.Collection))
	case ir.MapNew:
		var entries value
		for _, entry := range expression.Entries {
			entries.merge(a.value(entry[0]))
			entries.merge(a.value(entry[1]))
		}
		if expression.Pairs != nil {
			// Each pair's key and value: a tuple's fields, or another map's entries.
			pairs := a.load(a.value(expression.Pairs), elementKey)
			entries.merge(pairs)
			entries.merge(a.everything(pairs))
		}
		return a.fresh(elementKey, entries)
	case ir.MapKeys:
		return a.fresh(elementKey, a.load(a.value(expression.Map), elementKey))
	case ir.MapValues:
		return a.fresh(elementKey, a.load(a.value(expression.Map), elementKey))
	case ir.MapClear:
		a.value(expression.Map)
		return value{}
	case ir.MapForEach:
		return a.call([]value{a.value(expression.Map), a.value(expression.Callback)}, 0)
	case ir.ReadDirectory:
		a.value(expression.Path)
		return a.fresh(anyField, a.fresh(elementKey, value{}))
	case ir.NodeFSFile:
		return a.nodeFSFile(expression)
	case ir.RealPath:
		a.value(expression.Path)
		return a.fresh(anyField, value{})
	case ir.FileStatus:
		a.value(expression.Path)
		return a.fresh(anyField, value{})
	case ir.SetNew:
		return a.fresh(elementKey, a.load(a.value(expression.Values), elementKey))
	case ir.SetValues:
		return a.fresh(elementKey, a.load(a.value(expression.Set), elementKey))
	case ir.MapEntries:
		contents := a.load(a.value(expression.Map), elementKey)
		pairs := a.fresh(anyField, contents)
		return a.fresh(elementKey, pairs)
	case ir.ArraySlice:
		elements := a.load(a.value(expression.Array), elementKey)
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		return a.fresh(elementKey, elements)
	case ir.ArrayConcat:
		elements := a.load(a.value(expression.Array), elementKey)
		for _, other := range expression.Others {
			elements.merge(a.load(a.value(other), elementKey))
		}
		return a.fresh(elementKey, elements)
	case ir.HasAccessor:
		a.value(expression.Object)
		return value{}
	case ir.Property:
		return a.load(a.value(expression.Object), expression.Name)
	case ir.ArrayIndex:
		array := a.value(expression.Array)
		a.value(expression.Index)
		return a.load(array, elementKey)
	case ir.ArrayPop:
		return a.load(a.value(expression.Array), elementKey)
	case ir.ArraySearch:
		a.value(expression.Array)
		a.value(expression.Value)
		if expression.From != nil {
			a.value(expression.From)
		}
		return value{}
	case ir.ArrayJoin:
		a.value(expression.Array)
		a.value(expression.Separator)
		return value{}
	case ir.ArrayReverse:
		return a.value(expression.Array)
	case ir.ArrayPush:
		holder := a.value(expression.Array)
		held := a.value(expression.Value)
		a.write(WriteElement, expression.Site, "", holder, held, elementKey)
		return value{}
	case ir.ArrayFill:
		if expression.Array == nil {
			a.value(expression.Length)
			held := a.value(expression.Value)
			a.value(expression.Start)
			a.value(expression.End)
			return a.fresh(elementKey, held)
		}
		holder := a.value(expression.Array)
		held := a.value(expression.Value)
		a.value(expression.Start)
		a.value(expression.End)
		a.write(WriteElement, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.ArraySplice:
		holder := a.value(expression.Array)
		a.value(expression.Start)
		a.value(expression.Count)
		var items value
		for _, item := range expression.Items {
			items.merge(a.value(item))
		}
		removed := a.load(holder, elementKey)
		if len(expression.Items) > 0 {
			a.write(WriteElement, expression.Site, "", holder, items, elementKey)
		}
		return a.fresh(elementKey, removed)
	case ir.MapGet:
		object := a.value(expression.Map)
		a.value(expression.Key)
		return a.load(object, elementKey)
	case ir.MapSet:
		holder := a.value(expression.Map)
		held := a.value(expression.Key)
		held.merge(a.value(expression.Value))
		a.write(WriteMapEntry, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.MapHas:
		a.value(expression.Map)
		a.value(expression.Key)
		return value{}
	case ir.MapDelete:
		a.value(expression.Map)
		a.value(expression.Key)
		return value{}
	case ir.MapSize:
		a.value(expression.Map)
		return value{}
	case ir.HasOwn:
		a.value(expression.Object)
		a.value(expression.Key)
		return value{}
	case ir.SetAdd:
		holder := a.value(expression.Set)
		held := a.value(expression.Value)
		a.write(WriteSetElement, expression.Site, "", holder, held, elementKey)
		return holder
	case ir.MakeClosure:
		// What a function value holds is its cells, and what's in a cell has escaped.
		return outsideValue()
	case ir.InstanceOf:
		a.value(expression.Value)
		return value{}
	case ir.Call:
		operands := []value{}
		for _, argument := range expression.Arguments {
			operands = append(operands, a.value(argument))
		}
		if a.proof.top {
			return a.call(operands, expression.Returns)
		}
		before := a.state.copy()
		joined := before.copy()
		var result value
		for _, target := range a.proof.program.CallTargets(expression) {
			a.state = before.copy()
			if !a.proof.direct[target] {
				// Its parameters are outside to it, so it may keep them anywhere.
				a.call(operands, 0)
			}
			result.merge(a.made(target, a.proof.summaries[target], operands))
			joined.join(a.state)
		}
		a.state = joined
		if !mutable(expression.Returns) {
			return value{}
		}
		return result
	case ir.CallClosure:
		// Cells and structural receivers are outside this summary's model. The
		// union for any bounded target set, and for Unknown, is an arbitrary call.
		return a.call(a.operands(expression), expression.Returns)
	case ir.ArrayMap:
		return a.call([]value{a.value(expression.Array), a.value(expression.Callback)}, ir.Array)
	case ir.ArrayVisit:
		return a.call([]value{a.value(expression.Array), a.value(expression.Callback)}, expression.Type())
	case ir.ArrayReduce:
		return a.call([]value{a.value(expression.Array), a.value(expression.Callback), a.value(expression.Initial)}, expression.Result)
	case ir.ArrayFrom:
		return a.call([]value{a.value(expression.Length), a.value(expression.Callback)}, ir.Array)
	case ir.ArraySort:
		// The comparator, a function or a function value, is handed the elements.
		array := a.value(expression.Array)
		a.call([]value{array, a.value(expression.Callback)}, 0)
		return array
	}
	// A node added to the IR after this was written: it may write anything anywhere, so it keeps every
	// slot refused, and it's taken as a call.
	a.unknown(expression)
	return a.call(a.operands(expression), expression.Type())
}

// operands interprets every expression an IR node holds.
func (a *analysis) operands(expression ir.Expression) []value {
	var operands []value
	for _, operand := range expressionsIn(expression) {
		operands = append(operands, a.value(operand))
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
	for o := range made.strong {
		return o
	}
	return outside
}

// made is a summarized call, replayed here: its fresh objects made anew, each the newest instance of
// this instruction's site for where it was first made; what it stored into its parameters stored
// into the arguments; what it let escape of them let escape; and the writes it left to its callers
// judged with what's known here. Its result is what it returns.
func (a *analysis) made(function int, callee *summary, arguments []value) value {
	if callee == nil {
		// Returns nothing yet, on the way to the fixed point.
		return value{}
	}
	instances := map[int]object{}
	for _, origin := range callee.origins() {
		instances[origin] = newest(a.site(origin))
	}
	// What had escaped before the call: while the callee runs, only code it can't see into could
	// tie what it was given to what had, and then anything escaped counts.
	before := objects{}
	for o := range a.state.escaped {
		before.add(o)
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
		for _, t := range held.strongTerms {
			argument := a.argument(t, arguments)
			made.merge(argument)
		}
		for _, t := range held.weakTerms {
			argument := a.argument(t, arguments)
			for o := range argument.strong {
				made.addWeak(o)
			}
			for o := range argument.weak {
				made.addWeak(o)
			}
		}
		return made
	}
	// Until nothing more is added: a store through one argument can change what another reaches,
	// when the caller passed the same object for both.
	for {
		escaped, leaked := len(a.state.escaped), len(a.state.leaked)
		changed := false
		for _, origin := range callee.origins() {
			for field, held := range callee.fields[origin] {
				changed = a.state.storeInto(instances[origin], field, translate(held)) || changed
			}
		}
		for _, t := range callee.edgeTerms() {
			holders := translate(summaryValue{strongTerms: []term{t}})
			for field, held := range callee.edges[t] {
				stored := translate(held)
				for _, o := range holders.strong.sorted() {
					changed = a.state.storeInto(o, field, stored) || changed
				}
			}
		}
		for _, t := range callee.leaks {
			a.state.escape(translate(summaryValue{strongTerms: []term{t}}))
		}
		if !changed && escaped == len(a.state.escaped) && leaked == len(a.state.leaked) {
			break
		}
	}
	if callee.clobbers {
		a.state.clobbered = true
	}
	if a.judging {
		where := ", where " + a.proof.functionName(function) + " is called from " + a.proof.functionName(a.function)
		for _, key := range callee.deferredKeys() {
			deferred := callee.deferred[key]
			exposed := func(o object) bool {
				return o == outside || o < 0 || before[o] || (callee.clobbers && a.state.exposed(o))
			}
			a.judge(key, deferred.write, translate(deferred.holder), translate(deferred.stored), where, exposed)
		}
	}
	return translate(callee.result)
}

// argument is what a placeholder of a callee's stands for here: the argument, or everything the
// argument reaches through the field, strongly or weakly.
func (a *analysis) argument(t term, arguments []value) value {
	if t.param >= len(arguments) {
		return outsideValue()
	}
	if t.level == 0 {
		return arguments[t.param].copy()
	}
	var start value
	if t.field == anyField {
		start = a.everything(arguments[t.param])
	} else {
		start = a.load(arguments[t.param], t.field)
	}
	if t.level == 1 {
		return start
	}
	start = a.everything(start)
	var reached value
	queue := append(start.strong.sorted(), start.weak.sorted()...)
	visited := objects{}
	for len(queue) > 0 {
		o := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if !visited.add(o) {
			continue
		}
		reached.addStrong(o)
		reached.addWeak(o)
		switch {
		case o < 0:
			// One of this function's own placeholders: everything below it too.
			queue = append(queue, a.below(a.below(o, anyField), anyField))
		case a.state.exposed(o):
			queue = append(queue, outside)
		}
		for _, held := range a.state.heap[o] {
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

func (s *summary) origins() []int {
	seen := map[int]bool{}
	note := func(held summaryValue) {
		for _, origin := range append(append([]int{}, held.strong...), held.weak...) {
			if origin >= 0 {
				seen[origin] = true
			}
		}
	}
	note(s.result)
	for origin, fields := range s.fields {
		seen[origin] = true
		for _, held := range fields {
			note(held)
		}
	}
	for _, fields := range s.edges {
		for _, held := range fields {
			note(held)
		}
	}
	for _, deferred := range s.deferred {
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

func (s *summary) edgeTerms() []term {
	list := []term{}
	for t := range s.edges {
		list = append(list, t)
	}
	sortTerms(list)
	return list
}

func (s *summary) deferredKeys() []writeKey {
	list := []writeKey{}
	for key := range s.deferred {
		list = append(list, key)
	}
	sortWriteKeys(list)
	return list
}

func sortTerms(list []term) {
	slices.SortFunc(list, func(a, b term) int {
		if a.param != b.param {
			return a.param - b.param
		}
		if a.level != b.level {
			return a.level - b.level
		}
		return strings.Compare(a.field, b.field)
	})
}

func sortWriteKeys(keys []writeKey) {
	slices.SortFunc(keys, func(a, b writeKey) int {
		if a.function != b.function {
			return a.function - b.function
		}
		if a.instruction != b.instruction {
			return int(a.instruction) - int(b.instruction)
		}
		return a.ordinal - b.ordinal
	})
}

func (s *summary) String() string {
	if s == nil {
		return "nothing"
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%v clobbers=%v leaks=%v", s.result, s.clobbers, s.leaks)
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
	for _, origin := range s.origins() {
		fieldsOf(fmt.Sprint(origin), s.fields[origin])
	}
	for _, t := range s.edgeTerms() {
		fieldsOf(fmt.Sprintf("%v", t), s.edges[t])
	}
	for _, key := range s.deferredKeys() {
		fmt.Fprintf(&text, " deferred %v=%v,%v", key, s.deferred[key].holder, s.deferred[key].stored)
	}
	return text.String()
}

func (s *summary) equal(other *summary) bool {
	if (s == nil) != (other == nil) {
		return false
	}
	return s.String() == other.String()
}

// summarize is the function's summary from the state it leaves in: each site's instances taken
// together by where they were first made, leaked if either has, and each placeholder as its term.
func (a *analysis) summarize(exit *state) *summary {
	leaked := func(o object) bool {
		if o == outside {
			return true
		}
		if o < 0 {
			return false
		}
		site := siteOf(o)
		return exit.leaked[newest(site)] || exit.leaked[older(site)]
	}
	collapse := func(held value) summaryValue {
		strong, weak := map[int]bool{}, map[int]bool{}
		made := summaryValue{}
		for o := range held.strong {
			switch {
			case o < 0:
				made.strongTerms = append(made.strongTerms, a.termOf(o))
			case leaked(o):
				strong[-1] = true
			default:
				strong[a.siteOrigin[siteOf(o)]] = true
			}
		}
		for o := range held.weak {
			switch {
			case o < 0:
				made.weakTerms = append(made.weakTerms, a.termOf(o))
			case leaked(o):
				weak[-1] = true
			default:
				weak[a.siteOrigin[siteOf(o)]] = true
			}
		}
		made.strong, made.weak = sortedKeys(strong), sortedKeys(weak)
		sortTerms(made.strongTerms)
		sortTerms(made.weakTerms)
		return made
	}
	made := &summary{fields: map[int]map[string]summaryValue{}, edges: map[term]map[string]summaryValue{}, deferred: map[writeKey]summaryDeferral{}, clobbers: exit.clobbered}
	returned := exit.locals[a.returnedLocal]
	made.result = collapse(returned)
	roots := []value{returned}
	for o, fields := range exit.heap {
		if o >= 0 {
			continue
		}
		t := a.termOf(o)
		made.edges[t] = map[string]summaryValue{}
		for field, held := range fields {
			made.edges[t][field] = collapse(held)
			roots = append(roots, held)
		}
	}
	for o := range exit.leaked {
		if o < 0 {
			made.leaks = append(made.leaks, a.termOf(o))
		}
	}
	sortTerms(made.leaks)
	for key, deferred := range a.deferred {
		made.deferred[key] = summaryDeferral{write: deferred.write, holder: collapse(deferred.holder), stored: collapse(deferred.stored)}
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
		o := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if o < 0 || leaked(o) || !visited.add(o) {
			continue
		}
		origin := a.siteOrigin[siteOf(o)]
		if gathered[origin] == nil {
			gathered[origin] = map[string]value{}
		}
		for field, held := range exit.heap[o] {
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

// This is the existing escape/identity analysis, with a stricter freshness fact:
// even a confined local is an alias, so storing a literal ends this proof.
func (a *analysis) unaliased(holder value) bool {
	if len(holder.strong) != 1 || len(holder.weak) != 0 {
		return false
	}
	for object := range holder.strong {
		if object <= 0 || object%2 != 1 || a.state.exposed(object) {
			return false
		}
		for _, held := range a.state.locals {
			if a.state.reach(held)[object] {
				return false
			}
		}
		for _, fields := range a.state.heap {
			for _, held := range fields {
				if a.state.reach(held)[object] {
					return false
				}
			}
		}
	}
	return true
}
