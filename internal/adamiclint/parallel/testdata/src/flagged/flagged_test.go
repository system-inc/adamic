package flagged

import "testing"

var other *testing.T

var fake lookalike

func TestNoParallel(t *testing.T) { // want `^a test runs in parallel unless it says why: call t\.Parallel\(\) first, or add // Not parallel: <the shared state it touches> directly above it$`
	t.Log("serial")
}

func TestEmptyBody(t *testing.T) {} // want `a test runs in parallel unless it says why`

func TestParallelSecond(t *testing.T) { // want `a test runs in parallel unless it says why`
	t.Helper()
	t.Parallel()
}

func TestParallelDeferred(t *testing.T) { // want `a test runs in parallel unless it says why`
	defer t.Parallel()
}

func TestParallelInASubtest(t *testing.T) { // want `a test runs in parallel unless it says why`
	t.Run("inner", func(inner *testing.T) {
		inner.Parallel()
	})
}

func TestParallelOnAnotherT(t *testing.T) { // want `a test runs in parallel unless it says why`
	other.Parallel()
}

func TestLookalikeParallel(t *testing.T) { // want `a test runs in parallel unless it says why`
	fake.Parallel()
}

// A blank parameter can't be called, so the test is serial.
func TestBlankParameter(_ *testing.T) {} // want `a test runs in parallel unless it says why`

func Test_underscored(t *testing.T) {} // want `a test runs in parallel unless it says why`

// Not parallel:
func TestEmptyReason(t *testing.T) {} // want `a test runs in parallel unless it says why`

/* Not parallel: */
func TestEmptyBlockReason(t *testing.T) {} // want `a test runs in parallel unless it says why`

// TestReasonElsewhere says Not parallel, but with no colon, so it gives no reason.
func TestReasonElsewhere(t *testing.T) {} // want `a test runs in parallel unless it says why`

// Not parallel: this reason is a paragraph away from the test below, so it is not that test's.

func TestReasonSeparated(t *testing.T) { // want `a test runs in parallel unless it says why`
	t.Log("serial")
}
