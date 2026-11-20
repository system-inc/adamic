package fresh

import (
	"slices"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// An object's layout holds a slot for an optional field its factory writes, on any path, inside the
// construction window: from the allocation until the object first escapes (Adamic's ruling on #w8nanzw,
// rulings.md:78 applied). The window is this package's escape walk, never a second one: the object
// escapes where the walk says whoever is outside may hold it, stored in a global or a captured
// variable, into anything escaped, thrown, or handed to a call. A function leaving by a return closes
// the window too, since its caller holds what it made from then on. A write after the window never adds
// a slot.
//
// ConstructionWindow interprets one function with every call taken as one the walk can't see into (the
// walk's top, as prepare falls back to): a call the object is handed closes the window at that call,
// and so does any call reaching it some other way, and only the function's own statements are read. So
// lowering can ask it of a function before the program around it is whole, and a callee's summary never
// widens a window.

// WindowWrite is how a construction window writes a field.
type WindowWrite uint8

const (
	// WindowUnwritten is a field no write inside the window stores, on any path.
	WindowUnwritten WindowWrite = iota
	// WindowSomePaths is a field a write inside the window may store, though not on every path that
	// lets the object be seen.
	WindowSomePaths
	// WindowEveryPath is a field every path that lets the object be seen has written before it does.
	WindowEveryPath
)

// Window is what a construction window writes.
type Window struct {
	// Written is each field a write inside the window may store, on some path.
	Written map[string]bool

	// Settled is set when every path that lets the object be seen has written the same fields, in
	// Order, and nothing it only may have: then each of them is an own key before anyone can look, in
	// Order, the first-write order each path took.
	Settled bool
	Order   []string

	// every is each field every path that lets the object be seen has written.
	every map[string]bool
}

// Writes is the one predicate over (allocation, field): whether the construction window writes the
// field, on every path that lets the object be seen, on some path, or not at all. A write after the
// window is no write here.
func (w Window) Writes(field string) WindowWrite {
	switch {
	case w.every[field]:
		return WindowEveryPath
	case w.Written[field]:
		return WindowSomePaths
	}
	return WindowUnwritten
}

// ConstructionWindow is the construction window of the object function's allocation statement makes:
// a Declare or an Assign of a local, whose value is the allocation, a statement of the function's
// own body (or a block's in it) by its address there. An allocation the walk finds escaping as it's
// made, or never reaching a local it tracks, has a window that writes nothing.
func ConstructionWindow(program *ir.Program, function int, allocation *ir.Statement) Window {
	watch := &windowWatch{allocation: allocation, site: -1, facts: map[flow.InstructionId]*windowFact{}, entryExposed: map[flow.BlockId]bool{}, exitExposed: map[flow.BlockId]bool{}, judged: map[flow.BlockId]bool{}}
	proof := &freshness{program: program, summaries: map[int]*summary{}, origins: map[originKey]int{}, writes: map[writeKey]*Write{}, top: true, window: watch}
	graph := flow.Build(program, function)
	proof.analyze(function, graph)
	return watch.settle(graph)
}

// windowWatch is what the judging pass keeps of one allocation's window: its site once the
// allocation has run, and for each instruction and block it judged, the window's own facts.
type windowWatch struct {
	allocation *ir.Statement
	site       int

	facts        map[flow.InstructionId]*windowFact
	entryExposed map[flow.BlockId]bool
	exitExposed  map[flow.BlockId]bool
	judged       map[flow.BlockId]bool
}

// windowFact is one instruction's part in the window: whether it makes the object, the stores into
// it while it's confined, in the order they run, and whether whoever is outside may hold it after.
type windowFact struct {
	opens   bool
	stores  []windowStore
	exposed bool
}

// windowStore is a store into the confined object: the field, and whether it surely runs and surely
// stores into this object, a statement whose receiver can be nothing else.
type windowStore struct {
	name  string
	exact bool
}

func (w *windowWatch) fact(id flow.InstructionId) *windowFact {
	fact := w.facts[id]
	if fact == nil {
		fact = &windowFact{}
		w.facts[id] = fact
	}
	return fact
}

// confined is whether the allocation's newest instance is made and nobody outside may hold it.
func (w *windowWatch) confined(a *analysis) bool {
	return w.site >= 0 && !a.state.exposed(newest(w.site))
}

// entered and left keep, for a block the judging pass runs, whether the object is exposed as it
// enters and as it leaves (a suspension lets every local go after the last instruction).
func (w *windowWatch) entered(a *analysis, block flow.BlockId) {
	if a.judging {
		w.judged[block] = true
		w.entryExposed[block] = !w.confined(a)
	}
}

func (w *windowWatch) left(a *analysis, block flow.BlockId) {
	if a.judging {
		w.exitExposed[block] = !w.confined(a)
	}
}

// ran notes an instruction that has run: the allocation names the object's site, the newest instance
// the local it defines holds and nothing else, and the judging pass keeps whether it's exposed after.
func (w *windowWatch) ran(a *analysis, id flow.InstructionId) {
	instruction := a.graph.Instructions[id]
	if instruction.At == w.allocation && instruction.Part == 0 {
		local := -1
		switch statement := (*instruction.At).(type) {
		case ir.Declare:
			local = statement.Local
		case ir.Assign:
			local = statement.Local
		}
		if held, tracked := a.state.locals[local]; tracked && len(held.strong) == 1 {
			for made := range held.strong {
				if made > 0 && made == newest(siteOf(made)) {
					w.site = siteOf(made)
				}
			}
		}
		if a.judging {
			w.fact(id).opens = true
		}
	}
	if a.judging {
		w.fact(id).exposed = !w.confined(a)
	}
}

// stored notes a field write the judging pass meets whose holder may be the confined object.
func (w *windowWatch) stored(a *analysis, name string, holder value) {
	if !w.confined(a) || !holder.strong[newest(w.site)] {
		return
	}
	fact := w.fact(a.instruction)
	fact.stores = append(fact.stores, windowStore{name: name, exact: a.statementWrite && len(holder.strong) == 1})
}

// windowState is the window at one point of the function: not reached (no path here has made the
// object), closed (the object may be seen), or open, with the fields every path here has surely written
// inside it, and whether every path wrote them in that one order and nothing it only may have.
type windowState struct {
	reached, open bool
	written       []string
	ordered       bool
}

func (s windowState) equal(other windowState) bool {
	return s.reached == other.reached && s.open == other.open && s.ordered == other.ordered && slices.Equal(s.written, other.written)
}

// stored is the state after a store into the confined object. A store that may not run, or may store
// elsewhere, leaves no field surely written and the order unknown.
func (s windowState) stored(store windowStore) windowState {
	if !s.open {
		return s
	}
	if !store.exact {
		s.ordered = false
		return s
	}
	if !slices.Contains(s.written, store.name) {
		s.written = append(slices.Clone(s.written), store.name)
	}
	return s
}

// joinWindows is the window where two paths meet: the fields both surely wrote, in one order only when
// both wrote the same ones in it. An open path meeting a closed one keeps nothing surely written.
func joinWindows(x, y windowState) windowState {
	switch {
	case !x.reached:
		return y
	case !y.reached:
		return x
	case !x.open && !y.open:
		return x
	case x.open != y.open:
		return windowState{reached: true, open: true}
	}
	joined := windowState{reached: true, open: true, ordered: x.ordered && y.ordered && slices.Equal(x.written, y.written)}
	for _, name := range x.written {
		if slices.Contains(y.written, name) {
			joined.written = append(joined.written, name)
		}
	}
	return joined
}

// block runs one block's window from its state on entry, handing shown each state the object may be
// seen in: where an instruction or the block's end lets it be held outside, and at a return. A second
// allocation while the last instance is still open closes that one too, since an older instance can be
// seen later. It gives the state before the block's last instruction, for the handler a throw there
// reaches, and after the block.
func (w *windowWatch) block(block *flow.BasicBlock, entry windowState, shown func(windowState)) (windowState, windowState) {
	current, before := entry, entry
	for index, id := range block.Instructions {
		if index == len(block.Instructions)-1 {
			before = current
		}
		fact := w.facts[id]
		if fact == nil {
			continue
		}
		if fact.opens {
			if current.open {
				shown(current)
			}
			current = windowState{reached: true, open: true, ordered: true}
		}
		for _, store := range fact.stores {
			current = current.stored(store)
		}
		if fact.exposed && current.open {
			shown(current)
			current = windowState{reached: true}
		}
	}
	if w.exitExposed[block.Id] && current.open {
		shown(current)
		current = windowState{reached: true}
	}
	if _, returns := block.Terminal.(*flow.Return); returns && current.open {
		shown(current)
	}
	return before, current
}

// edges hands a block's window to each block after it, as the walk's flowOut hands its state: a
// handler takes either the state before the last instruction or after it. A successor the walk enters
// with the object exposed closes the window handed to it.
func (w *windowWatch) edges(block *flow.BasicBlock, before, after windowState, shown func(windowState), visit func(flow.BlockId, windowState)) {
	hand := func(successor flow.BlockId, handed windowState) {
		if w.entryExposed[successor] && handed.open {
			shown(handed)
			handed = windowState{reached: true}
		}
		visit(successor, handed)
	}
	if throws, ok := block.Terminal.(*flow.MayThrow); ok {
		hand(throws.Next, after)
		hand(throws.Handler, joinWindows(before, after))
		return
	}
	flow.EachSuccessor(block.Terminal, func(successor flow.BlockId) { hand(successor, after) })
}

// settle walks the window over the blocks the judging pass ran, to its fixed point, then once more
// keeping every state the object may be seen in, and reads the window from those.
func (w *windowWatch) settle(graph *flow.Function) Window {
	window := Window{Written: map[string]bool{}, every: map[string]bool{}}
	for _, fact := range w.facts {
		for _, store := range fact.stores {
			window.Written[store.name] = true
		}
	}
	in := map[flow.BlockId]windowState{}
	entered := map[flow.BlockId]bool{graph.Entry: true}
	ignore := func(windowState) {}
	changed := true
	for changed {
		changed = false
		for _, block := range graph.Blocks {
			if !entered[block.Id] || !w.judged[block.Id] {
				continue
			}
			before, after := w.block(block, in[block.Id], ignore)
			w.edges(block, before, after, ignore, func(successor flow.BlockId, handed windowState) {
				joined := handed
				if entered[successor] {
					joined = joinWindows(in[successor], handed)
				}
				if !entered[successor] || !joined.equal(in[successor]) {
					in[successor], entered[successor] = joined, true
					changed = true
				}
			})
		}
	}
	var seen []windowState
	keep := func(state windowState) { seen = append(seen, state) }
	for _, block := range graph.Blocks {
		if !entered[block.Id] || !w.judged[block.Id] {
			continue
		}
		before, after := w.block(block, in[block.Id], keep)
		w.edges(block, before, after, keep, func(flow.BlockId, windowState) {})
	}
	if len(seen) == 0 {
		// Nothing lets the object be seen: no field is one every path shows.
		return window
	}
	for _, name := range seen[0].written {
		window.every[name] = true
	}
	window.Settled = seen[0].ordered
	for _, state := range seen[1:] {
		for name := range window.every {
			if !slices.Contains(state.written, name) {
				delete(window.every, name)
			}
		}
		window.Settled = window.Settled && state.ordered && slices.Equal(state.written, seen[0].written)
	}
	if window.Settled {
		window.Order = slices.Clone(seen[0].written)
	}
	return window
}
