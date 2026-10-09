package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/system-inc/cohere/internal/edit"
)

// earlyFormat is a bare run's format pass begun before the graph is built (#g3046x5).
//
// Traced on ahra, cold: in the 540ms before the walk, about 6.3 of the 8.6 CPU-seconds the cores could give
// went unused, since discovery, the tsconfig's enumeration and the program's file loads are bound by the
// kernel rather than the CPU. Begun here, about 2.6 of formatting's 5.8 CPU-seconds are done before the walk
// and none after it, against 1.35 after it before: about 0.06s of a cold run's wall, measured. It is not more
// because the early pass's own file opens wait on the same kernel path as the program's. Formatting needs no program, only each file's bytes and the options its
// directory resolves to, so the speculation the fix phase starts with the walk can start here instead, on
// every idle core, and narrow to its usual share when the walk begins. What it computes is judged exactly as
// before: keepable discards any attempt the walk proposed a fix for, or whose bytes the walk read differently.
//
// Only the scope is computed off the main goroutine. Declaring the walk's inputs to the run cache, or declining
// it, stays on the main goroutine where the scope is taken (formatScopeOf), as it always was.
type earlyFormat struct {
	record      *formatRecord
	formatting  *formatClock
	lineEndings *crlfFiles

	// ready closes once universe, scope and speculation are set.
	ready       chan struct{}
	universe    formatUniverse
	err         error
	scope       formatScope
	speculation *formatSpeculation

	// narrowing guards speculation's arrival against the fix phase's narrowing of it. The walk can begin before
	// this pass has a speculation, since it no longer waits for the scope (#g3046x5), so the share the fix phase
	// asks for is kept here and applied when the speculation starts. Zero until the fix phase asks.
	narrowing sync.Mutex
	narrowTo  int

	// programs is where the graph's program arrives once it is built, for the speculation to format TypeScript
	// from its trees (speculateFormat).
	programs programOffer
}

// startEarlyFormat begins the default scope's walk and its speculation, and returns at once. record is the one the
// run would have loaded, and transform is built from it the way the fix phase builds its own.
func startEarlyFormat(engine formatEngine, record *formatRecord, root string, maxPasses int) *earlyFormat {
	early := &earlyFormat{record: record, formatting: newFormatClock(), lineEndings: &crlfFiles{}, ready: make(chan struct{})}
	go func() {
		defer close(early.ready)
		// A test instrument: hold the pass this long before it enumerates, so its scope is ready only after the
		// walk is done, as a cold run's often was. See TestALateEarlyFormatScopeFindsWhatAPromptOneDoes.
		if held, err := time.ParseDuration(os.Getenv("COHERE_TEST_LATE_EARLY_FORMAT")); err == nil {
			time.Sleep(held)
		}
		early.universe, early.err = enumerateFormatUniverse(engine, root)
		if early.err != nil {
			early.started(speculateFormatOn(nil, nil, maxPasses, 0, nil))
			return
		}
		early.scope = unformattedScopeOf(engine, record, early.universe)
		transform := early.lineEndings.observing(early.transform(engine))
		early.started(speculateFormatOn(early.scope.formatCandidates(), transform, maxPasses, earlyFormatWorkers(), &early.programs))
	}()
	return early
}

// started keeps the pass's speculation, narrowed at once if the fix phase has already asked.
func (early *earlyFormat) started(speculation *formatSpeculation) {
	early.narrowing.Lock()
	defer early.narrowing.Unlock()
	early.speculation = speculation
	if early.narrowTo > 0 {
		speculation.narrow(early.narrowTo)
	}
}

// narrow gives the speculation its share of the cores once the walk begins: now if it is running, and as it
// starts if not.
func (early *earlyFormat) narrow(workers int) {
	early.narrowing.Lock()
	defer early.narrowing.Unlock()
	early.narrowTo = workers
	if early.speculation != nil {
		early.speculation.narrow(workers)
	}
}

// speculationOf waits for the pass and returns its speculation.
func (early *earlyFormat) speculationOf() *formatSpeculation {
	<-early.ready
	return early.speculation
}

// transform is the fix phase's format transform over this pass's scope, built the same way.
func (early *earlyFormat) transform(engine formatEngine) edit.Transform {
	return early.formatting.timing(scopedTransform(early.record.observe(formatTransform(engine), optionsFingerprintOf(engine)), early.scope))
}

// formatScopeOf waits for the early pass and returns the default scope, declaring the walk to the run cache, or
// declining it, as unformattedScope does.
func (early *earlyFormat) formatScopeOf() (formatScope, []string) {
	<-early.ready
	if early.err != nil {
		declineRunCache("the format walk failed")
		return formatScope{
			index:       map[string]struct{}{},
			Description: fmt.Sprintf("nothing (could not enumerate the tree: %v)", early.err),
			failure:     early.err,
		}, nil
	}
	declareFormatWalk(early.universe.root)
	return early.scope, early.universe.files
}

// earlyFormatWorkers is how many files the early pass formats at once: half the cores.
//
// Not all of them, though most sit idle before the walk: what bounds that window is the kernel, which on macOS
// runs file opens one at a time, and the early pass opens its files on the same path as the program's loads.
// Measured on ahra (#g3046x5), cold, 6 rotated rounds at load 4 to 6, against 1.787s with no early pass: 16
// workers 1.720s with the graph build about 45ms slower, 12 workers 1.730s, 8 workers 1.729s with the graph
// about 10ms slower. Half keeps the saving and leaves the build its pace.
func earlyFormatWorkers() int {
	return max(runtime.GOMAXPROCS(0)/2, 1)
}
