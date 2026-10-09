package main

import (
	"errors"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// TestTheFormatClockCountsWallTimeNotCalls holds 💅 to the phase's wall clock: two formats overlapping
// for part of their run count once for the overlap, and a gap between formats counts for nothing. Summing
// the calls would say 7s here, past the 6s the phase could have taken.
func TestTheFormatClockCountsWallTimeNotCalls(t *testing.T) {
	t.Parallel()
	var moment time.Time
	clock := newFormatClock()
	clock.now = func() time.Time { return moment }
	at := func(seconds int) { moment = time.Unix(int64(seconds), 0) }

	// One from 0 to 3, another from 1 to 4, then nothing until a third from 5 to 6.
	at(0)
	clock.start()
	at(1)
	clock.start()
	at(3)
	clock.stop()
	at(4)
	clock.stop()
	at(5)
	clock.start()
	at(6)
	clock.stop()

	if got, want := clock.Total(), 5*time.Second; got != want {
		t.Errorf("the clock says %s with a format in flight, want %s", got, want)
	}
}

// TestTheFormatClockKeepsATransformsAnswer holds the wrapper to timing and nothing else: a nil transform
// stays nil, which the fix phase reads as no transform, and a transform's text and error pass through.
func TestTheFormatClockKeepsATransformsAnswer(t *testing.T) {
	t.Parallel()
	clock := newFormatClock()
	if clock.timing(nil) != nil {
		t.Error("a nil transform came back as a transform")
	}
	refused := errors.New("refused")
	timed := clock.timing(func(fileName string, text string, _ *ast.SourceFile) (string, error) { return text + "!", refused })
	if text, err := timed("a.ts", "x", nil); text != "x!" || !errors.Is(err, refused) {
		t.Errorf("the timed transform answered %q, %v", text, err)
	}
}

// TestFixingAndFormattingAddUpToTheFixPhase is the footer's half: 🪄 is the fix phase less formatting,
// so the two partition it, and a formatting time past the phase's, which no run produces, never prints a
// negative 🪄.
func TestFixingAndFormattingAddUpToTheFixPhase(t *testing.T) {
	t.Parallel()
	records := []phaseRecord{{Name: phaseFix, Outcome: outcomeRan, Elapsed: 700 * time.Millisecond}}
	got := phaseTimes(0, records, 300*time.Millisecond)
	if len(got) != 2 || got[0] != "🪄 0.4s" || got[1] != "💅 0.3s" {
		t.Errorf("the fix phase's 0.7s with 0.3s formatting printed %q", got)
	}
	if got := phaseTimes(0, records, time.Second); got[0] != "🪄 0.000s" {
		t.Errorf("formatting past the phase printed %q", got)
	}
}
