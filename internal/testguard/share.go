package testguard

import (
	"flag"
	"runtime"
	"strconv"
	"sync"
)

// Share is how many child processes one test may keep running at once: the unit's CPUs divided among the tests go
// test runs in parallel, never less than one. The unit's CPUs are GOMAXPROCS, which Go sets from the cgroup's quota and
// the affinity mask, never runtime.NumCPU, the host's cores: a runner gives a unit 8 CPUs of a 64-core box beside other
// units, and json's shards, each sizing to 64 cores, ran 8 ASan children apiece across 8 parallel shards there until
// one passed its 90 s deadline (landable-16c, Judge). -test.parallel defaults to GOMAXPROCS, so under go test's defaults
// a parallel test runs one child at a time and the unit as a whole runs one per CPU.
func Share() int {
	processors := runtime.GOMAXPROCS(0)
	parallel := processors
	if setting := flag.Lookup("test.parallel"); setting != nil {
		if value, err := strconv.Atoi(setting.Value.String()); err == nil && value > 0 {
			parallel = value
		}
	}
	return max(1, processors/parallel)
}

// Each runs work(0) through work(count-1), at most Share() of them at once, and returns when every one has. A test that
// starts its sides together (a Go oracle, Node, the release build, the sanitizers) runs them one at a time under go
// test's defaults, so the unit runs about one process per CPU, not a test's whole wave per CPU (landable-20b: json's
// _445, 29.5 s of work alone on one core, passed 90 s at 8 CPUs beside the top shards' waves of four).
func Each(count int, work func(index int)) {
	slots := make(chan struct{}, Share())
	var group sync.WaitGroup
	for index := range count {
		slots <- struct{}{}
		group.Go(func() {
			defer func() { <-slots }()
			work(index)
		})
	}
	group.Wait()
}
