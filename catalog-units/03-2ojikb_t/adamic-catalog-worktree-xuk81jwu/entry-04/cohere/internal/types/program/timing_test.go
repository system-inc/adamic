package program

import (
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// TestSortedPutsTheExpensiveRuleFirst is what makes an outlier obvious without arithmetic.
func TestSortedPutsTheExpensiveRuleFirst(t *testing.T) {
	t.Parallel()
	timings := NewTimings([]string{"cheap", "expensive", "middling"})
	timings.forRule("cheap").ListenerCPU = time.Millisecond
	timings.forRule("expensive").ListenerCPU = 500 * time.Millisecond
	timings.forRule("middling").ListenerCPU = 50 * time.Millisecond

	sorted := timings.Sorted()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(sorted))
	}
	if sorted[0].Name != "expensive" {
		t.Fatalf("the most expensive rule is not first: %v", sorted[0].Name)
	}
	if sorted[2].Name != "cheap" {
		t.Fatalf("the cheapest rule is not last: %v", sorted[2].Name)
	}
}

// TestSetupAndListenerTimeStaySeparate pins a distinction that decides which fix a slow rule needs.
//
// Time in setup means a rule is doing expensive work to decide it has nothing to do, which is the
// shape that made one rule 2,580ms of a 3,700ms run under the previous gate. Time in listeners
// means expensive work per node. The fixes have nothing in common, so a single total would hide
// the more actionable of the two.
func TestSetupAndListenerTimeStaySeparate(t *testing.T) {
	t.Parallel()
	timings := NewTimings([]string{"decides-slowly"})
	timing := timings.forRule("decides-slowly")
	timing.SetupCPU = 300 * time.Millisecond
	timing.ListenerCPU = 2 * time.Millisecond

	sorted := timings.Sorted()
	if sorted[0].SetupCPU != 300*time.Millisecond {
		t.Fatalf("setup time was lost: %v", sorted[0].SetupCPU)
	}
	if sorted[0].ListenerCPU != 2*time.Millisecond {
		t.Fatalf("listener time was lost: %v", sorted[0].ListenerCPU)
	}
	if sorted[0].TotalCPU() != 302*time.Millisecond {
		t.Fatalf("total does not sum its parts: %v", sorted[0].TotalCPU())
	}
}

// TestMergeAccumulatesAcrossWorkers covers the path every real run takes.
//
// Workers count locally and merge once under the mutex, because counting through a shared lock would
// put contention into the walk being measured. A merge that dropped or double-counted would make every
// number quietly wrong in a way no single-worker test could see. CPU is not merged: it comes from the
// one profile of the whole walk.
func TestMergeAccumulatesAcrossWorkers(t *testing.T) {
	t.Parallel()
	shared := NewTimings([]string{"a-rule"})

	for worker := 0; worker < 4; worker++ {
		local := NewTimings(nil)
		timing := local.forRule("a-rule")
		timing.NodesOffered = 100
		timing.FilesListened = 5
		timing.FilesDeclined = 2
		timing.Findings = 3
		shared.merge(local)
	}

	merged := shared.Sorted()[0]
	if merged.NodesOffered != 400 || merged.FilesListened != 20 || merged.FilesDeclined != 8 || merged.Findings != 12 {
		t.Fatalf("counts did not accumulate: %+v", merged)
	}
}

// TestMergePicksUpARuleTheRunAddedLate keeps a worker's rule from being dropped because the shared
// collector was built without it.
func TestMergePicksUpARuleTheRunAddedLate(t *testing.T) {
	t.Parallel()
	shared := NewTimings(nil)
	local := NewTimings(nil)
	local.forRule("late-rule").NodesOffered = 1
	shared.merge(local)

	if len(shared.Sorted()) != 1 {
		t.Fatalf("a rule known only to a worker was dropped: %+v", shared.Sorted())
	}
}

// TestNilCollectorIsInert is the guard that keeps an ordinary run free.
//
// Every timing site is nil-checked so a run without --timing reads no clocks and allocates nothing.
// If that ever stops being true, the instrument starts costing what it was designed not to.
func TestNilCollectorIsInert(t *testing.T) {
	t.Parallel()
	var timings *Timings

	timings.merge(NewTimings([]string{"anything"}))

	if timings.Sorted() != nil {
		t.Fatal("a nil collector produced rows")
	}
	if timings.TotalCPU() != 0 {
		t.Fatal("a nil collector reported time")
	}
	if timings.forRule("anything") != nil {
		t.Fatal("a nil collector handed out an accumulator")
	}
}

// spinFor burns cpu of this thread's CPU, by the thread's own clock, so it is CPU however the scheduler
// treats it.
func spinFor(t *testing.T, cpu time.Duration) {
	t.Helper()
	start, known := threadCPU()
	if !known {
		t.Skip("this platform has no thread CPU clock, so --timing reports no CPU here and the meter has nothing to test")
	}
	for {
		if now, _ := threadCPU(); now-start >= cpu {
			return
		}
	}
}

// within reports whether got is want give or take a tenth, plus a little for the clock's step.
func within(got time.Duration, want time.Duration) bool {
	slack := want/10 + 200*time.Microsecond
	return got >= want-slack && got <= want+slack
}

// startTestMeter starts a meter on the test's goroutine and finishes it when the test ends.
func startTestMeter(t *testing.T) (*workerMeter, *Timings) {
	t.Helper()
	timings := NewTimings(nil)
	meter := startWorkerMeter(timings)
	if meter == nil {
		// Only a platform without a thread CPU clock starts no meter, and there --timing says so rather than
		// measuring; see cpu_clock_other.go.
		t.Skipf("no meter on this platform: %s", timings.Account.Unavailable)
	}
	t.Cleanup(meter.finish)
	return meter, timings
}

// TestAMeteredCallIsBilledItsCPU covers the wrapper the whole table rests on. Five calls each spinning 2ms
// of CPU are billed 10ms, and each call is counted as a node offered.
func TestAMeteredCallIsBilledItsCPU(t *testing.T) {
	t.Parallel()
	meter, _ := startTestMeter(t)
	timing := &RuleTiming{Name: "spins"}
	spins := func(node *ast.Node) { spinFor(t, 2*time.Millisecond) }

	for range 5 {
		meter.call(timing, spins, nil)
	}

	if timing.NodesOffered != 5 {
		t.Fatalf("calls were not counted: %d", timing.NodesOffered)
	}
	if !within(timing.ListenerCPU, 10*time.Millisecond) {
		t.Fatalf("five calls spinning 2ms were billed %v", timing.ListenerCPU)
	}
}

// TestAWaitingCallCostsNothing is the property the wall clock lacked (#8qyzmxw). Five calls each sleeping
// 5ms spend 25ms of wall time and almost no CPU, and they are billed the CPU.
func TestAWaitingCallCostsNothing(t *testing.T) {
	t.Parallel()
	meter, _ := startTestMeter(t)
	timing := &RuleTiming{Name: "sleeps"}
	sleeps := func(node *ast.Node) { time.Sleep(5 * time.Millisecond) }

	wallStart := time.Now()
	for range 5 {
		meter.call(timing, sleeps, nil)
	}
	wall := time.Since(wallStart)

	if wall < 25*time.Millisecond {
		t.Fatalf("the calls slept %v, under the 25ms planted, so they cannot show a wall clock's error", wall)
	}
	// Relative to the sleep, so it holds at any load: a busy machine stretches the sleep and the few
	// instructions around it alike. A wall clock bills all of it, about 100%.
	if timing.ListenerCPU > wall/20 {
		t.Fatalf("five calls that only slept were billed %v of CPU, having slept %v", timing.ListenerCPU, wall)
	}
}

// TestSetupIsBilledApartFromListeners keeps the distinction that decides which fix a slow rule needs.
func TestSetupIsBilledApartFromListeners(t *testing.T) {
	t.Parallel()
	meter, _ := startTestMeter(t)
	timing := &RuleTiming{Name: "decides"}
	meter.setup(timing, func() { spinFor(t, 3*time.Millisecond) })

	if !within(timing.SetupCPU, 3*time.Millisecond) || timing.ListenerCPU != 0 {
		t.Fatalf("a 3ms setup was billed setup %v and listen %v", timing.SetupCPU, timing.ListenerCPU)
	}
}

// TestSharedFillIsBilledToTheDerivationNotTheRuleThatAsked is the fix for an instrument that lied.
//
// A cached derivation is computed by whichever rule asks first, and files are walked in parallel,
// so that rule is arbitrary. Measured on three comment rules sharing one scan before this existed:
// 171ms, 132ms, and 1.0ms for identical work. The cheap one had simply asked last, and a reader
// would have concluded the 171ms rule was expensive and optimized the wrong thing.
//
// So a fill is a frame of its own, billed to its derivation's family, and the frame around it is billed
// less the fill: the rule that asked, or an outer fill when one derivation fills another. The family is
// the key up to its first colon, because the HIR cache keys per function node and once reported 11,150
// rows of 0.00ms.
func TestSharedFillIsBilledToTheDerivationNotTheRuleThatAsked(t *testing.T) {
	t.Parallel()
	meter, timings := startTestMeter(t)
	cache := rule.NewFileCache()
	meter.watchFills(cache)
	timing := &RuleTiming{Name: "asked-first"}

	meter.call(timing, func(node *ast.Node) {
		spinFor(t, 2*time.Millisecond)
		rule.Cached(cache, "comments.All", func() int {
			spinFor(t, 3*time.Millisecond)
			rule.Cached(cache, "hir.Function:175:1001", func() int {
				spinFor(t, 4*time.Millisecond)
				return 1
			})
			return 1
		})
	}, nil)

	if !within(timing.ListenerCPU, 2*time.Millisecond) {
		t.Fatalf("the rule that asked was billed %v for its own 2ms", timing.ListenerCPU)
	}
	if got := timings.SharedCPU()["comments.All"]; !within(got, 3*time.Millisecond) {
		t.Fatalf("the outer derivation was billed %v for its own 3ms", got)
	}
	if got := timings.SharedCPU()["hir.Function"]; !within(got, 4*time.Millisecond) {
		t.Fatalf("the per-node derivation was billed %v to its family for its 4ms", got)
	}
	if cache.Fills()["comments.All"] != 1 || cache.Fills()["hir.Function:175:1001"] != 1 {
		t.Fatalf("each derivation should have been computed once: %v", cache.Fills())
	}
}

// TestARuleThatPanicsStillBalancesItsFrames covers the crash path. A rule's panic is recovered by its
// containment and the walk goes on, so a frame that closed only on a normal return would leave every
// later call nested inside the crashed one, billed against it.
func TestARuleThatPanicsStillBalancesItsFrames(t *testing.T) {
	t.Parallel()
	meter, _ := startTestMeter(t)
	timing := &RuleTiming{Name: "crashes"}
	recovering := func(run func()) {
		defer func() { _ = recover() }()
		run()
	}

	recovering(func() { meter.call(timing, func(node *ast.Node) { panic("planted") }, nil) })
	recovering(func() { meter.setup(timing, func() { panic("planted") }) })

	if meter.depth != 0 {
		t.Fatalf("frames were left open after a recovered panic: depth %d", meter.depth)
	}
}

// TestTheWorkersCPUIsAccountedFor holds that the rows and the account add up to what the worker's thread
// spent, so nothing the table prints is a residual it cannot name.
func TestTheWorkersCPUIsAccountedFor(t *testing.T) {
	t.Parallel()
	timings := NewTimings(nil)
	meter := startWorkerMeter(timings)
	if meter == nil {
		t.Skipf("no meter on this platform: %s", timings.Account.Unavailable)
	}
	timing := timings.forRule("spins")
	spinFor(t, 2*time.Millisecond)
	meter.call(timing, func(node *ast.Node) { spinFor(t, 3*time.Millisecond) }, nil)
	meter.finish()

	account := timings.Account
	if account.Calls != 1 {
		t.Fatalf("one call was counted as %d", account.Calls)
	}
	if !within(account.WalkCPU, 2*time.Millisecond) {
		t.Fatalf("2ms spent outside any call was billed to the walk as %v", account.WalkCPU)
	}
	parts := timing.ListenerCPU + account.InstrumentCPU + account.WalkCPU
	if difference := account.WorkersCPU - parts; difference < -time.Microsecond || difference > time.Microsecond {
		t.Fatalf("the worker spent %v and its parts sum to %v", account.WorkersCPU, parts)
	}
}

// TestTheMeterIsPassThroughWhenNotTiming proves the ordinary path costs nothing: no meter, the listener
// itself rather than a wrapper, and a cache whose fills run as they always did.
func TestTheMeterIsPassThroughWhenNotTiming(t *testing.T) {
	t.Parallel()
	meter := startWorkerMeter(nil)
	if meter != nil {
		t.Fatal("a walk without --timing started a meter")
	}
	cache := rule.NewFileCache()
	meter.watchFills(cache)

	called := 0
	meter.call(nil, func(node *ast.Node) { called++ }, nil)
	meter.setup(nil, func() { called++ })
	meter.finish()
	if called != 2 {
		t.Fatalf("the rule was not called through: %d of 2", called)
	}
	if got := rule.Cached(cache, "a.derivation", func() int { return 7 }); got != 7 {
		t.Fatalf("an unmetered cache did not fill: %d", got)
	}
}
