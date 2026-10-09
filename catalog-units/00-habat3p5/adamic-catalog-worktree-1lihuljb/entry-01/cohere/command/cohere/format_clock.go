package main

import (
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
)

// formatClock measures how much of the fix phase's wall clock had a format in flight, which the footer's
// --phases shows as 💅 beside 🪄.
//
// The two are not separate phases. A file is fixed to its fixpoint and then formatted, by several workers
// at once, so fixing and formatting interleave across the tree. Summing each call's time would count four
// workers formatting for a second as four seconds, past the phase's own wall clock. What is measured
// instead is the time during which at least one format was running: every moment of the phase is either
// that or not, so 💅 and 🪄 partition the phase and add up to it. A moment where one worker formats while
// another fixes is billed to 💅, which is the side a reader asking "is the formatter slow?" wants it on.
type formatClock struct {
	mutex    sync.Mutex
	inFlight int
	since    time.Time
	total    time.Duration
	// now is the clock, a variable so a test can step it.
	now func() time.Time
}

func newFormatClock() *formatClock {
	return &formatClock{now: time.Now}
}

// timing wraps a transform so each call is on the clock while it runs. A nil transform stays nil, since
// the fix phase reads nil as no transform at all.
func (clock *formatClock) timing(inner edit.Transform) edit.Transform {
	if inner == nil {
		return nil
	}
	return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		clock.start()
		defer clock.stop()
		return inner(fileName, text, parsed)
	}
}

func (clock *formatClock) start() {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	if clock.inFlight == 0 {
		clock.since = clock.now()
	}
	clock.inFlight++
}

func (clock *formatClock) stop() {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	clock.inFlight--
	if clock.inFlight == 0 {
		clock.total += clock.now().Sub(clock.since)
	}
}

// Total is the wall clock with a format in flight, so far.
func (clock *formatClock) Total() time.Duration {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	return clock.total
}
