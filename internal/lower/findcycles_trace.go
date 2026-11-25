package lower

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A scratch instrument for #c0eypws, never for main: with ADAMIC_FINDCYCLES_TRACE set, findCycles writes to
// stderr, each line prefixed FINDCYCLES, how many relation queries the cycle finder asks at each call site, how
// many reachLinks calls and distinct proven types it was asked for, and how the checker's diagnostics collection
// grows under it. Unset, every hook is a nil receiver and returns at once.

// findCyclesTraceOn is read once, so an unset variable costs one nil check per hook.
var findCyclesTraceOn = os.Getenv("ADAMIC_FINDCYCLES_TRACE") != ""

// findCyclesArm is one IsTypeAssignableTo call site inside the cycle finder.
type findCyclesArm int

const (
	findCyclesArmReachStatics findCyclesArm = iota
	findCyclesArmReachClosures
	findCyclesArmReachAccessorCaptures
	findCyclesArmReachShapes
	findCyclesArmRelatedForward
	findCyclesArmRelatedBack
	findCyclesArmGraphStatics
	findCyclesArmGraphClosures
	findCyclesArmGraphShapes
	findCyclesArmGraphAccessorCaptures
	findCyclesArmViewPairFunction
	findCyclesArmUnplannedPair
	findCyclesArmUnplannedNodeArray
	findCyclesArmCount
)

var findCyclesArmNames = [findCyclesArmCount]string{
	"reachStatics(cycles.go:855)",
	"reachClosures(cycles.go:863)",
	"reachAccessorCaptures(cycles.go:883)",
	"reachShapes(cycles.go:900)",
	"relatedForward(cycles.go:919)",
	"relatedBack(cycles.go:919)",
	"graphStatics(graph_types.go:467)",
	"graphClosures(graph_types.go:473)",
	"graphShapes(graph_types.go:496)",
	"graphAccessorCaptures(graph_types.go:501)",
	"viewPairFunction(graph_types.go:789)",
	"unplannedPair(graph_types.go:1038)",
	"unplannedNodeArray(graph_types.go:1114)",
}

type findCyclesTrace struct {
	checker *checker.Checker
	started time.Time
	printed time.Time

	queries   [findCyclesArmCount]int
	total     int
	growths   int
	grown     [findCyclesArmCount]int
	reachLink int
	proven    map[*checker.Type]bool
	resets    int

	// diagnostics is the checker's diagnostics collection, read by reflection as load.RelationOverflow reads it,
	// and count its unexported count, which Add raises once per diagnostic it keeps.
	diagnostics reflect.Value
	count       *int
}

// startFindCyclesTrace is the trace for one findCycles, or nil when the variable is unset.
func startFindCyclesTrace(l *lowering) *findCyclesTrace {
	if !findCyclesTraceOn {
		return nil
	}
	trace := &findCyclesTrace{checker: l.checker, started: time.Now(), proven: map[*checker.Type]bool{}}
	trace.printed = trace.started
	field := reflect.ValueOf(l.checker).Elem().FieldByName("diagnostics")
	if field.IsValid() {
		trace.diagnostics = field
		if count := field.FieldByName("count"); count.IsValid() && count.Kind() == reflect.Int {
			trace.count = (*int)(unsafe.Pointer(count.UnsafeAddr()))
		}
	}
	groups := groupClosures(l)
	fmt.Fprintf(os.Stderr, "FINDCYCLES start at=%s closureRecords=%d capturingClosures=%d capturingTypes=%d statics=%d accessorCaptures=%d diagnostics=%d\n", trace.started.Format(time.RFC3339Nano), len(l.closureRecords), len(groups.captures), len(groups.types), len(l.statics), len(l.accessorCaptures), trace.diagnosticCount())
	return trace
}

func (trace *findCyclesTrace) diagnosticCount() int {
	if trace.count == nil {
		return -1
	}
	return *trace.count
}

// collisions is how many location keys the collection holds more than one diagnostic under, and the longest such
// list, which Add scans in full for every diagnostic added at that key.
func (trace *findCyclesTrace) collisions() (keys int, longest int) {
	if !trace.diagnostics.IsValid() {
		return -1, -1
	}
	table := trace.diagnostics.FieldByName("diagnosticCollisions")
	if !table.IsValid() || table.Kind() != reflect.Map {
		return -1, -1
	}
	for iterator := table.MapRange(); iterator.Next(); {
		keys++
		if length := iterator.Value().Len(); length > longest {
			longest = length
		}
	}
	return keys, longest
}

// assignablePairs is the size of the checker's assignable relation cache.
func (trace *findCyclesTrace) assignablePairs() int {
	relation := reflect.ValueOf(trace.checker).Elem().FieldByName("assignableRelation")
	if !relation.IsValid() || relation.Kind() != reflect.Pointer || relation.IsNil() {
		return -1
	}
	results := relation.Elem().FieldByName("results")
	if !results.IsValid() || results.Kind() != reflect.Map {
		return -1
	}
	return results.Len()
}

// assignable asks the checker, counting the query under its arm, and prints the first twenty after which the
// diagnostics collection grew.
func (f *cycleFinder) assignable(arm findCyclesArm, source, target *checker.Type) bool {
	trace := f.trace
	if trace == nil {
		return f.l.checker.IsTypeAssignableTo(source, target)
	}
	before := trace.diagnosticCount()
	value := f.l.checker.IsTypeAssignableTo(source, target)
	after := trace.diagnosticCount()
	trace.queries[arm]++
	trace.total++
	if after > before {
		trace.grown[arm]++
		if trace.growths < 20 {
			trace.growths++
			fmt.Fprintf(os.Stderr, "FINDCYCLES grew #%d arm=%s before=%d after=%d source=%q target=%q\n", trace.growths, findCyclesArmNames[arm], before, after, findCyclesCut(f.l.checker.TypeToString(source)), findCyclesCut(f.l.checker.TypeToString(target)))
		}
	}
	if trace.total%1000 == 0 && time.Since(trace.printed) >= 30*time.Second {
		trace.printed = time.Now()
		trace.line("tick")
	}
	return value
}

// reachLinksCalled counts a reachLinks call and the proven type it was asked for.
func (trace *findCyclesTrace) reachLinksCalled(node cycleNode) {
	if trace == nil {
		return
	}
	trace.reachLink++
	if node.proven != nil {
		trace.proven[node.proven] = true
	}
}

// reachEdgesReset counts reaches dropping its reachLinks memo, which it does whenever shapes has grown.
func (trace *findCyclesTrace) reachEdgesReset() {
	if trace == nil {
		return
	}
	trace.resets++
}

// phase prints the finder's sizes where a phase begins.
func (trace *findCyclesTrace) phase(f *cycleFinder, name string) {
	if trace == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "FINDCYCLES phase %s elapsed=%s seen=%d shapes=%d where=%d diagnostics=%d\n", name, time.Since(trace.started).Round(time.Millisecond), len(f.seen), len(f.shapes), len(f.where), trace.diagnosticCount())
}

func (trace *findCyclesTrace) finish() {
	if trace == nil {
		return
	}
	trace.line("end")
}

func (trace *findCyclesTrace) line(kind string) {
	var counts strings.Builder
	for arm, count := range trace.queries {
		fmt.Fprintf(&counts, " %s=%d", findCyclesArmNames[arm], count)
		if trace.grown[arm] != 0 {
			fmt.Fprintf(&counts, "/grew%d", trace.grown[arm])
		}
	}
	keys, longest := trace.collisions()
	fmt.Fprintf(os.Stderr, "FINDCYCLES %s elapsed=%s queries=%d reachLinks=%d distinctProven=%d reachEdgesResets=%d diagnostics=%d collisionKeys=%d collisionLongest=%d assignablePairs=%d overflowMarked=%t%s\n", kind, time.Since(trace.started).Round(time.Millisecond), trace.total, trace.reachLink, len(trace.proven), trace.resets, trace.diagnosticCount(), keys, longest, trace.assignablePairs(), load.RelationOverflowMarked(trace.checker), counts.String())
}

func findCyclesCut(text string) string {
	if len(text) > 200 {
		return text[:200]
	}
	return text
}
