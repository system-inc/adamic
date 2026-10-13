package testguard

import (
	"flag"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
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

// Not parallel: this test changes GOMAXPROCS and -test.parallel for the whole process.
func TestEachRunsEveryIndexAtMostShareAtOnce(t *testing.T) {
	setting := flag.Lookup("test.parallel")
	previous, previousParallel := runtime.GOMAXPROCS(8), setting.Value.String()
	defer func() {
		runtime.GOMAXPROCS(previous)
		if err := flag.Set("test.parallel", previousParallel); err != nil {
			t.Fatal(err)
		}
	}()
	for _, parallel := range []string{"8", "2"} {
		if err := flag.Set("test.parallel", parallel); err != nil {
			t.Fatal(err)
		}
		var running, most atomic.Int32
		var ran [12]atomic.Bool
		Each(len(ran), func(index int) {
			now := running.Add(1)
			for current := most.Load(); now > current && !most.CompareAndSwap(current, now); current = most.Load() {
			}
			time.Sleep(5 * time.Millisecond)
			ran[index].Store(true)
			running.Add(-1)
		})
		for index := range ran {
			if !ran[index].Load() {
				t.Errorf("-test.parallel %s: index %d never ran", parallel, index)
			}
		}
		if got := int(most.Load()); got > Share() {
			t.Errorf("-test.parallel %s: %d ran at once, Share is %d", parallel, got, Share())
		}
	}
}
