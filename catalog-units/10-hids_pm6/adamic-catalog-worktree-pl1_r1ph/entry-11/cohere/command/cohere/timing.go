package main

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// printTimings renders per-rule cost, most expensive first.
//
// A rule budget you cannot see is a rule budget you cannot hold. Every performance win in the
// migration this tool replaces came from a measurement nobody could have inferred: the worst
// offender cost 2,580ms of a 3,700ms run because it resolved a scope lookup on every identifier
// before testing whether the name even looked abbreviated, and a cheap structural gate took it to
// 152ms with byte-identical findings. Nobody predicted that. A timing table found it.
//
// Setup and listeners are separate columns because they are different defects. Time in setup means
// a rule is doing expensive work to decide it has nothing to do. Time in listeners means it is
// doing expensive work per node. The fixes have nothing in common.
//
// Every cost here is the walk threads' own CPU clock (see program.Timings), so a rule that waits costs
// nothing and the first line says so.
func printTimings(out io.Writer, timings *program.Timings, lintDuration time.Duration) {
	sorted := timings.Sorted()
	if len(sorted) == 0 {
		// A timing table with no rows is the instrument measuring nothing, which must not print as
		// though it measured everything and found it fast.
		fmt.Fprintf(out, "\ntiming: no rules ran, so there is nothing to measure\n")
		return
	}

	account := timings.Account
	if account.Unavailable != "" {
		// No CPU is not zero CPU. A table of zeros would read as every rule being free, so the CPU
		// columns are left out and the reason given, and the counts, which were still counted, stay.
		fmt.Fprintf(out, "\ntiming: no rule CPU this run, because %s. The counts below are real; the CPU columns are left out rather than shown as zero.\n",
			account.Unavailable)
		fmt.Fprintf(out, "  %-42s %8s %8s %8s\n", "rule", "files", "nodes", "found")
		for _, timing := range sorted {
			fmt.Fprintf(out, "  %-42s %8d %8d %8d\n", timing.Name, timing.FilesListened, timing.NodesOffered, timing.Findings)
		}
		return
	}

	// CPU is summed across workers, so it exceeds wall clock by roughly the worker count. Shares are
	// taken against the rules' total, which is the number they are actually a share of.
	ruleCPU := timings.TotalCPU()

	// The unit is in the first line, because the table this replaced printed wall time under the same
	// columns and was read as cost (#8qyzmxw): a rule descheduled mid-call was billed for the wait.
	fmt.Fprintf(out, "\ntiming: rule CPU, from each walk thread's own CPU clock, not wall time: "+
		"%d rules, %s of rule CPU across all workers, %s wall clock\n",
		len(sorted), formatMilliseconds(ruleCPU), formatMilliseconds(lintDuration))
	perCall := time.Duration(0)
	if account.Calls > 0 {
		perCall = account.InstrumentCPU / time.Duration(account.Calls)
	}
	fmt.Fprintf(out, "  (a rule that waits, sleeps or is descheduled costs nothing here. Each of the %d measured calls "+
		"costs about %dns to measure, subtracted from every row and counted once below. Compare rules to each other: "+
		"the measurement makes a --timing run slower than a real one.)\n",
		account.Calls, perCall.Nanoseconds())
	printCoverage(out, timings)
	fmt.Fprintf(out, "  %-42s %9s %9s %9s %8s %8s %8s\n",
		"rule", "cpu", "setup", "listen", "files", "nodes", "found")

	for _, timing := range sorted {
		share := ""
		if ruleCPU > 0 {
			share = fmt.Sprintf(" %5.1f%%", 100*float64(timing.TotalCPU())/float64(ruleCPU))
		}

		fmt.Fprintf(out, "  %-42s %9s %9s %9s %8d %8d %8d%s\n",
			timing.Name,
			formatMilliseconds(timing.TotalCPU()),
			formatMilliseconds(timing.SetupCPU),
			formatMilliseconds(timing.ListenerCPU),
			timing.FilesListened,
			timing.NodesOffered,
			timing.Findings,
			share,
		)
	}

	printSharedFills(out, timings)
	printTimingNotes(out, sorted)
}

// printSharedFills reports work that several rules share, which belongs to none of them.
//
// Without this the first rule to ask for a cached derivation carries its whole cost, and since
// files are walked in parallel that rule is arbitrary. Measured before this existed: three comment
// rules doing identical work reported 171ms, 132ms, and 1.0ms, and the cheap one had simply asked
// last. A reader would have concluded the 171ms rule was expensive and optimized the wrong thing.
//
// Reported by family, which is a cache key up to its first colon, because a key identifies a cache
// entry and a cache entry is not a line worth reading. The HIR cache keys per function node, by kind
// and source offset, which is correct for a cache and once produced 11,150 rows here against three
// real ones: `--timing` was 17,000 lines, two thirds of them a single derivation reported one
// function at a time, each costing 0.00ms and none of them actionable. The family is the label the
// fill's CPU carries, so the collapse happens where it is measured.
func printSharedFills(out io.Writer, timings *program.Timings) {
	shared := timings.SharedCPU()
	if len(shared) == 0 {
		return
	}

	families := make([]string, 0, len(shared))
	for family := range shared {
		families = append(families, family)
	}
	sort.Strings(families)

	for _, family := range families {
		fmt.Fprintf(out, "  shared: %s cost %s of CPU, computed once per file and used by several rules\n",
			family, formatMilliseconds(shared[family]))
	}
}

// printTimingNotes names the two shapes a per-rule table does not make obvious on its own.
//
// The coverage line already reports a rule that listened to nothing. These are the other two the
// brief asked for: a rule that listens to everything, and work duplicated across rules.
func printTimingNotes(out io.Writer, sorted []program.RuleTiming) {
	mostOffered := 0
	for _, timing := range sorted {
		if timing.NodesOffered > mostOffered {
			mostOffered = timing.NodesOffered
		}
	}
	if mostOffered == 0 {
		return
	}

	// The rule offered the most nodes is the reference point for every other row, not a warning
	// about that rule. Registering for a common kind and returning early is cheap, so this row shows
	// what a well-gated rule costs at maximum node volume. A rule with comparable nodes and much
	// higher time is doing expensive work per node, and that comparison is what makes the case.
	//
	// This started as a warning and was reworded, because the framing was backwards on the first
	// real run: consistency-no-ambiguous-identifier had the most nodes and was among the cheapest,
	// while a rule with slightly fewer nodes cost sixty times more.
	for _, timing := range sorted {
		if timing.NodesOffered < mostOffered || timing.FilesListened == 0 {
			continue
		}

		perNode := ""
		if timing.NodesOffered > 0 {
			perNode = fmt.Sprintf(", %.0fns per node", float64(timing.TotalCPU().Nanoseconds())/float64(timing.NodesOffered))
		}
		fmt.Fprintf(out,
			"  note: %s saw the most nodes of any rule (%d%s) — compare a rule's cost against this one, not against its own node count\n",
			timing.Name, timing.NodesOffered, perNode)
	}

	// Duplicated work is invisible per rule and obvious in aggregate: three rules independently
	// scanning the same comment trivia each look reasonable and together cost three scans. Rules
	// offered the same node count are the candidates, since they registered for the same kinds.
	byOffered := map[int][]string{}
	for _, timing := range sorted {
		if timing.NodesOffered > 0 {
			byOffered[timing.NodesOffered] = append(byOffered[timing.NodesOffered], timing.Name)
		}
	}
	// A pair counts, not just three or more. The defect that motivated this note was found on a
	// pair: two rules offered exactly 3,407 nodes and differing 83x in cost, where one had filled a
	// shared cache the other read for free. Requiring three would have missed it.
	//
	// The cost spread is what makes it worth printing. Rules registered for the same kind and
	// costing about the same are unremarkable; the same rules an order of magnitude apart mean one
	// of them is carrying work the others are not.
	offeredCounts := make([]int, 0, len(byOffered))
	for offered := range byOffered {
		offeredCounts = append(offeredCounts, offered)
	}
	sort.Ints(offeredCounts)

	for _, offered := range offeredCounts {
		names := byOffered[offered]
		if len(names) < 2 {
			continue
		}

		extremes := costRange(sorted, names)
		if extremes.cheapestCost <= 0 || float64(extremes.dearestCost)/float64(extremes.cheapestCost) < 10 {
			continue
		}

		fmt.Fprintf(out,
			"  note: %s and %s were each offered %d nodes but differ %.0fx in cost (%s against %s) — if they share work, this table may be billing it to whichever ran first\n",
			extremes.dearestName, extremes.cheapestName, offered,
			float64(extremes.dearestCost)/float64(extremes.cheapestCost),
			formatMilliseconds(extremes.dearestCost), formatMilliseconds(extremes.cheapestCost))
	}
}

// costExtremes is the cheapest and dearest rule in a group, each with the name it belongs to.
//
// Name and cost travel together rather than being looked up separately, because separating them is
// exactly the bug this replaced: the caller printed names[0] and names[len-1] while the costs came
// from a scan for the minimum and maximum. With two rules those coincide; with three they come
// apart, and the note then quotes one rule's cost beside another rule's name.
//
// Reproduced before the fix on three rules at 3,407 nodes each: the line named a 50ms rule and
// quoted 400ms, and the rule that actually cost 400ms did not appear at all. A reader chases the
// wrong rule, which is the failure this note exists to prevent.
type costExtremes struct {
	cheapestName string
	cheapestCost time.Duration
	dearestName  string
	dearestCost  time.Duration
}

// costRange finds the cheapest and dearest among the named rules.
//
// Reported as a spread rather than a total because the spread is the signal: equal node counts with
// equal cost is a coincidence, and equal node counts an order of magnitude apart is a rule carrying
// somebody else's work.
func costRange(sorted []program.RuleTiming, names []string) costExtremes {
	named := map[string]bool{}
	for _, name := range names {
		named[name] = true
	}

	extremes := costExtremes{}
	for _, timing := range sorted {
		if !named[timing.Name] {
			continue
		}
		total := timing.TotalCPU()
		if extremes.cheapestName == "" || total < extremes.cheapestCost {
			extremes.cheapestName, extremes.cheapestCost = timing.Name, total
		}
		if extremes.dearestName == "" || total > extremes.dearestCost {
			extremes.dearestName, extremes.dearestCost = timing.Name, total
		}
	}
	return extremes
}

// formatMilliseconds reads at a glance, which is the resolution any of these numbers means anything
// at. Sub-millisecond costs print as a fraction rather than as a bare 0, because a rule at 0.4ms
// and a rule that never ran are different facts.
func formatMilliseconds(duration time.Duration) string {
	milliseconds := float64(duration) / float64(time.Millisecond)
	if milliseconds >= 100 {
		return fmt.Sprintf("%.0fms", milliseconds)
	}
	if milliseconds >= 1 {
		return fmt.Sprintf("%.1fms", milliseconds)
	}
	return fmt.Sprintf("%.2fms", milliseconds)
}

// printCoverage says what share of the walk's CPU the rows account for, and what the rest was.
//
// Every number in the table is true and the table implies something false about what it covers. The
// rows are rule listeners and per-file setup, and a reader who sums the share column gets 100 percent
// and concludes the phase is explained. It is not: the walk that offers nodes to those listeners costs
// CPU too. On the wall-clock table this replaced, that walk could only be guessed at by subtracting
// rule time from wall clock, which absorbed every wait, and in a parallel run it vanished entirely.
//
// The threads' clocks account for it. Each worker's CPU across its walk is read, the calls into rules,
// derivations and the measurement itself are read inside it, and the walk is what of the worker's CPU
// lay outside every call: a difference between readings of one thread's clock, which holds no waiting.
// The process's CPU outside the workers is read the same way.
func printCoverage(out io.Writer, timings *program.Timings) {
	account := timings.Account
	ruleCPU := timings.TotalCPU()
	var sharedCPU time.Duration
	for _, cost := range timings.SharedCPU() {
		sharedCPU += cost
	}
	workersCPU := ruleCPU + sharedCPU + account.InstrumentCPU + account.WalkCPU
	if workersCPU <= 0 {
		// Nothing measured on the walk workers means no share to state. A share computed from a zero
		// would be worse than the silence it replaces.
		return
	}

	fmt.Fprintf(out,
		"  coverage: the walk workers spent %s of CPU: rules %.0f%%, shared derivations %.0f%%, measuring %.0f%%, and %.0f%% "+
			"the walk itself (traversal, dispatch, configuration). Another %s of process CPU ran on no walk worker "+
			"(the garbage collector, the runtime).\n",
		formatMilliseconds(workersCPU),
		100*float64(ruleCPU)/float64(workersCPU),
		100*float64(sharedCPU)/float64(workersCPU),
		100*float64(account.InstrumentCPU)/float64(workersCPU),
		100*float64(account.WalkCPU)/float64(workersCPU),
		formatMilliseconds(account.OtherCPU),
	)
}

// printGraphTiming says what the graph phase was made of, under the line that says how long it took.
//
// The phase was a single number for weeks, and every lever proposed for it was a guess about which part
// was large (#cazsft3). The wall parts come first because they add up to the phase. The summed parts
// come after and say so, because the compiler loads files on many goroutines and their summed time runs
// well past the wall: read as wall, it would look like a measurement that disagrees with itself.
//
// Parse and import resolution are not split. Both happen inside one load, behind the compiler's own
// loader, and a number for either would be a guess presented as a measurement. What the loads and the
// disk leave of the program's wall is resolution and the loader's bookkeeping together, and the last
// line says so.
func printGraphTiming(out io.Writer, timing *program.GraphTiming, buildDuration time.Duration, contentPackOpened time.Duration, checkers int) {
	accounted := contentPackOpened + timing.Config + timing.Program + timing.Verify
	builds := ""
	if timing.Builds > 1 {
		builds = fmt.Sprintf(" (%d builds: the first one's files moved under it)", timing.Builds)
	}
	fmt.Fprintf(out, "graph timing: %s wall = content pack %s + tsconfig %s + program %s + verify %s + %s elsewhere%s\n",
		formatMilliseconds(buildDuration), formatMilliseconds(contentPackOpened), formatMilliseconds(timing.Config),
		formatMilliseconds(timing.Program), formatMilliseconds(timing.Verify),
		formatMilliseconds(max(buildDuration-accounted, 0)), builds)
	fmt.Fprintf(out, "  tsconfig, its include patterns enumerated: disk %s summed over %s\n",
		formatMilliseconds(timing.ConfigDisk.Summed()), describeDiskCalls(timing.ConfigDisk))
	pack := "no content pack"
	if timing.PackServed+timing.PackRead > 0 {
		pack = fmt.Sprintf("the content pack served %d of the reads and %d were read from disk", timing.PackServed, timing.PackRead)
	}
	if timing.CheckedAnswers > 0 {
		pack += fmt.Sprintf("; %d existence checks and stats, config's and program's, were answered from the run cache's check rather than by the disk", timing.CheckedAnswers)
	}
	fmt.Fprintf(out, "  program, summed over the compiler's parallel loaders: %d files loaded (each a read and a parse) in %s; "+
		"disk %s over %s; %s\n",
		timing.SourceFileLoads, formatMilliseconds(timing.SourceFileSummed),
		formatMilliseconds(timing.ProgramDisk.Summed()), describeDiskCalls(timing.ProgramDisk), pack)
	fmt.Fprintf(out, "  (summed times run past the program's %s wall because the loaders run at once; parse and import "+
		"resolution happen inside one load and are not split, so what the loads leave of the wall is resolution "+
		"and the loader's own work together)\n", formatMilliseconds(timing.Program))
	// The count every later phase runs at, said where a reader timing the run looks: by default one per core
	// up to 12, where the measured wall stops improving (program.defaultCheckerCount, #xwv641q).
	fmt.Fprintf(out, "  checkers: %d, on %d cores\n", checkers, runtime.GOMAXPROCS(0))
}

// describeDiskCalls lists each kind of filesystem call that happened, with its count and summed time.
func describeDiskCalls(disk program.DiskTiming) string {
	parts := []string{}
	for _, kind := range []struct {
		name  string
		calls program.DiskCalls
	}{
		{"reads", disk.Reads},
		{"existence checks", disk.Existence},
		{"stats", disk.Stats},
		{"directory listings", disk.Listings},
		{"realpaths", disk.Realpaths},
	} {
		if kind.calls.Count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s %s", kind.calls.Count, kind.name, formatMilliseconds(kind.calls.Summed)))
	}
	if len(parts) == 0 {
		// Said, rather than printing an empty list that reads like a sentence cut off: a build answered
		// entirely by the cache above the disk made no call that reached it.
		return "no calls (the stat and listing cache answered all of them)"
	}
	return strings.Join(parts, ", ")
}
