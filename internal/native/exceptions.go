package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strconv"
	"strings"
)

// Exceptions by cleanup paths (docs/memory.md, "Exceptions, designed into counting"). The runtime
// has one pending-exception word, adamic_thrown. A throw sets it and jumps to its handler: the
// innermost try around it in this function, letting go of what the statement and every scope between
// hold, or the function's way out, letting go of everything the frame holds and returning a zero
// value. After a call to a function that can throw, the word is tested and the same jump made. Out of
// main, an error is a panic. A finally is emitted at every way out of its try.

// handler is one try being emitted: where a throw inside it lands, the scope depth its body starts
// at, and its finally, which every way out of it runs.
type handler struct {
	label      string
	depth      int
	finally    []ir.Statement
	hasFinally bool

	// jumped says something jumps to label: a try whose body can't throw has no landing to emit.
	jumped bool
}

// jumpThrown emits the jump for a pending exception, once the statement's temporaries are let go: to
// the innermost handler, or out of the function, or, out of main, the panic.
func (nativeEmitter *emitter) jumpThrown() {
	if nativeEmitter.asyncHandler != "" {
		nativeEmitter.line("adamic_release(frame->error); frame->error = adamic_thrown; adamic_thrown = NULL;")
		nativeEmitter.line("goto %s;", nativeEmitter.asyncHandler)
		return
	}
	if len(nativeEmitter.handlers) > 0 {
		inner := nativeEmitter.handlers[len(nativeEmitter.handlers)-1]
		inner.jumped = true
		nativeEmitter.releaseScopes(inner.depth)
		nativeEmitter.line("goto %s;", inner.label)
		return
	}
	if nativeEmitter.function == nil {
		// Out of main: String(error) as a panic, stdout flushed first. What the program holds then
		// isn't let go, as with any panic.
		nativeEmitter.line("adamic_uncaught();")
		return
	}
	nativeEmitter.releaseScopes(nativeEmitter.functionDepth)
	switch {
	case nativeEmitter.function.Closure:
		// A function value returns a slot whatever its type; the caller tests the word before reading it.
		nativeEmitter.line("return (adamic_value){.number = 0};")
	case nativeEmitter.function.Returns == 0:
		nativeEmitter.line("return;")
	default:
		nativeEmitter.line("return %s;", zero(nativeEmitter.function.Returns))
	}
}

// checkThrown emits, after a call that can throw, the test of the pending word and, when it's set,
// the release of what holds names (what a loop around the call holds), then of every temporary the
// statement owns so far, and the jump. In a statement, what comes after the holds can be the
// statement's ladder instead (settleThrows), and which it is waits for the statement's end: the
// check leaves a marker where its text goes.
func (nativeEmitter *emitter) checkThrown(holds ...string) {
	// Async graph edges use the precise call-target proof. Some ordinary library
	// emitters conservatively request a check whenever any closure can throw.
	// A block without an exception edge must not invent a Goto handler for it.
	if nativeEmitter.asyncSlots != nil && nativeEmitter.asyncHandler == "" {
		message := nativeEmitter.function.Name + ": exception where the flow graph proved none"
		nativeEmitter.line("if (adamic_thrown != NULL) { adamic_panic(%s, %d); }", cString(message), len(message))
		return
	}
	// An async function's way out writes frame->error and jumps to its handler, and has no ladder;
	// neither has code outside a statement, nor a list the statement doesn't own (placeholder.go).
	if nativeEmitter.ladder == nil ||
		nativeEmitter.asyncSlots != nil ||
		len(nativeEmitter.outerOwned) < nativeEmitter.ladder.depth {
		nativeEmitter.out.WriteString(nativeEmitter.thrownSite(holds, len(nativeEmitter.outerOwned)).inline)
		return
	}
	site := nativeEmitter.thrownSite(holds, nativeEmitter.ladder.depth)
	nativeEmitter.ladder.sites = append(nativeEmitter.ladder.sites, site)
	nativeEmitter.ladder.batch = append(nativeEmitter.ladder.batch, site)
	fmt.Fprintf(&nativeEmitter.out, "\x00%d:%d\x00", nativeEmitter.ladder.serial, len(nativeEmitter.ladder.sites)-1)
}

// thrownSite is one check's way out, in its parts. depth is the aside the statement's own list is
// in: the lists of asides inside it, what a throw from inside one lets go of first, are the site's
// own, and the lists outside it are part of the rest of the way out.
func (nativeEmitter *emitter) thrownSite(holds []string, depth int) *throwSite {
	site := &throwSite{indent: nativeEmitter.indent}
	statement := nativeEmitter.owned
	if len(nativeEmitter.outerOwned) > depth {
		statement = nativeEmitter.outerOwned[depth]
	}
	// A copy: kept and taken delete from the list in place, and the site lets go of what the
	// statement owned when it ran.
	site.owned = slices.Clone(statement)
	site.before = nativeEmitter.captured(nativeEmitter.indent+1, func() {
		for _, hold := range holds {
			nativeEmitter.line("adamic_release(%s);", hold)
		}
		if len(nativeEmitter.outerOwned) > depth {
			for index := len(nativeEmitter.owned) - 1; index >= 0; index-- {
				nativeEmitter.line("adamic_release(%s);", nativeEmitter.owned[index])
			}
			for outer := len(nativeEmitter.outerOwned) - 1; outer > depth; outer-- {
				for index := len(nativeEmitter.outerOwned[outer]) - 1; index >= 0; index-- {
					nativeEmitter.line("adamic_release(%s);", nativeEmitter.outerOwned[outer][index])
				}
			}
		}
	})
	site.way = nativeEmitter.captured(0, func() {
		for outer := min(depth, len(nativeEmitter.outerOwned)) - 1; outer >= 0; outer-- {
			for index := len(nativeEmitter.outerOwned[outer]) - 1; index >= 0; index-- {
				nativeEmitter.line("adamic_release(%s);", nativeEmitter.outerOwned[outer][index])
			}
		}
		if nativeEmitter.statementRegion != "" {
			// The statement's region ends on this way out of it too, or its blocks, and what its values
			// hold on the heap, leak (region.go).
			nativeEmitter.line("adamic_region_end(&%s);", nativeEmitter.statementRegion)
		}
		nativeEmitter.jumpThrown()
	})
	site.inline = nativeEmitter.captured(nativeEmitter.indent, func() {
		nativeEmitter.line("if (adamic_thrown != NULL) {")
		nativeEmitter.out.WriteString(site.before)
		nativeEmitter.indent++
		for index := len(site.owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", site.owned[index])
		}
		nativeEmitter.out.WriteString(indented(site.way, nativeEmitter.indent))
		nativeEmitter.indent--
		nativeEmitter.line("}")
	})
	return site
}

// A statement with several checks would let go of the same temporaries at each, the later ones of
// more of them, so a long expression's checks grew as the square of its length. A statement's checks
// share a ladder instead, written where the statement ends, after its own temporaries are let go of:
//
//	goto done;
//	rung_0_3:;   release(c);   a check that ran owning a, b and c enters here,
//	rung_0_1:;   release(a);   one that ran owning a alone, here,
//	out_0:;      the lists of asides outside the statement, its region's end, and the jump
//	done:;
//
// Each check keeps its holds and its asides' lists, and jumps in. A chain of rungs is one list, and a
// check joins a chain whose list starts with all of its own: a temporary moved out (kept, taken,
// handedOver) or never made on its path isn't in its list, so it starts a chain of its own and never
// lets go of what it doesn't own. The ladder is written only when it's shorter than the checks
// written in place, so a statement with one check reads as it always has.

// ladder is the statement being emitted's checks: depth is how many asides were open when it began,
// sites every check in it, by the marker each left, and batch those not yet settled.
type ladder struct {
	depth int
	sites []*throwSite
	batch []*throwSite

	// serial names the ladder in its markers. A marker carries its ladder's serial as well as its
	// check's index, so a statement replaces only its own: text an enclosing statement wrote, should
	// it ever reach a nested statement's buffer, passes through to the ladder it belongs to, and is
	// never read as a check of the nested statement's that has the same index.
	serial int
}

// throwSite is one check: the indent it's at, what it lets go of first (its holds and its asides'
// lists), the statement's own list as it ran, the rest of its way out written at indent 0 (outer
// lists, the region's end and the jump), its text in place, and the text its marker becomes.
type throwSite struct {
	indent int
	before string
	owned  []string
	way    string
	inline string
	text   string

	// chain and rung are where it enters the ladder: rung is how many of the chain's list it lets go
	// of, and a check owning nothing enters at its way out, chain -1.
	chain int
	rung  int
}

// ladderChain is one list of rungs, longest first, and the way out it falls into.
type ladderChain struct {
	owned   []string
	way     int
	entries map[int]bool
}

// ladderFrame is what a statement's ladder stands in for while the statement is emitted: the
// enclosing statement's, and the text before it.
type ladderFrame struct {
	ladder *ladder
	out    strings.Builder
}

// openLadder starts a statement's ladder, its text in a buffer of its own so its markers can be
// replaced.
func (nativeEmitter *emitter) openLadder() ladderFrame {
	frame := ladderFrame{ladder: nativeEmitter.ladder, out: nativeEmitter.out}
	nativeEmitter.ladders++
	nativeEmitter.ladder, nativeEmitter.out = &ladder{
		depth:  len(nativeEmitter.outerOwned),
		serial: nativeEmitter.ladders,
	}, strings.Builder{}
	return frame
}

// closeLadder ends a statement's ladder: a check its end never settled (a statement that ends with a
// panic, or checks after its last release) is written in place, and every marker becomes its text.
func (nativeEmitter *emitter) closeLadder(frame ladderFrame) {
	ladder := nativeEmitter.ladder
	for _, site := range ladder.batch {
		site.text = site.inline
	}
	text := nativeEmitter.out.String()
	nativeEmitter.ladder, nativeEmitter.out = frame.ladder, frame.out
	if len(ladder.sites) == 0 {
		nativeEmitter.out.WriteString(text)
		return
	}
	for {
		start := strings.IndexByte(text, 0)
		if start < 0 {
			nativeEmitter.out.WriteString(text)
			return
		}
		length := strings.IndexByte(text[start+1:], 0)
		if length < 0 {
			panic("native: compiler bug: an unterminated throw marker")
		}
		serial, index, throwMarkerError := throwMarker(text[start+1 : start+1+length])
		if throwMarkerError != nil {
			panic("native: compiler bug: a throw marker that doesn't read: " + throwMarkerError.Error())
		}
		if serial != ladder.serial {
			// An enclosing statement's check: its marker passes through for that statement's own end.
			// With no enclosing ladder, nothing would ever replace it, so it's a compiler bug.
			if frame.ladder == nil {
				panic("native: compiler bug: a throw marker of no open statement")
			}
			nativeEmitter.out.WriteString(text[:start+length+2])
			text = text[start+length+2:]
			continue
		}
		if index < 0 || index >= len(ladder.sites) {
			panic("native: compiler bug: a throw marker of no check in its statement")
		}
		nativeEmitter.out.WriteString(text[:start])
		nativeEmitter.out.WriteString(ladder.sites[index].text)
		text = text[start+length+2:]
	}
}

// unmarked is a program's or a unit's finished C, which no throw marker may reach: every one is
// replaced when its statement closes, and one left over would be a check with no text, compiled as
// a stray byte or, worse, read past.
func unmarked(text string) string {
	if strings.IndexByte(text, 0) >= 0 {
		panic("native: compiler bug: a throw marker reached the C")
	}
	return text
}

// throwMarker reads a marker's "serial:index".
func throwMarker(text string) (int, int, error) {
	serialText, indexText, found := strings.Cut(text, ":")
	if !found {
		return 0, 0, fmt.Errorf("%q has no serial", text)
	}
	serial, parseError := strconv.Atoi(serialText)
	if parseError != nil {
		return 0, 0, parseError
	}
	index, parseError := strconv.Atoi(indexText)
	return serial, index, parseError
}

// settleThrows decides, where the statement's own temporaries have just been let go of, between the
// ladder for its checks so far and their text in place, whichever is shorter (or the ladder whenever
// there's one, under Options.LadderEveryStatement), and writes the ladder here if it's the one.
func (nativeEmitter *emitter) settleThrows() {
	ladder := nativeEmitter.ladder
	// An end inside an aside's work (an expression's effects) isn't the statement's, and its
	// checks wait for the statement's own.
	if ladder == nil || len(ladder.batch) == 0 || len(nativeEmitter.outerOwned) != ladder.depth {
		return
	}
	batch := ladder.batch
	ladder.batch = nil
	ways, chains := ladderPlan(batch)
	inline, shared := 0, strings.Count(nativeEmitter.ladderText("", ways, chains, batch), "\n")
	for _, site := range batch {
		inline += strings.Count(site.inline, "\n")
		shared += strings.Count(site.jump(""), "\n")
	}
	if shared >= inline && !nativeEmitter.ladderEveryStatement {
		for _, site := range batch {
			site.text = site.inline
		}
		return
	}
	base := nativeEmitter.temporary()
	nativeEmitter.out.WriteString(nativeEmitter.ladderText(base, ways, chains, batch))
	for _, site := range batch {
		site.text = site.jump(ladderEntry(base, site.chain, site.rung, site.way, ways))
	}
}

// ladderPlan puts each check on a chain: the ways out are the distinct rests of the way (a hold
// taken in the statement changes the scopes a later check lets go of), and a check joins a chain
// with the same way out whose list starts with its own, or one whose whole list starts its own,
// which it lengthens, or else starts one.
func ladderPlan(batch []*throwSite) ([]string, []*ladderChain) {
	var ways []string
	var chains []*ladderChain
	for _, site := range batch {
		way := slices.Index(ways, site.way)
		if way < 0 {
			way = len(ways)
			ways = append(ways, site.way)
		}
		site.chain, site.rung = -1, len(site.owned)
		if len(site.owned) == 0 {
			continue
		}
		for index, chain := range chains {
			if chain.way == way && (startsWith(chain.owned, site.owned) || startsWith(site.owned, chain.owned)) {
				site.chain = index
				break
			}
		}
		if site.chain < 0 {
			site.chain = len(chains)
			chains = append(chains, &ladderChain{way: way, entries: map[int]bool{}})
		}
		chain := chains[site.chain]
		if len(site.owned) > len(chain.owned) {
			chain.owned = site.owned
		}
		chain.entries[site.rung] = true
	}
	return ways, chains
}

// startsWith says list begins with every name of prefix, in order.
func startsWith(list, prefix []string) bool {
	return len(prefix) <= len(list) && slices.Equal(list[:len(prefix)], prefix)
}

// ladderEntry is the label a check jumps to: its rung on its chain, or its way out.
func ladderEntry(base string, chain, rung int, way string, ways []string) string {
	if chain < 0 {
		return fmt.Sprintf("%s_out_%d", base, slices.Index(ways, way))
	}
	return fmt.Sprintf("%s_rung_%d_%d", base, chain, rung)
}

// ladderText is the ladder at the emitter's indent, its labels named from base. Each way out
// follows its chains, the last falling into it and the others jumping; a way out ends in a jump, a
// return or adamic_uncaught, so nothing falls from one into the next.
func (nativeEmitter *emitter) ladderText(base string, ways []string, chains []*ladderChain, batch []*throwSite) string {
	return nativeEmitter.captured(nativeEmitter.indent, func() {
		nativeEmitter.line("goto %s_done;", base)
		for way, text := range ways {
			jumped := slices.ContainsFunc(
				batch,
				func(site *throwSite) bool { return site.chain < 0 && site.way == text },
			)
			var own []*ladderChain
			for _, chain := range chains {
				if chain.way == way {
					own = append(own, chain)
				}
			}
			for index, chain := range own {
				number := slices.Index(chains, chain)
				for rung := len(chain.owned); rung >= 1; rung-- {
					if chain.entries[rung] {
						nativeEmitter.line("%s_rung_%d_%d:;", base, number, rung)
					}
					nativeEmitter.line("adamic_release(%s);", chain.owned[rung-1])
				}
				if index < len(own)-1 {
					nativeEmitter.line("goto %s_out_%d;", base, way)
					jumped = true
				}
			}
			if jumped {
				nativeEmitter.line("%s_out_%d:;", base, way)
			}
			nativeEmitter.out.WriteString(indented(text, nativeEmitter.indent))
		}
		nativeEmitter.line("%s_done:;", base)
	})
}

// jump is the check's text when it enters the ladder at label.
func (site *throwSite) jump(label string) string {
	tabs := strings.Repeat("\t", site.indent)
	if site.before == "" {
		return tabs + "if (adamic_thrown != NULL) goto " + label + ";\n"
	}
	return tabs + "if (adamic_thrown != NULL) {\n" + site.before + tabs + "\tgoto " + label + ";\n" + tabs + "}\n"
}

// captured is what emit writes at indent, kept rather than written where the emitter stands.
func (nativeEmitter *emitter) captured(indent int, emit func()) string {
	saved, savedIndent := nativeEmitter.out, nativeEmitter.indent
	nativeEmitter.out, nativeEmitter.indent = strings.Builder{}, indent
	emit()
	text := nativeEmitter.out.String()
	nativeEmitter.out, nativeEmitter.indent = saved, savedIndent
	return text
}

// indented is text written at indent 0, moved in by depth tabs.
func indented(text string, depth int) string {
	if depth == 0 {
		return text
	}
	tabs := strings.Repeat("\t", depth)
	var builder strings.Builder
	for _, line := range strings.SplitAfter(text, "\n") {
		if line != "" {
			builder.WriteString(tabs)
			builder.WriteString(line)
		}
	}
	return builder.String()
}

// closureThrown is checkThrown after a call through a function value, written out or made by a loop
// the emitter writes (map, the visits, reduce, Array.from), when a function value in the program can
// throw: which one the call reaches isn't known. holds is what the loop holds across the call.
func (nativeEmitter *emitter) closureThrown(holds ...string) {
	if nativeEmitter.program.ClosuresMayThrow {
		nativeEmitter.checkThrown(holds...)
	}
}

// throwNamed throws a new Error whose name is class and whose message is message, as Node throws one
// of its own (a ReferenceError, a TypeError): String(error) reads "class: message", and a catch takes it.
func (nativeEmitter *emitter) throwNamed(class, message string) {
	nativeEmitter.line("static adamic_string error_message = ADAMIC_STRING(%s);", cString(message))
	nativeEmitter.line("static adamic_string error_name = ADAMIC_STRING(%s);", cString(class))
	nativeEmitter.line("adamic_thrown = adamic_error_new(&error_message);")
	nativeEmitter.line("adamic_release(adamic_thrown->slots[0].reference);")
	nativeEmitter.line("adamic_thrown->slots[0].reference = adamic_retain(&error_name);")
	nativeEmitter.checkThrown()
}

// throwStatement emits throw: the error is the pending word's, and the jump is made.
func (nativeEmitter *emitter) throwStatement(statement ir.Throw) {
	value := nativeEmitter.value(statement.Value)
	if nativeEmitter.asyncSlots != nil {
		// The pending exception takes the statement's owner before cleanup.
		nativeEmitter.line("adamic_thrown = %s;", nativeEmitter.kept(value))
	} else {
		nativeEmitter.line("adamic_thrown = adamic_retain(%s);", value)
	}
	nativeEmitter.end()
	nativeEmitter.jumpThrown()
}

// finallies emits, for a way out that leaves the tries open from index on, each one's finally,
// innermost first, each with only the tries outside it open, so a throw from it goes on outward.
func (nativeEmitter *emitter) finallies(from int) {
	saved := nativeEmitter.handlers
	for index := len(saved) - 1; index >= from; index-- {
		if !saved[index].hasFinally {
			continue
		}
		nativeEmitter.handlers = saved[:index]
		nativeEmitter.line("{")
		nativeEmitter.nested(saved[index].finally, nil)
		nativeEmitter.line("}")
	}
	nativeEmitter.handlers = saved
}

// innerHandlers is the index of the first open try inside scope depth: the ones a break or continue
// to a loop at that depth leaves.
func (nativeEmitter *emitter) innerHandlers(depth int) int {
	index := len(nativeEmitter.handlers)
	for index > 0 && nativeEmitter.handlers[index-1].depth > depth {
		index--
	}
	return index
}

// tryStatement emits try, catch and finally:
//
//	{ body }            a throw lands at landing
//	{ finally }         the body finished
//	goto done;
//	landing:;
//	{ error = adamic_thrown; catch }   a throw lands at rethrow (with a finally) or goes outward
//	{ finally }         the catch finished
//	goto done;
//	rethrow:;           with a finally: the pending error held while finally runs, then on outward
//	done:;
func (nativeEmitter *emitter) tryStatement(statement ir.Try) {
	name := nativeEmitter.temporary()
	landing, rethrow, done := name+"_landing", name+"_rethrow", name+"_done"
	if !statement.HasCatch {
		// Without a catch, a throw in the body goes straight to the finally that runs on the way out.
		landing = rethrow
	}

	body := &handler{
		label:      landing,
		depth:      len(nativeEmitter.scopes),
		finally:    statement.Finally,
		hasFinally: statement.HasFinally,
	}
	nativeEmitter.handlers = append(nativeEmitter.handlers, body)
	nativeEmitter.line("{")
	nativeEmitter.nested(statement.Body, nil)
	nativeEmitter.line("}")
	nativeEmitter.handlers = nativeEmitter.handlers[:len(nativeEmitter.handlers)-1]
	if statement.HasFinally {
		nativeEmitter.line("{")
		nativeEmitter.nested(statement.Finally, nil)
		nativeEmitter.line("}")
	}
	nativeEmitter.line("goto %s;", done)

	// A catch nothing can land in, or a rethrow nothing jumps to, isn't emitted: clang refuses a label
	// nothing uses, and the code would be dead.
	rethrown := !statement.HasCatch && body.jumped
	if statement.HasCatch && body.jumped {
		nativeEmitter.line("%s:;", landing)
		caughtHandler := &handler{
			label:      rethrow,
			depth:      len(nativeEmitter.scopes),
			finally:    statement.Finally,
			hasFinally: true,
		}
		if statement.HasFinally {
			nativeEmitter.handlers = append(nativeEmitter.handlers, caughtHandler)
		}
		nativeEmitter.line("{")
		nativeEmitter.indent++
		nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
		caught := nativeEmitter.temporary()
		// What was thrown is the catch's now; with nothing bound to it, it's let go at once.
		nativeEmitter.line("adamic_object *%s = adamic_thrown;", caught)
		nativeEmitter.line("adamic_thrown = NULL;")
		if statement.CatchLocal >= 0 {
			nativeEmitter.declareLocal(statement.CatchLocal, caught, true)
		} else {
			nativeEmitter.line("adamic_release(%s);", caught)
		}
		for _, inner := range statement.Catch {
			nativeEmitter.statement(inner)
		}
		nativeEmitter.releaseScopes(len(nativeEmitter.scopes) - 1)
		nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
		nativeEmitter.indent--
		nativeEmitter.line("}")
		if statement.HasFinally {
			nativeEmitter.handlers = nativeEmitter.handlers[:len(nativeEmitter.handlers)-1]
			rethrown = caughtHandler.jumped
			nativeEmitter.line("{")
			nativeEmitter.nested(statement.Finally, nil)
			nativeEmitter.line("}")
		}
		nativeEmitter.line("goto %s;", done)
	}

	if statement.HasFinally && rethrown {
		nativeEmitter.line("%s:;", rethrow)
		nativeEmitter.line("{")
		nativeEmitter.indent++
		nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
		// The pending error waits in a scope of the finally's while it runs, so a throw from the
		// finally, which replaces it, lets go of it on the way out.
		pending := nativeEmitter.temporary()
		nativeEmitter.line("adamic_object *%s = adamic_thrown;", pending)
		nativeEmitter.line("adamic_thrown = NULL;")
		nativeEmitter.hold(pending)
		nativeEmitter.line("{")
		nativeEmitter.nested(statement.Finally, nil)
		nativeEmitter.line("}")
		nativeEmitter.line("adamic_thrown = %s;", pending)
		nativeEmitter.line("%s = NULL;", pending)
		nativeEmitter.jumpThrown()
		nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
		nativeEmitter.indent--
		nativeEmitter.line("}")
	}
	nativeEmitter.line("%s:;", done)
}
