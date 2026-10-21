package allowed

import (
	"os"
	"testing"
)

func TestParallelFirst(t *testing.T) {
	t.Parallel()
	t.Log("parallel")
}

func Test(t *testing.T) {
	t.Parallel()
}

func TestRenamedParameter(test *testing.T) {
	test.Parallel()
}

// Not parallel: it sets the process's working directory, which every test in the package shares.
func TestReason(t *testing.T) {
	_ = os.Chdir(".")
}

// TestReasonInADocComment says what it does first.
//
// Not parallel: t.Setenv refuses a parallel test, and the environment is the process's.
func TestReasonInADocComment(t *testing.T) {
	t.Setenv("ADAMIC_FIXTURE", "set")
}

/* Not parallel: a block comment may say it too. */
func TestReasonInABlockComment(t *testing.T) {}

func BenchmarkSerial(b *testing.B) {}

func FuzzSerial(f *testing.F) {}

func ExampleSerial() {}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// go test runs a Test name only when what follows Test isn't a lower-case letter.
func Testify(t *testing.T) {}

func helper(t *testing.T) {
	t.Helper()
}

// lookalike has a Parallel method of its own, and a method named like a test, which go test never runs.
type lookalike struct{}

func (lookalike) Parallel() {}

func (lookalike) TestSerial(t *testing.T) {
	lookalike{}.Parallel()
}

// A Parallel on another type after the real one changes nothing.
func TestLookalikeAfterParallel(t *testing.T) {
	t.Parallel()
	lookalike{}.Parallel()
	helper(t)
}
