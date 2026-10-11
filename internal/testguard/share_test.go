package testguard

import (
	"flag"
	"runtime"
	"testing"
)

// Not parallel: this test changes GOMAXPROCS and -test.parallel for the whole process.
func TestShareDividesTheUnitsCPUsAmongParallelTests(t *testing.T) {
	setting := flag.Lookup("test.parallel")
	previous, previousParallel := runtime.GOMAXPROCS(8), setting.Value.String()
	defer func() {
		runtime.GOMAXPROCS(previous)
		if err := flag.Set("test.parallel", previousParallel); err != nil {
			t.Fatal(err)
		}
	}()
	for _, check := range []struct {
		parallel string
		want     int
	}{{"8", 1}, {"2", 4}, {"1", 8}, {"16", 1}} {
		if err := flag.Set("test.parallel", check.parallel); err != nil {
			t.Fatal(err)
		}
		if got := Share(); got != check.want {
			t.Errorf("GOMAXPROCS 8, -test.parallel %s: Share %d, want %d", check.parallel, got, check.want)
		}
	}
}
