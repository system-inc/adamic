package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// occurrenceKind separates the three things an identifier occurrence can be to the liveness pass.
//
// The distinction is forced by the graph rather than chosen. `controlflow`'s Read hook fires for
// every identifier assignment target, plain ones included: `patternReads` calls `b.read(node)` on a
// `KindIdentifier` target before the assigned expression, because ESLint counts naming the target as
// a point where an enclosing `try` can throw. Measured directly — `let v = 'a'; f(v); v = 'dead';`
// emits `R(v@3) W(v@3) R(f@12) R(v@15) R(v@18) W(v@18)`, and the `R(v@18)` is the dead store's own
// target, not a read of it. Trusting the hook would make every write trivially live and the rule
// would report nothing at all.
//
// So the hook says where in the flow an occurrence sits, and the AST says what it is. Both are
// needed: the AST alone cannot order two occurrences across a branch, and the hook alone cannot tell
// a store from a load.
type occurrenceKind uint8

const (
	// occurrenceRead is a genuine load of the binding's current value.
	occurrenceRead occurrenceKind = iota
	// occurrenceWrite is a store that does not first load: `v = 1`, or a declarator's initializer.
	occurrenceWrite
	// occurrenceUpdate is a store that loads first: `v += 1`, `v++`, a destructuring default. The
	// load half keeps an earlier write live and the store half kills it, in that order.
	occurrenceUpdate
)

// deadStoreEvent is one identifier occurrence, recorded where the control-flow walk reached it.
//
// The symbol is resolved once at collection rather than per query, because resolution goes through
// the checker and the earlier inverted implementation measured that cost at 54% of all rule time in
// the tree before it was grouped.
type deadStoreEvent struct {
	node   *ast.Node
	symbol *ast.Symbol
	kind   occurrenceKind
	// reportable records whether this write is one this rule may name. A write inside a `try`
	// block still kills an earlier write on the normal path, but is never itself reported, so the
	// two properties are carried separately rather than by dropping the event.
	reportable bool
	// number is the symbol's bit in this root's live sets, given once the root's graph is built.
	number int
}

// analyzeDeadStoresByLiveness reports every write in one file whose value no later read observes.
//
// This is upstream's shape: a backward liveness dataflow over real successor edges, run to a fixed
// point per code path root. A write is dead when its binding's bit is not live at the point the
// backward walk reaches it.
//
// # Why this replaced an inverted walk over TypeScript's flow graph
//
// The question is forward — which reads follow this write — and `ast.FlowNode` carries only
// `Antecedent`/`Antecedents`, pointing backward, because it exists for type narrowing. The previous
// implementation inverted the question rather than the graph, asking each read whether its own
// antecedent chain passed back through the candidate write. That is the same predicate and it works,
// but it pays for the inversion in three places the graph now answers directly: an update expression
// and a destructuring target have no flow node lining up with a single-identifier reachability
// question, and a function body starts a fresh flow graph with no antecedent path to the enclosing
// writes at all.
//
// `internal/lint/ecmascript/control_flow_graph` supplies forward `Successors`, so the question is asked in the
// direction it is posed. Measured against the same imported corpus, that recovers the update
// expressions, the destructuring targets, and the try-block kills the inversion had to decline.
func analyzeDeadStoresByLiveness(ctx rule.Context, sourceFile *ast.Node) {
	// Export state is a property of the file rather than of any root, so it is collected once. A
	// binding another module can read has no provably dead write.
	exported := exportedNames(sourceFile)

	// Symbol-level judgments are file-wide even though the liveness pass is per-root: a closure in
	// one root reading a binding declared in another is exactly the shape the capture guard exists
	// for, and a per-root scan cannot see it.
	symbols, occupied := collectSymbolFacts(ctx, sourceFile)

	// Only a root holding an occurrence of a tracked binding emits an event, and a root with no event
	// solves to nothing, so only those are built. On a cold ahra run that is 8.1% of roots (#tmn9n27).
	for _, root := range codePathRoots(ctx, sourceFile) {
		if !occupied[root.Node] {
			if CheckCodePathGates {
				deadStoreCheckSkipped(ctx, root.Node, symbols)
			}
			continue
		}
		analyzeRootLiveness(ctx, root.Node, exported, symbols)
	}
}

// deadStoreCheckSkipped is the gate's check under CheckCodePathGates: a root skipped for holding no
// tracked occurrence must emit no event.
func deadStoreCheckSkipped(ctx rule.Context, root *ast.Node, symbols map[*ast.Symbol]*symbolFacts) {
	for _, block := range deadStoreGraph(ctx, root, symbols).Blocks {
		if len(block.Events) > 0 {
			panic("no-useless-assignment skipped a root whose graph holds an event")
		}
	}
}

// symbolFacts are the three questions that belong to a binding rather than to any write to it.
type symbolFacts struct {
	// hasRead records whether the binding is loaded anywhere in the file. A binding nothing reads
	// has every write trivially dead, which is `no-unused-vars`' judgment restated once per
	// assignment. Upstream carries the same guard as `is_used`.
	hasRead bool
	// capturedRead records whether the binding is loaded from a code path root other than the one
	// declaring it. See narrowCapturedRead for why this stays as coarse as upstream's.
	capturedRead bool
	// eligible records whether the binding is a mutable local this rule may judge at all.
	eligible bool
	// declaration is the binding's first declaration in this file. A mutable local is declared here,
	// so a binding with no declaration in this file is not one this rule judges.
	declaration *ast.Node
}

// collectSymbolFacts resolves every identifier occurrence in the file once and files the per-binding
// judgments.
//
// One pass over the file rather than one per candidate. The name filter is what keeps this cheap:
// only identifiers spelled like some assignment target can matter, and a name comparison is far
// cheaper than a symbol resolution while being unable to err in the unsafe direction, since two
// bindings sharing a name are still separated by symbol identity and a read of a differently-named
// binding can never keep a write live.
//
// The second result is the roots an eligible binding's occurrence runs in, by RootOf. They are the
// only roots whose graph can hold an event, since an event is an occurrence of an eligible binding.
func collectSymbolFacts(ctx rule.Context, sourceFile *ast.Node) (map[*ast.Symbol]*symbolFacts, map[*ast.Node]bool) {
	var identifiers []*ast.Node
	interesting := map[string]bool{}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			if !isNonReferenceIdentifier(node) {
				identifiers = append(identifiers, node)
				if _, isWrite := writeTargetOfIncludingUpdates(node); isWrite {
					interesting[node.Text()] = true
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	if len(interesting) == 0 {
		return nil, nil
	}

	facts := map[*ast.Symbol]*symbolFacts{}
	occupied := map[*ast.Node]bool{}
	for _, identifier := range identifiers {
		if !interesting[identifier.Text()] {
			continue
		}
		symbol := resolveOccurrence(ctx, identifier)
		entry := facts[symbol]
		if entry == nil {
			declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
			if len(declarations) == 0 {
				continue
			}
			entry = &symbolFacts{eligible: isLocalVariableDeclaration(declarations[0]), declaration: declarations[0]}
			facts[symbol] = entry
		}
		if !entry.eligible {
			continue
		}

		// A binding touched from a different code path root than its declaration's is out of reach
		// of any liveness pass here. The root, not the lexical scope, is the boundary that matters:
		// a root is what gets its own graph, so an occurrence inside one has no edge to or from any
		// write in another and no dataflow can see the two together.
		//
		// The two directions are NOT the same guard, and collapsing them costs a real diagnostic.
		// Measured against the release binary on the two shapes that separate them:
		//
		//	cross-root READ    `let v = 'used'; console.log(v); setTimeout(() => v = 42, 1);
		//	                    v = 'unused';`
		//	                   The closure only WRITES v, so nothing captured can observe the outer
		//	                   write, and oxlint REPORTS it. Upstream's `has_captured_read` is set
		//	                   from reads alone, so it stays false here.
		//
		//	cross-root WRITE   `let v = 'used'; console.log(v); function bar() { v = 'x'; } bar();
		//	                    console.log(v);`
		//	                   `bar` stores a value the enclosing graph never sees stored, and
		//	                   oxlint is SILENT. Upstream declines it at the report site through
		//	                   `has_same_parent_variable_scope`, comparing the WRITE's own scope
		//	                   against the declaration's rather than consulting the capture flag.
		//
		// So a cross-root read silences every write to the binding, and a cross-root write silences
		// only itself. A first attempt used one flag for both, which fixed the second shape and
		// silently gave up the first.
		root := control_flow_graph.RootOf(identifier)
		occupied[root] = true
		captured := root != control_flow_graph.RootOf(entry.declaration)
		// A declarator with no initializer is classified as a read so the liveness pass errs safe,
		// but it reads nothing, and counting it here made `let provenance: string;` look read. Every
		// write to a binding nothing reads was then reported, once per write, beside the single
		// `no-unused-vars` finding upstream leaves the whole judgment to.
		if occurrenceKindOf(identifier) != occurrenceWrite && !isUninitializedDeclaratorName(identifier) {
			entry.hasRead = true
			if captured {
				entry.capturedRead = true
			}
		}
	}
	return facts, occupied
}

// analyzeRootLiveness runs the backward liveness dataflow over one code path root.
//
// # The fixed point, and why one is needed
//
// A loop back-edge makes a block its own transitive successor, so the live set at the loop header
// depends on the live set at the loop bottom, which depends on the header. Upstream reaches this
// with an explicit loop-header search plus a cached per-loop liveness set — `find_loop_start`,
// `analyze_loop_recursive`, `merge_loop_liveness`, and their four scratch bitsets. Iterating the
// whole transfer function to a fixed point instead reaches the same answer with none of that
// machinery, because a fixed point is what those caches are approximating. The cost is bounded:
// liveness is a monotone union over a finite symbol set, so each round either adds a bit somewhere
// or ends, and the block count bounds the rounds.
func analyzeRootLiveness(ctx rule.Context, root *ast.Node, exported map[string]bool, symbols map[*ast.Symbol]*symbolFacts) {
	if len(symbols) == 0 {
		return
	}

	graph := deadStoreGraph(ctx, root, symbols)

	// Each symbol the events name gets a number in this root, its bit in the live sets below.
	numbers := map[*ast.Symbol]int{}
	for _, block := range graph.Blocks {
		for index := range block.Events {
			event := &block.Events[index]
			number, numbered := numbers[event.symbol]
			if !numbered {
				number = len(numbers)
				numbers[event.symbol] = number
			}
			event.number = number
		}
	}
	words := (len(numbers) + 63) / 64

	// A write laid out in any reachable block is reachable code, whatever other blocks also hold it.
	//
	// # The duplicate-layout trap, and why position-keyed reachability alone is not enough here
	//
	// The graph lays a `finally` block out TWICE, once for normal completion and once for the path
	// leaving the `try` through `return`, `throw`, or a suspended `yield`. Both copies carry the
	// same source positions. `unused_exports.FindUnreachable` hit this first and keyed liveness on source
	// position rather than block identity, because for its question — can this statement run — any
	// reachable copy settles it.
	//
	// That is not sufficient for this rule, and the difference cost a real false positive on an
	// imported clean case. Measured on upstream's clean case 39,
	// `let a; try { foo(); } finally { a = 5; } console.log(a);`, BOTH copies of `a = 5` are
	// reachable: one sits on the normal path and flows to the `console.log(a)` that reads it, and
	// the other sits on the abrupt path and is a terminal block with no successors at all. A
	// per-block verdict reports the terminal copy, because nothing follows it, and the write it
	// names certainly runs and is certainly read.
	//
	// So the reachability question and the liveness question need different keys. Reachability is
	// keyed by position, as `unused` established. Liveness has to be UNIONED across every copy of
	// one position before a write is judged, because the copies are one write in the source and a
	// value read after any copy is a value that write produced.
	reachable := map[int]bool{}
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, event := range block.Events {
			reachable[event.node.Pos()] = true
		}
	}

	// The backward liveness dataflow, run to a fixed point by `control_flow_graph.Solve`.
	//
	// # What moved out of this file, and what did not
	//
	// This used to be a hand-written loop here: iterate blocks by DESCENDING INDEX, rebuild each
	// block's successor union from scratch every round, compare, repeat. Both of those were
	// workarounds for the graph exposing neither an iteration order nor predecessor edges, and both
	// are now the framework's problem rather than this rule's. `Solve` walks reverse postorder
	// reversed, which is the order a backward analysis wants, and it excludes unreachable blocks for
	// the reason spelled out above.
	//
	// The LATTICE did not move. The meet is still the union of the reachable successors' entry sets
	// and the transfer is still `applyBlockTransfer` walking one block's events backward, because
	// those two are what this rule believes about liveness and they are exactly what the framework
	// declines to have an opinion about. What changed is who runs them, not what they say.
	//
	// The verdicts are unchanged and that is checked rather than asserted: a fixed point does not
	// depend on the order it is reached in, and the imported corpus plus the fixtures pin the
	// answers.
	solution := control_flow_graph.Solve[liveSymbols, deadStoreEvent](
		graph,
		control_flow_graph.Backward,
		deadStoreLiveness{reachable: reachable, words: words},
	)

	// The reporting pass runs once the sets have settled, walking each reachable block backward and
	// collecting a verdict per write. A write is named only when EVERY reachable copy of it found
	// its bit dead, which is the union described above expressed as an all-copies-agree test.
	type writeVerdict struct {
		event deadStoreEvent
		dead  bool
	}
	verdicts := map[int]*writeVerdict{}
	var order []int

	// Only reachable blocks are walked, and that costs one shape upstream reports.
	//
	// A write that exists ONLY in an unreachable block is never judged. `function f() { let v = 1;
	// g(v); return v; v = 2; }` puts `v = 2` after the return, in a block nothing arrives at, and
	// the release binary reports it while this stays silent. Measured, with a control case in the
	// same invocation firing to prove the rule was running.
	//
	// It is a decline rather than an oversight. A write control cannot arrive at is not a dead
	// store, it is unreachable code, and that is a different finding with a different repair.
	// Confirmed rather than assumed: `cohere --unused` on that same input reports the statement
	// with the cause spelled out, "an earlier return, throw, break, or continue leaves before this
	// line", which is the sentence a reader can act on. Reporting it here as well would tell them
	// to remove an assignment when the whole line cannot run.
	//
	// The trade is also what keeps the duplicate-layout traps safe. Two mutants weakening this
	// survived every fixture and the whole imported corpus, and the input above is what separates
	// them; both are recorded here rather than left as unexplained survivors, because the honest
	// verdict is a stated scope limit and not an equivalence.
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		// `In` on a Backward solution is the value on EXIT from the block: the symbols live after it.
		// That is what the reporting walk starts from, exactly as `liveOnExit` used to supply it.
		liveAfter, ok := solution.In(block)
		if !ok {
			continue
		}
		current := liveAfter.withWords(words)
		applyBlockTransfer(block, current, reachable, func(event deadStoreEvent, dead bool) {
			position := event.node.Pos()
			existing := verdicts[position]
			if existing == nil {
				verdicts[position] = &writeVerdict{event: event, dead: dead}
				order = append(order, position)
				return
			}
			existing.dead = existing.dead && dead
		})
	}

	for _, position := range order {
		verdict := verdicts[position]
		if verdict.dead {
			reportDeadStore(ctx, verdict.event, exported, symbols)
		}
	}
}

// deadStoreGraph builds one root's graph with an event for each occurrence of an eligible binding.
func deadStoreGraph(ctx rule.Context, root *ast.Node, symbols map[*ast.Symbol]*symbolFacts) *control_flow_graph.Graph[deadStoreEvent] {
	return control_flow_graph.Build(root, control_flow_graph.Hooks[deadStoreEvent]{
		Read: func(builder *control_flow_graph.Builder[deadStoreEvent], node *ast.Node) {
			// The Read hook fires for plain write targets too, so what this occurrence is comes
			// from the AST. A plain write emits nothing here; its store is recorded by the Write
			// hook, which is where it belongs in evaluation order.
			if kind := occurrenceKindOf(node); kind != occurrenceWrite {
				recordDeadStoreEvent(builder, ctx, node, occurrenceRead, symbols)
			}
		},
		Write: func(builder *control_flow_graph.Builder[deadStoreEvent], node *ast.Node) {
			recordDeadStoreEvent(builder, ctx, node, occurrenceWrite, symbols)
		},
	})
}

// deadStoreLiveness is this rule's lattice: a set of symbols, met by union, transferred by walking
// one block's events backward.
//
// # The meet, and why union is the conservative choice here
//
// Upstream distinguishes six edge kinds — Normal, Jump, Backedge, Error, Finalize, Join — and unions
// all but Unreachable into one of two sets, keeping the error set separate so a write bypassed by a
// throw is not reported. Our graph carries no edge kinds, so the same conservatism is obtained by a
// different route: the fork to a handler is a real successor edge, so the handler's live set is
// unioned in like any other, and a write followed by a throwing call stays live because the catch
// path reads the earlier value. Measured on
// `function f() { let v = 1; try { h(); } catch (e) { v = 2; } return v; }`, where the fork out of
// the `try` block reaches the catch and the join before it.
//
// # Bottom and Entry are both empty, and they mean different things
//
// Bottom is the identity for union, so it is the empty set. Entry is the claim that NOTHING is live
// where the code path leaves — no symbol declared inside this root can be read after it returns —
// and it is also the empty set. They coincide numerically and not in meaning, which is why they are
// written separately rather than sharing one constructor.
type deadStoreLiveness struct {
	// reachable is keyed by source position, and `applyBlockTransfer` consults it to decide whether
	// a write is judged at all. It is threaded through the lattice because the transfer needs it and
	// the framework hands the lattice nothing but the block.
	reachable map[int]bool
	// words is how long a live set this root needs, one bit per symbol its events name.
	words int
}

// Bottom is the empty set, nil, which costs nothing. Solve asks for it for every block on every
// round, and the map it used to be was the largest single allocation in this rule (#tmn9n27).
func (deadStoreLiveness) Bottom() liveSymbols { return nil }

func (deadStoreLiveness) Entry() liveSymbols { return nil }

func (l deadStoreLiveness) Meet(left, right liveSymbols) liveSymbols {
	// Neither argument is written: `Solve` holds one of them as a settled boundary value and
	// compares against it to decide whether anything moved. One that already holds the other is
	// returned as it is, which is most meets, since Solve folds every block's neighbours from
	// Bottom. Returning it shares it, and that is safe because nothing writes into a live set it did
	// not just make: Transfer and the reporting walk each copy before they write.
	if left.covers(right) {
		return left
	}
	if right.covers(left) {
		return right
	}
	merged := left.withWords(l.words)
	for index, bits := range right {
		merged[index] |= bits
	}
	return merged
}

func (l deadStoreLiveness) Transfer(
	block *control_flow_graph.Block[deadStoreEvent],
	incoming liveSymbols,
) liveSymbols {
	// A block with no event passes its set through unchanged, so it is handed back rather than
	// copied.
	if len(block.Events) == 0 {
		return incoming
	}
	outgoing := incoming.withWords(l.words)
	// nil observer: the fixed-point rounds must report nothing, because a block is transferred many
	// times before the sets settle. Reporting happens once afterwards, in a separate walk.
	applyBlockTransfer(block, outgoing, l.reachable, nil)
	return outgoing
}

func (deadStoreLiveness) Equal(left, right liveSymbols) bool {
	return left.covers(right) && right.covers(left)
}

// liveSymbols is the set of a root's symbols live at one point: bit n stands for the symbol its
// events number n (deadStoreEvent.number). Nil is the empty set, and a set reads as zero past its
// end, so sets of different lengths compare and combine as the sets they stand for.
//
// A bitset rather than the map it replaced, because a root's symbols are few and fixed once its
// graph is built: one or two words, against a map per block per round (#tmn9n27).
type liveSymbols []uint64

// has reports whether symbol number is in the set.
func (set liveSymbols) has(number int) bool {
	word := number / 64
	return word < len(set) && set[word]&(1<<(number%64)) != 0
}

// covers reports whether every symbol in other is also in set.
func (set liveSymbols) covers(other liveSymbols) bool {
	for index, bits := range other {
		var mine uint64
		if index < len(set) {
			mine = set[index]
		}
		if bits&^mine != 0 {
			return false
		}
	}
	return true
}

// withWords returns a copy of set, words long, which the caller may write without touching set.
func (set liveSymbols) withWords(words int) liveSymbols {
	copied := make(liveSymbols, words)
	copy(copied, set)
	return copied
}

// applyBlockTransfer walks one block's events backward, updating the live set and optionally
// reporting each write it finds dead.
//
// The order within a block is what makes `a = a + 1` work without upstream's deferred-write
// reordering. The graph already emits the right-hand read before the store — measured on
// `let a = 42; f(a); a = 10; a = a + 1; f(a);`, which lays out `R(a@25) R(a@29) W(a@25)`, the
// target's throwable read, the operand read, then the store. Upstream reaches the same ordering by
// pushing a synthetic Read op before a deferred write and carries a `pending_assignment_lhs`
// state machine to do it; here the graph's own evaluation order supplies it.
func applyBlockTransfer(
	block *control_flow_graph.Block[deadStoreEvent],
	current liveSymbols,
	reachable map[int]bool,
	observe func(event deadStoreEvent, dead bool),
) {
	for index := len(block.Events) - 1; index >= 0; index-- {
		event := block.Events[index]
		switch event.kind {
		case occurrenceRead:
			current[event.number/64] |= 1 << (event.number % 64)

		case occurrenceUpdate:
			// An update is two operations at one position: a load and then a store. Read backward
			// that is the store first, then the load.
			//
			// Only the store is handled here, and that is deliberate rather than an omission. The
			// load half is already a separate event: the graph's Read hook fires for an update
			// target as well as for a genuine read, and `analyzeRootLiveness` records every
			// non-plain-write occurrence the hook offers, so an `occurrenceRead` for this same
			// position sits immediately before this event in the block and sets the bit when the
			// backward walk reaches it.
			//
			// Measured rather than argued, because the argument runs the wrong way round. A mutant
			// replacing a set-live here with a clear SURVIVED the whole imported corpus and every
			// fixture, and the first hypothesis — that no fixture wrote a compound assignment — was
			// wrong: three fixtures added for it did not kill the mutant either. The line was
			// subsumed by the Read event, so no input could distinguish the two versions. The
			// inverse mutation, forcing the store's verdict to dead, fails 23 lines, which is what
			// establishes that the arm is reached and that the store half is load-bearing.
			//
			// So the store is judged at the point upstream judges it, before the load restores the
			// bit. `let a = 42; console.log(a); a++;` reports at the increment, and
			// `let v = 1; v += 1; g(v);` keeps `v = 1` live, both confirmed against the release
			// binary.
			if observe != nil && event.reportable && reachable[event.node.Pos()] {
				observe(event, !current.has(event.number))
			}

		case occurrenceWrite:
			if observe != nil && event.reportable && reachable[event.node.Pos()] {
				observe(event, !current.has(event.number))
			}
			current[event.number/64] &^= 1 << (event.number % 64)
		}
	}
}

// recordDeadStoreEvent files one identifier occurrence into the block the walk is currently in.
func recordDeadStoreEvent(
	builder *control_flow_graph.Builder[deadStoreEvent],
	ctx rule.Context,
	node *ast.Node,
	hookKind occurrenceKind,
	symbols map[*ast.Symbol]*symbolFacts,
) {
	if isNonReferenceIdentifier(node) {
		return
	}
	symbol := resolveOccurrence(ctx, node)
	if symbol == nil {
		return
	}
	facts := symbols[symbol]
	if facts == nil || !facts.eligible {
		return
	}

	kind := hookKind
	if hookKind == occurrenceWrite {
		// The Write hook does not distinguish a store that loaded first, so the AST is asked.
		//
		// This is kept for what it SAYS rather than for what it decides, and that is recorded
		// because the honest verdict is equivalence rather than coverage. A mutant forcing this
		// branch off survived every fixture and the whole imported corpus, and five constructed
		// inputs across `+=`, `++` twice over, an update killed by a later write, and `||=` produced
		// byte-identical findings. The reason is structural: the graph's Read hook fires for an
		// update target too, so `analyzeRootLiveness` has already emitted an `occurrenceRead` at
		// this same position, and a backward walk over `Read, Update` and over `Read, Write` reaches
		// the same live set by the same two steps.
		//
		// It stays because the classification is the one the rest of the file reasons in, and a
		// later change to what the Read hook is filtered on would make the difference real without
		// anything else moving. An equivalent line that documents its own equivalence is cheaper
		// than a reader rediscovering why the kinds disagree.
		if occurrenceKindOf(node) == occurrenceUpdate {
			kind = occurrenceUpdate
		}
	}

	builder.Emit(deadStoreEvent{
		node:   node,
		symbol: symbol,
		kind:   kind,
		// An update's store half is as reportable as a plain write's; only a read is never a
		// candidate. The event this flag rides on is emitted by the Write hook in both cases, so a
		// read event never carries it regardless.
		reportable: kind != occurrenceRead && isReportableWrite(node, facts.declaration),
	})
}

// reportDeadStore applies the per-binding guards and names the write.
//
// The guards run here rather than at collection because a write inside a `try` block must still
// kill an earlier write on the normal path while never being reported itself. Dropping such an
// event entirely would make the write it overwrites look live, which costs the five diagnostics the
// previous implementation declined for exactly this reason.
func reportDeadStore(
	ctx rule.Context,
	event deadStoreEvent,
	exported map[string]bool,
	symbols map[*ast.Symbol]*symbolFacts,
) {
	facts := symbols[event.symbol]
	if facts == nil || !facts.eligible {
		return
	}
	if !facts.hasRead || facts.capturedRead {
		return
	}
	if exported[event.symbol.Name] {
		return
	}
	ctx.ReportNode(event.node, messageNoUselessAssignment)
}

// occurrenceKindOf classifies one identifier occurrence from the AST alone.
//
// This is the discrimination the Read hook cannot make. Three answers, and each one changes the
// liveness transfer: a read sets the bit, a plain write clears it, an update does both in the order
// that leaves it set.
func occurrenceKindOf(identifier *ast.Node) occurrenceKind {
	parent := identifier.Parent
	if parent == nil {
		return occurrenceRead
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return occurrenceRead
		}
		if ast.SkipParentheses(binary.Left) != identifier {
			return occurrenceRead
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return occurrenceWrite
		}
		// A logical assignment (`||=`, `&&=`, `??=`) reads the previous value to decide whether to
		// store at all, so it is an update rather than a plain write, and the store is conditional
		// on top of that. Both make the earlier write live.
		return occurrenceUpdate

	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration.Name() != identifier {
			return occurrenceRead
		}
		if declaration.Initializer == nil {
			// A declarator with no initializer stores nothing. It is not a read either, but the
			// liveness pass has no third answer that means "neither", and treating it as a read is
			// the safe direction: it can only keep a write live.
			return occurrenceRead
		}
		return occurrenceWrite

	case ast.KindPrefixUnaryExpression:
		operator := parent.AsPrefixUnaryExpression().Operator
		if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			return occurrenceUpdate
		}
		return occurrenceRead

	case ast.KindPostfixUnaryExpression:
		return occurrenceUpdate

	case ast.KindBindingElement:
		// A destructured binding, `let [a, b] = arr` or `let { a } = obj`. The graph emits a real
		// Write for it — measured on `let [a, b] = arr; a = 3; console.log(a, b);`, which lays out
		// `W(a@24)` at the binding — and upstream reports the binding when a later write kills it.
		// The previous implementation declined the whole shape because a pattern element's flow
		// node did not line up with a single-identifier reachability question; the liveness pass
		// has no such difficulty, since the graph already placed the write where it happens.
		//
		// Only the element's own name is the write. A `PropertyName` is the key being read out of
		// the object and names no variable, and a `default` initializer is an ordinary expression.
		binding := parent.AsBindingElement()
		if binding == nil || binding.Name() != identifier {
			return occurrenceRead
		}
		if binding.DotDotDotToken != nil {
			// A rest element binds a fresh array or object every time, so it is a write like any
			// other. Kept explicit rather than folded in, because the name check above is what
			// separates it from the property key and a reader should see that it was considered.
			return occurrenceWrite
		}
		return occurrenceWrite
	}

	return occurrenceRead
}

// isUninitializedDeclaratorName reports whether an identifier is the name of a declarator that has
// no initializer, `let v;`, which neither stores nor loads anything.
func isUninitializedDeclaratorName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := parent.AsVariableDeclaration()
	return declaration.Name() == identifier && declaration.Initializer == nil
}

// writeTargetOfIncludingUpdates reports whether this identifier is written, by any form.
//
// Wider than `writeTargetOf`, which the inverted implementation kept narrow because an update
// expression's flow node did not line up with its reachability question. The liveness pass has no
// such restriction, so the name filter is widened to match what the pass can actually judge.
func writeTargetOfIncludingUpdates(identifier *ast.Node) (*ast.Node, bool) {
	if occurrenceKindOf(identifier) == occurrenceRead {
		return nil, false
	}
	return identifier, true
}

// isReportableWrite reports whether a write is one this rule may name, as opposed to one that only
// participates in the dataflow.
//
// Two declines, and both are upstream's own.
//
// A write inside a `try` block is never named. A try block has an edge out of every operation
// rather than only its end, so a write there may be abandoned partway by a throw from itself or
// from anything after it, and the value it overwrote is still observable on the catch and finally
// paths. Upstream declines the same writes under `is_in_try_block`, reading the block's outgoing
// error and finalize edges; our graph carries no edge kinds, so the same set is identified from the
// syntax that produced it, which is where the property comes from in the first place.
//
// A write in a different code path root than its own declaration is never named either, because
// the root holding the declaration is the only one whose liveness can see what follows the write.
// This is upstream's `has_same_parent_variable_scope`, tested at the report site on every write.
// The declaration argument is what lets this be per-write rather than per-binding, which is the
// distinction `collectSymbolFacts` records the measurement for.
func isReportableWrite(node *ast.Node, declaration *ast.Node) bool {
	if isInsideTryBlock(node) {
		return false
	}
	if control_flow_graph.RootOf(node) != control_flow_graph.RootOf(declaration) {
		return false
	}
	return !isWriteNestedInOwnDestructuring(node, declaration)
}

// isWriteNestedInOwnDestructuring reports whether a write to a destructured binding sits inside the
// very declarator that binds it.
//
// Upstream declines these outright rather than reporting them: `process_reference_deferred` returns
// before pushing any op when the declarator's binding is a pattern and the reference's span is
// contained in the declarator's own span. The shape is upstream's failing case 32,
// `let { a, b = (a = 2) } = obj; a = 3; console.log(a, b);`, where the default initializer for `b`
// assigns `a`. Measured against the release binary: oxlint reports one finding at the `a` binding
// and says nothing about the `(a = 2)`, and reporting both was a genuine overcount this rule
// produced before the decline was reproduced.
//
// The reason it is a decline rather than a liveness answer is evaluation order inside a pattern:
// whether `b`'s default runs at all depends on the incoming value, so a write there is conditional
// in a way the declarator's single position cannot express.
func isWriteNestedInOwnDestructuring(node *ast.Node, declaration *ast.Node) bool {
	// Walk to the declarator this binding belongs to, if it is a destructured one.
	declarator := declaration
	for declarator != nil && (declarator.Kind == ast.KindBindingElement ||
		declarator.Kind == ast.KindObjectBindingPattern ||
		declarator.Kind == ast.KindArrayBindingPattern) {
		declarator = declarator.Parent
	}
	if declarator == nil || declarator.Kind != ast.KindVariableDeclaration {
		return false
	}
	if name := declarator.AsVariableDeclaration().Name(); name == nil ||
		(name.Kind != ast.KindObjectBindingPattern && name.Kind != ast.KindArrayBindingPattern) {
		return false
	}
	// The binding's own name is the declarator's write and stays reportable; only a write nested
	// elsewhere inside the declarator is declined.
	if node == declaration || (declaration.Kind == ast.KindBindingElement && declaration.Name() == node) {
		return false
	}
	for current := node; current != nil; current = current.Parent {
		if current == declarator {
			return true
		}
	}
	return false
}
