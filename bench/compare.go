// Compare times one suite of bench/suites.json at two commits of Adamic, and says whether the
// candidate is faster, slower, or no different at 95% confidence, with the variation that verdict
// rests on.
//
//	go run ./bench -compare <baseline sha> <candidate sha> -suite everyday [-samples 15] [-only trees] [-plant candidate [-plant-ms 25]] [-seed n]
//
// cmd/adamic is built at each commit in a detached worktree, and every program in the suite is built
// with each of the two compilers from the same source, the one in this checkout, so the compiler is
// the only thing that differs. Each sample runs the floor and every program eight times (sampleRuns):
// Node, the baseline, Node, the candidate, Node, the candidate, Node, the baseline, with the sides
// swapped on odd samples. Every native run follows a Node run, so a cost a run inherits from the one
// before it lands on both sides alike within each sample, and the sides' positions mirror, so a drift
// through the sample does too. A runner's runs in a sample are combined as their mean.
//
// Beside the runs' order, nothing about a side stays fixed for a whole run. Before each sample every
// native binary, the floor's too, gets fresh pages in the page cache: a copy beside it is renamed over
// it, so its path is the same and its physical placement, and with it how it aliases in the caches,
// is new.
// Every run gets ADAMIC_BENCH_PAD, a value of a random length from 0 to 4095 drawn for that run, so the
// stack starts somewhere new each time. The lengths come from one seeded source; the header prints the
// seed, and -seed draws the same lengths again.
//
// A row whose two binaries are byte for byte the same is a control: nothing the compiler did can move
// it, so it never gets a verdict, and the line above the table says how many controls would have
// claimed a change and by how much, the size of change this run can't tell from the machine.
//
// The statistic is paired: per sample, the candidate's time over the baseline's on the program's
// judge, user time unless the suite says wall (a parallel program spends more user time with more
// threads by design). The median of those ratios is reported with the sign test's 95% interval, the
// k-th smallest and k-th largest ratio, which assumes nothing about how the times are distributed. A
// verdict needs the whole interval on one side of 1.00; a program whose answer differs from Node's at
// either commit gets none, and so does a run refused the way run.go refuses one (at or below the
// clock's resolution, or within its runtime's floor).
//
// Every verdict is printed with the change the run could see: a "no change" is "no change larger
// than" its interval's farther bound, so it never reads as proof of none. Native against Node is a
// paired ratio too, with its own interval, and a program is listed as a loss or a win only when that
// interval excludes 1.00; otherwise it's within the run's noise and listed as such.
//
// -plant candidate proves the comparison can fail: it builds each candidate program from a copy of
// its source with a fixed busy loop appended, -plant-ms long (25 by default), so an A/A comparison
// with the plant must read slower once the plant is larger than the program's sensitivity. The plant
// lives here, in the harness, and never in the compiler or the runtime.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// A compareRequest is what -compare was asked.
type compareRequest struct {
	baseline  string
	candidate string
	suite     string
	samples   int
	plant     string
	plantMs   int
	only      string
	timeout   time.Duration
	seed      uint64 // ADAMIC_BENCH_PAD's lengths are drawn from it
}

// suiteFile is bench/suites.json.
type suiteFile struct {
	Suites map[string]suiteDefinition `json:"suites"`
}

// A suiteDefinition is one named set and the rule that decides what belongs in it.
type suiteDefinition struct {
	Rule     string          `json:"rule"`
	Includes []string        `json:"includes"`
	Programs []suiteMember   `json:"programs"`
	Retired  []retiredMember `json:"retired"`
	Revision string          `json:"revision"`
	Note     string          `json:"note"`
}

// A suiteMember is a program in bench/ (its name without the extension, which may name a
// subdirectory), or, of kind harness, a measurement that runs outside run.go.
type suiteMember struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Judge    string   `json:"judge"`
	Threads  []int    `json:"threads"`
	Findings []string `json:"findings"`
	Path     string   `json:"path"`
	// Input, when set, is a file the program reads, named by its first argument: made once, before any
	// timing, by running the generator (a bench/ program, built natively by the baseline's compiler)
	// with the size as its argument, so making it is never inside a timed run (Language on #re2xhwa).
	Input  *programInput `json:"input"`
	Branch string        `json:"branch"`
	Note   string        `json:"note"`
}

// A programInput is how a program's input file is made.
type programInput struct {
	Generator string `json:"generator"`
	Size      int    `json:"size"`
}

// A retiredMember left a suite by a ruling, which names why.
type retiredMember struct {
	Name   string `json:"name"`
	Ruling string `json:"ruling"`
	Why    string `json:"why"`
}

// A compareEntry is one row: a program, at one thread count when the suite names several, and the
// two native binaries built from it.
type compareEntry struct {
	program     program
	judge       string
	environment []string
	input       *programInput
	arguments   []string // the input file's path, once made
	baseline    string
	candidate   string
	failure     string // why it can't be compared, from a build
	identical   bool   // the two binaries are byte for byte the same, so the row is a control
}

// The three runners of every sample.
const (
	baselineRunner = iota
	candidateRunner
	nodeRunner
)

var runnerNames = [...]string{"baseline", "candidate", "node"}

// sampleRuns is the runs one sample makes, of the floor and of every program alike: Node, the
// baseline, Node, the candidate, Node, the candidate, Node, the baseline, and on odd samples the same
// with the sides swapped, so which side runs first evens out too.
//
// Every native run follows a Node run, so a cost a run inherits from the run before it lands on both
// sides alike inside every sample, and the sign test judges a sample at a time, so it has to cancel
// there, not only across samples. A rotation of three ran the candidate right after the baseline every
// time (M3, J-d1emr4h, where three identical binaries read a change), and balancing who follows whom
// across samples still split the ratios into samples that carried the cost and samples that didn't.
// No shorter order cancels both kinds of cost: one tied to the baseline as the run before (in N, B, C,
// N, C, B the candidate follows the baseline once and the baseline never follows itself, so only the
// candidate pays), and one tied to the other side as the run before. Only Node before every native
// run cancels both.
//
// Counting from 0, the baseline runs at positions 1 and 7 and the candidate at 3 and 5, both averaging
// position 4, so a drift that grows steadily through a sample adds the same to both sides' mean. The
// cost is eight runs a program a sample: each side's two and Node's four, each runner's combined as
// their mean (combined).
func sampleRuns(sample int) [8]int {
	if sample%2 == 0 {
		return [8]int{nodeRunner, baselineRunner, nodeRunner, candidateRunner, nodeRunner, candidateRunner, nodeRunner, baselineRunner}
	}
	return [8]int{nodeRunner, candidateRunner, nodeRunner, baselineRunner, nodeRunner, baselineRunner, nodeRunner, candidateRunner}
}

// A programComparison is one row's result: the median candidate-over-baseline ratio and its 95%
// interval with a verdict, or why there isn't one.
type programComparison struct {
	median          float64
	low             float64
	high            float64
	verdict         string
	failure         string
	candidateMedian time.Duration
	nodeMedian      time.Duration
	baselineMedian  time.Duration
	// nodeRatio and its interval are the candidate over Node, sample by sample, on the same judge.
	nodeRatio, nodeLow, nodeHigh float64
	// claimed is, on a control, the verdict its interval would have given, faster or slower, or empty
	// when it would have read no change.
	claimed string
}

const noChange = "no change at this confidence"

// control is a control row's verdict cell: its two binaries are the same, so it never claims a change.
const control = "control"

// plantLoop is the plant: a fixed number of milliseconds of busy CPU, the same whatever the program's
// size, appended to a copy of a candidate program's source, so it runs inside the timed process at the
// end of the program's synchronous code (before any timer the program set fires). It prints nothing,
// so the planted program still gives Node's answer, and only its time can tell it apart.
func plantLoop(milliseconds int) string {
	return fmt.Sprintf(`
// Planted by go run ./bench -compare -plant candidate (bench/compare.go): a fixed %d ms of busy CPU.
const adamicBenchPlantStart = Date.now();
while (Date.now() - adamicBenchPlantStart < %d) {
	// Spins, printing nothing.
}
`, milliseconds, milliseconds)
}

func compareCommits(request compareRequest) {
	if _, err := signTestRank(request.samples); err != nil {
		fail(err)
	}
	if request.plant != "" && request.plant != "candidate" {
		fail(fmt.Errorf("-plant takes only candidate, not %q", request.plant))
	}
	if request.plant != "" && (request.plantMs < 1 || request.plantMs > 60000) {
		fail(fmt.Errorf("-plant-ms must be from 1 to 60000, not %d", request.plantMs))
	}
	directory, err := benchDirectory()
	if err != nil {
		fail(err)
	}
	root := filepath.Dir(directory)
	members, harnesses, err := loadSuite(filepath.Join(directory, "suites.json"), request.suite, request.only)
	if err != nil {
		fail(err)
	}
	baselineCommit, err := resolveCommit(request.baseline)
	if err != nil {
		fail(err)
	}
	candidateCommit, err := resolveCommit(request.candidate)
	if err != nil {
		fail(err)
	}
	entries, err := compareEntries(directory, members)
	if err != nil {
		fail(err)
	}
	work, err := os.MkdirTemp("", "adamic-compare-")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(work)
	if reaper, err = buildReaper(directory, work); err != nil {
		fail(err)
	}

	describeMachine([]runtimeKind{{name: runnerNames[baselineRunner]}, {name: runnerNames[candidateRunner]}, {name: runnerNames[nodeRunner]}})
	fmt.Printf("baseline: %s\ncandidate: %s\n", baselineCommit, candidateCommit)
	plantDescription := "none"
	if request.plant == "candidate" {
		plantDescription = fmt.Sprintf("candidate, a fixed %d ms busy loop appended to every candidate program", request.plantMs)
	}
	fmt.Printf("suite: %s, %d rows; samples: %d; plant: %s\n", request.suite, len(entries), request.samples, plantDescription)
	fmt.Printf("seed: %d, which draws every run's ADAMIC_BENCH_PAD length; -seed %d draws the same again\n", request.seed, request.seed)
	fmt.Printf("load before: %s\n", loadAverage())
	fmt.Printf("processes over 5%% CPU at start: %s\n\n", busyProcesses())

	// Every binary the two sides run sits at a path of the same length, side 1 for the baseline and side 2
	// for the candidate, with the same file name: a path's length moves the argument and environment
	// strings, and with them the stack's alignment, enough to slow a floating-point loop by a few percent
	// on its own (M1's third run, J-eewt3vf: an A/A read nbody 3% slower with nothing changed, when
	// the two sides' binaries were nbody.baseline and nbody.candidate).
	sides := [2]string{filepath.Join(work, "1"), filepath.Join(work, "2")}
	for _, side := range sides {
		if err := os.MkdirAll(side, 0755); err != nil {
			fail(err)
		}
	}
	baselineAdamic := filepath.Join(sides[0], "adamic")
	if err := buildAdamicAt(root, baselineCommit, baselineAdamic); err != nil {
		fail(err)
	}
	candidateAdamic := filepath.Join(sides[1], "adamic")
	if err := buildAdamicAt(root, candidateCommit, candidateAdamic); err != nil {
		fail(err)
	}

	floorSource := filepath.Join(work, "floor.ts")
	if err := os.WriteFile(floorSource, []byte("console.log(\"0\");\n"), 0644); err != nil {
		fail(err)
	}
	floor := program{name: "floor", source: floorSource}
	floorBinaries := [2]string{filepath.Join(sides[0], "floor"), filepath.Join(sides[1], "floor")}
	for index, adamic := range []string{baselineAdamic, candidateAdamic} {
		if output, err := step("building the floor with the "+runnerNames[index]+"'s adamic", exec.Command(adamic, "build", floor.source, "-o", floorBinaries[index])); err != nil {
			fail(fmt.Errorf("%s's adamic build %s: %v\n%s", runnerNames[index], floor.source, err, output))
		}
	}

	plantedDirectory := filepath.Join(work, "planted")
	if err := os.MkdirAll(plantedDirectory, 0755); err != nil {
		fail(err)
	}
	// A program with several thread counts is one source, built once by each compiler.
	built := map[string]compareEntry{}
	for index := range entries {
		entry := &entries[index]
		if previous, isBuilt := built[entry.program.source]; isBuilt {
			entry.baseline, entry.candidate, entry.failure = previous.baseline, previous.candidate, previous.failure
			continue
		}
		base := strings.ReplaceAll(entry.program.name, "/", "_")
		entry.baseline = filepath.Join(sides[0], base)
		entry.candidate = filepath.Join(sides[1], base)
		candidateSource := entry.program.source
		if request.plant == "candidate" {
			candidateSource = filepath.Join(plantedDirectory, base+filepath.Ext(entry.program.source))
			err = plantedCopy(entry.program.source, candidateSource, request.plantMs)
			if err != nil {
				fail(err)
			}
		}
		if output, err := step("building "+entry.program.name+" with the baseline's adamic", exec.Command(baselineAdamic, "build", entry.program.source, "-o", entry.baseline)); err != nil {
			entry.failure = fmt.Sprintf("the baseline's build failed: %v %s", err, firstLine(string(output)))
		} else if output, err := step("building "+entry.program.name+" with the candidate's adamic", exec.Command(candidateAdamic, "build", candidateSource, "-o", entry.candidate)); err != nil {
			entry.failure = fmt.Sprintf("the candidate's build failed: %v %s", err, firstLine(string(output)))
		}
		built[entry.program.source] = *entry
	}
	// Each input is made once, by the generator built with the baseline's compiler, before any sample.
	inputs := map[string]string{}
	for index := range entries {
		entry := &entries[index]
		if entry.input == nil || entry.failure != "" {
			continue
		}
		key := fmt.Sprintf("%s-%d", strings.ReplaceAll(entry.input.Generator, "/", "_"), entry.input.Size)
		if path, made := inputs[key]; made {
			entry.arguments = []string{path}
			continue
		}
		path, err := makeInput(directory, work, baselineAdamic, *entry.input, key)
		if err != nil {
			entry.failure = err.Error()
			continue
		}
		inputs[key] = path
		entry.arguments = []string{path}
		fmt.Printf("input %s: %s at size %d\n", path, entry.input.Generator, entry.input.Size)
	}
	// Whether each row's two binaries are byte for byte the same: in an A/A run they should be, so a
	// verdict there is the machine or the harness, never the compiler.
	identical, differing := []string{}, []string{}
	for index := range entries {
		entry := &entries[index]
		if entry.failure != "" {
			continue
		}
		baselineBytes, baselineError := os.ReadFile(entry.baseline)
		candidateBytes, candidateError := os.ReadFile(entry.candidate)
		if baselineError == nil && candidateError == nil && string(baselineBytes) == string(candidateBytes) {
			entry.identical = true
			identical = append(identical, entry.program.name)
		} else {
			differing = append(differing, entry.program.name)
		}
	}
	fmt.Printf("binaries byte for byte the same on both sides: %d (%s); differing: %d (%s)\n\n", len(identical), strings.Join(identical, ", "), len(differing), strings.Join(differing, ", "))

	// samples[entry] holds each sample's three measurements, by runner.
	samples := make([][][3]measurement, len(entries))
	// floors[runner] holds every run of the floor program, each on its own, so its slowest is one run's.
	floors := [3][]measurement{}
	floorCommands := [3][]string{{floorBinaries[0]}, {floorBinaries[1]}, nodeCommand(floor, root)}
	stopped := make([]bool, len(entries))
	// natives is every native binary a sample runs, each once: a program with several thread counts
	// shares its two.
	natives := []string{floorBinaries[0], floorBinaries[1]}
	for _, entry := range entries {
		if entry.failure == "" && !slices.Contains(natives, entry.baseline) {
			natives = append(natives, entry.baseline, entry.candidate)
		}
	}
	random := rand.New(rand.NewPCG(request.seed, 0))
	for sample := 0; sample < request.samples; sample++ {
		for _, native := range natives {
			if err := freshPages(native); err != nil {
				fail(err)
			}
		}
		runs := sampleRuns(sample)
		floorTurns := [3]int{}
		for _, runner := range runs {
			result := run(floorCommands[runner], request.timeout, padded(random, nil)...)
			floors[runner] = append(floors[runner], result)
			floorTurns[runner]++
			fmt.Fprintf(os.Stderr, "sample %d floor %s run %d: %s\n", sample+1, runnerNames[runner], floorTurns[runner], short(result))
		}
		for index, entry := range entries {
			// A row that failed to build, failed a run, or gave another answer than Node's gets no
			// verdict however many samples follow; don't spend them.
			if entry.failure != "" || stopped[index] {
				continue
			}
			commands := [3][]string{append([]string{entry.baseline}, entry.arguments...), append([]string{entry.candidate}, entry.arguments...), append(nodeCommand(entry.program, root), entry.arguments...)}
			settings := [3][]string{entry.environment, entry.environment, nil}
			turns := [3][]measurement{}
			for _, runner := range runs {
				result := run(commands[runner], request.timeout, padded(random, settings[runner])...)
				turns[runner] = append(turns[runner], result)
				fmt.Fprintf(os.Stderr, "sample %d %s %s run %d: %s\n", sample+1, entry.program.name, runnerNames[runner], len(turns[runner]), short(result))
			}
			taken := [3]measurement{combined(turns[baselineRunner]), combined(turns[candidateRunner]), combined(turns[nodeRunner])}
			samples[index] = append(samples[index], taken)
			for _, runner := range []int{baselineRunner, candidateRunner} {
				if taken[runner].failure != "" || taken[nodeRunner].failure != "" || taken[runner].stdout != taken[nodeRunner].stdout {
					stopped[index] = true
				}
			}
		}
	}

	resolution := clockResolution()
	floorSlowest := [3]measurement{}
	floorCells := []string{}
	for runner := range floors {
		floorSlowest[runner] = slowestOf(floors[runner])
		if floorSlowest[runner].failure != "" {
			fail(fmt.Errorf("the %s's floor, a program that only prints, did not run, so no row of its can be judged: %s", runnerNames[runner], floorSlowest[runner].failure))
		}
		floorCells = append(floorCells, fmt.Sprintf("%s %s to %s", runnerNames[runner], seconds(bestOf(floors[runner]).elapsed), seconds(floorSlowest[runner].elapsed)))
	}
	rank, _ := signTestRank(request.samples)
	fmt.Printf("wall clock resolution: %s; floor, a program that only prints, fastest to slowest of every run: %s\n", resolution, strings.Join(floorCells, ", "))
	fmt.Printf("%d interleaved samples of eight runs a program: Node, baseline, Node, candidate, Node, candidate, Node, baseline, the sides swapped on odd samples, so every native run follows Node and a cost carried over from the run before lands on both sides alike, and the baseline at positions 1 and 7 and the candidate at 3 and 5 (from 0) put a steady drift on both alike too. Each runner's runs in a sample are combined as their mean. Every native binary gets fresh pages before each sample, and every run an ADAMIC_BENCH_PAD of random length. The ratio is the candidate's time over the baseline's in the same sample, on the program's judge: user, CPU time in user mode across every thread, or wall. The interval is the sign test's 95%% interval for the median ratio, the %s and %s smallest of %d ratios, which assumes nothing about how times are distributed. faster needs the interval's top below 1.00, slower its bottom above 1.00. A row whose two binaries are the same is a control and gets no verdict.\n\n", request.samples, ordinal(rank), ordinal(request.samples+1-rank), request.samples)
	results := make([]programComparison, len(entries))
	controls, claims, claimedSize := 0, []string{}, 0.0
	for index, entry := range entries {
		if entry.failure != "" {
			results[index] = programComparison{failure: entry.failure}
			continue
		}
		results[index] = compareSamples(samples[index], entry.judge, floorSlowest, resolution)
		if !entry.identical {
			continue
		}
		controls++
		results[index] = asControl(results[index])
		if results[index].claimed != "" {
			claims = append(claims, fmt.Sprintf("%s %s %.3f (%.3f to %.3f)", entry.program.name, results[index].claimed, results[index].median, results[index].low, results[index].high))
			claimedSize = max(claimedSize, 1-results[index].low, results[index].high-1)
		}
	}
	// The controls go above the table, so a verdict is never read without the size of change the same
	// bytes made on this run.
	fmt.Printf("controls: %d rows run the same binary on both sides; %d would have claimed a change", controls, len(claims))
	if len(claims) > 0 {
		fmt.Printf(": %s", strings.Join(claims, ", "))
	}
	fmt.Println()
	if len(claims) > 0 {
		fmt.Printf("so a verdict on this run within %.1f%% of 1.00, the farthest a control's interval reached, is unproven: identical binaries moved that far\n", 100*claimedSize)
	}
	fmt.Println()
	fmt.Println("| program | judge | baseline, median | candidate / baseline, median | 95% interval | verdict | candidate / node, median and 95% interval |")
	fmt.Println("|---|---|---|---|---|---|---|")
	for index, entry := range entries {
		result := results[index]
		if result.failure != "" {
			fmt.Printf("| %s | %s | | | | %s | |\n", entry.program.name, entry.judge, result.failure)
			continue
		}
		fmt.Printf("| %s | %s | %s | %.3f | %.3f to %.3f | %s | %.2fx, %.2f to %.2f |\n", entry.program.name, entry.judge, seconds(result.baselineMedian), result.median, result.low, result.high, verdictSentence(result.verdict, result.low, result.high), result.nodeRatio, result.nodeLow, result.nodeHigh)
	}

	// Losses are listed first and beside the wins, so a comparison that improves one row can't hide
	// that native still trails Node on another.
	losses, wins, ties := []string{}, []string{}, []string{}
	for index, entry := range entries {
		result := results[index]
		if result.failure != "" {
			continue
		}
		line := fmt.Sprintf("- %s: %s against Node's %s on %s, %.2fx (95%% interval %.2f to %.2f)", entry.program.name, seconds(result.candidateMedian), seconds(result.nodeMedian), entry.judge, result.nodeRatio, result.nodeLow, result.nodeHigh)
		switch verdictOf(result.nodeLow, result.nodeHigh) {
		case "slower":
			losses = append(losses, line)
		case "faster":
			wins = append(wins, line)
		default:
			ties = append(ties, line)
		}
	}
	fmt.Println()
	fmt.Println("losses, where native at the candidate is slower than Node with the whole interval above 1.00 (on each program's judge):")
	printList(os.Stdout, losses)
	fmt.Println("wins, where it is faster with the whole interval below 1.00:")
	printList(os.Stdout, wins)
	fmt.Println("within the noise, where the interval holds 1.00, so neither is claimed:")
	printList(os.Stdout, ties)
	if len(harnesses) > 0 {
		fmt.Println("outside run.go, listed and not timed here:")
		for _, harness := range harnesses {
			fmt.Printf("- %s: %s on branch %s (%s): %s\n", harness.Name, harness.Path, harness.Branch, strings.Join(harness.Findings, ", "), harness.Note)
		}
	}
	fmt.Printf("\nload after: %s\n", loadAverage())
}

// loadSuite reads one suite's members, its included suites' first, and splits off the harnesses,
// which run outside run.go. only, when set, keeps the members it names.
func loadSuite(path, name, only string) ([]suiteMember, []suiteMember, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var file suiteFile
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", path, err)
	}
	members, err := suiteMembers(file, name, map[string]bool{})
	if err != nil {
		return nil, nil, err
	}
	wanted := map[string]bool{}
	for _, wantedName := range strings.Split(only, ",") {
		if wantedName != "" {
			wanted[wantedName] = true
		}
	}
	programs, harnesses := []suiteMember{}, []suiteMember{}
	for _, member := range members {
		if len(wanted) > 0 && !wanted[member.Name] {
			continue
		}
		if member.Kind == "harness" {
			harnesses = append(harnesses, member)
		} else {
			programs = append(programs, member)
		}
	}
	if len(programs) == 0 && only != "" {
		return nil, nil, fmt.Errorf("suite %s has no program named in -only %s", name, only)
	}
	if len(programs) == 0 {
		return nil, nil, fmt.Errorf("suite %s has no programs yet: %s", name, file.Suites[name].Note)
	}
	return programs, harnesses, nil
}

func suiteMembers(file suiteFile, name string, visiting map[string]bool) ([]suiteMember, error) {
	suite, found := file.Suites[name]
	if !found {
		names := []string{}
		for known := range file.Suites {
			names = append(names, known)
		}
		slices.Sort(names)
		return nil, fmt.Errorf("no suite %q in suites.json; there are %s", name, strings.Join(names, ", "))
	}
	if visiting[name] {
		return nil, fmt.Errorf("suite %s includes itself", name)
	}
	visiting[name] = true
	members := []suiteMember{}
	for _, included := range suite.Includes {
		includedMembers, err := suiteMembers(file, included, visiting)
		if err != nil {
			return nil, err
		}
		members = append(members, includedMembers...)
	}
	for _, member := range suite.Programs {
		if member.Kind != "" && member.Kind != "harness" {
			return nil, fmt.Errorf("suite %s: %s has kind %q; a member is a program, or of kind harness", name, member.Name, member.Kind)
		}
		if member.Judge != "" && member.Judge != "user" && member.Judge != "wall" {
			return nil, fmt.Errorf("suite %s: %s is judged on %q; a program is judged on user or wall", name, member.Name, member.Judge)
		}
		members = append(members, member)
	}
	visiting[name] = false
	return members, nil
}

// compareEntries finds each member's source in bench/, and makes a row for each of its thread
// counts, or one row at the pool's default.
func compareEntries(directory string, members []suiteMember) ([]compareEntry, error) {
	entries := []compareEntry{}
	seen := map[string]bool{}
	for _, member := range members {
		sources := []string{}
		for _, extension := range []string{".ts", ".a"} {
			source := filepath.Join(directory, filepath.FromSlash(member.Name)+extension)
			if _, err := os.Stat(source); err == nil {
				sources = append(sources, source)
			}
		}
		if len(sources) != 1 {
			return nil, fmt.Errorf("suite member %s must be exactly one of bench/%s.ts and bench/%s.a; found %d", member.Name, member.Name, member.Name, len(sources))
		}
		contents, err := os.ReadFile(sources[0])
		if err != nil {
			return nil, err
		}
		judgeOn := member.Judge
		if judgeOn == "" {
			judgeOn = "user"
		}
		rows := []compareEntry{{program: program{name: member.Name, source: sources[0], adamic: importsAdamic(string(contents))}, judge: judgeOn, input: member.Input}}
		if len(member.Threads) > 0 {
			rows = []compareEntry{}
			for _, threads := range member.Threads {
				if threads < 1 || threads > 4096 {
					return nil, fmt.Errorf("suite member %s: threads must be from 1 to 4096, not %d", member.Name, threads)
				}
				setting := strconv.Itoa(threads)
				rows = append(rows, compareEntry{program: program{name: member.Name + "/" + setting, source: sources[0], adamic: importsAdamic(string(contents))}, judge: judgeOn, environment: []string{"ADAMIC_THREADS=" + setting}})
			}
		}
		for _, row := range rows {
			if seen[row.program.name] {
				return nil, fmt.Errorf("%s is in the suite twice", row.program.name)
			}
			seen[row.program.name] = true
			entries = append(entries, row)
		}
	}
	return entries, nil
}

// compareSamples judges one row. Every run must have finished, and every answer must match Node's
// at both commits, before any time is read, so a wrong answer is never hidden behind a refusal; then
// every reading must pass run.go's refusals; then the ratios give the median, the interval and the
// verdict.
func compareSamples(samples [][3]measurement, judgeOn string, floors [3]measurement, resolution time.Duration) programComparison {
	for _, sample := range samples {
		for runner, result := range sample {
			if result.failure != "" {
				return programComparison{failure: "the " + runnerNames[runner] + " " + result.failure}
			}
		}
	}
	for _, sample := range samples {
		for _, runner := range []int{baselineRunner, candidateRunner} {
			if sample[runner].stdout != sample[nodeRunner].stdout {
				return programComparison{failure: "differs from Node at the " + runnerNames[runner]}
			}
		}
	}
	if _, err := signTestRank(len(samples)); err != nil {
		return programComparison{failure: err.Error()}
	}
	ratios, nodeRatios := []float64{}, []float64{}
	baselines, candidates, nodes := []time.Duration{}, []time.Duration{}, []time.Duration{}
	for _, sample := range samples {
		for runner, result := range sample {
			if judged := judge(result, floors[runner], resolution); judged.failure != "" {
				return programComparison{failure: "the " + runnerNames[runner] + " " + judged.failure}
			}
		}
		baselineTime := judgedTime(sample[baselineRunner], judgeOn)
		if baselineTime <= 0 {
			return programComparison{failure: fmt.Sprintf("the baseline's %s time read as zero, so it has no ratio", judgeOn)}
		}
		nodeTime := judgedTime(sample[nodeRunner], judgeOn)
		if nodeTime <= 0 {
			return programComparison{failure: fmt.Sprintf("Node's %s time read as zero, so it has no ratio", judgeOn)}
		}
		ratios = append(ratios, float64(judgedTime(sample[candidateRunner], judgeOn))/float64(baselineTime))
		nodeRatios = append(nodeRatios, float64(judgedTime(sample[candidateRunner], judgeOn))/float64(nodeTime))
		baselines = append(baselines, baselineTime)
		candidates = append(candidates, judgedTime(sample[candidateRunner], judgeOn))
		nodes = append(nodes, nodeTime)
	}
	median, low, high, err := signTestInterval(ratios)
	if err != nil {
		return programComparison{failure: err.Error()}
	}
	nodeMedianRatio, nodeLow, nodeHigh, err := signTestInterval(nodeRatios)
	if err != nil {
		return programComparison{failure: err.Error()}
	}
	return programComparison{median: median, low: low, high: high, verdict: verdictOf(low, high), baselineMedian: medianDuration(baselines), candidateMedian: medianDuration(candidates), nodeMedian: medianDuration(nodes), nodeRatio: nodeMedianRatio, nodeLow: nodeLow, nodeHigh: nodeHigh}
}

// asControl makes a judged row a control: its two binaries are byte for byte the same, so nothing the
// compiler did can move it, and whatever its interval says is the machine or the harness. It never gets
// a verdict; the one it would have had is kept as claimed, so the run can say how far identical bytes
// read apart. A refused row stays refused.
func asControl(result programComparison) programComparison {
	if result.failure != "" {
		return result
	}
	if result.verdict != noChange {
		result.claimed = result.verdict
	}
	result.verdict = control
	return result
}

// judgedTime is a run's time on its program's judge.
func judgedTime(result measurement, judgeOn string) time.Duration {
	if judgeOn == "wall" {
		return result.elapsed
	}
	return result.user
}

// signTestRank is k for n pairs: the largest k with P(Binomial(n, 1/2) < k) <= 0.025, so the k-th
// smallest and the k-th largest of n ratios bound a 95% interval for their median whatever their
// distribution. Below 6 pairs no k exists: even the smallest and largest ratio leave more than 2.5%
// in a tail (1/32 at 5), and the comparison is refused.
func signTestRank(pairs int) (int, error) {
	// P(B < k) is the sum of C(n, i) for i below k, over 2^n, and it's at most 1/40 exactly when 40
	// times that sum is at most 2^n: integers, so no rounding decides a boundary.
	total := new(big.Int).Lsh(big.NewInt(1), uint(max(pairs, 0)))
	cumulative := new(big.Int)
	binomial := big.NewInt(1) // C(n, 0)
	rank := 0
	for k := 1; k <= pairs; k++ {
		cumulative.Add(cumulative, binomial)
		if new(big.Int).Mul(cumulative, big.NewInt(40)).Cmp(total) > 0 {
			break
		}
		rank = k
		// C(n, k) from C(n, k-1).
		binomial.Mul(binomial, big.NewInt(int64(pairs-k+1)))
		binomial.Quo(binomial, big.NewInt(int64(k)))
	}
	if rank < 1 {
		return 0, fmt.Errorf("%d samples can't give a 95%% interval: the sign test needs at least 6, since below that even the smallest and largest ratio leave more than 2.5%% in a tail", pairs)
	}
	return rank, nil
}

// signTestInterval is the median ratio and its sign-test 95% interval.
func signTestInterval(ratios []float64) (float64, float64, float64, error) {
	rank, err := signTestRank(len(ratios))
	if err != nil {
		return 0, 0, 0, err
	}
	sorted := slices.Clone(ratios)
	slices.Sort(sorted)
	return medianOf(sorted), sorted[rank-1], sorted[len(sorted)-rank], nil
}

// verdictOf reads an interval: faster or slower only when all of it is on one side of 1.00.
func verdictOf(low, high float64) string {
	if high < 1 {
		return "faster"
	}
	if low > 1 {
		return "slower"
	}
	return noChange
}

// verdictSentence is a verdict with the change the run could see: "no change" names the interval's
// farther bound from 1.00, the smallest change in either direction this run couldn't rule out, so it
// never reads as proof that nothing changed.
func verdictSentence(verdict string, low, high float64) string {
	if verdict != noChange {
		return verdict
	}
	return fmt.Sprintf("%s, none larger than %.1f%%", noChange, 100*max(1-low, high-1))
}

// medianOf is the middle of sorted values, or the mean of the middle two.
func medianOf(sorted []float64) float64 {
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func medianDuration(durations []time.Duration) time.Duration {
	values := []float64{}
	for _, duration := range durations {
		values = append(values, float64(duration))
	}
	slices.Sort(values)
	return time.Duration(medianOf(values))
}

// resolveCommit is a commit's full sha, so the header names exactly what was built.
func resolveCommit(revision string) (string, error) {
	output, err := exec.Command("git", "rev-parse", "--verify", revision+"^{commit}").Output()
	if err != nil {
		return "", fmt.Errorf("%q is not a commit here: %w", revision, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// buildAdamicAt builds cmd/adamic as it stands at a commit, in a detached worktree removed once the
// binary is built: the binary embeds its runtime, so it needs nothing from the tree after.
func buildAdamicAt(root, commit, binary string) error {
	parent, err := os.MkdirTemp("", "adamic-compare-tree-")
	if err != nil {
		return err
	}
	defer os.Remove(parent)
	tree := filepath.Join(parent, "tree")
	add := exec.Command("git", "worktree", "add", "--detach", tree, commit)
	add.Dir = root
	if output, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add at %s: %v\n%s", commit, err, output)
	}
	defer func() {
		// layCohere's link goes first, back to the empty directory git left, so the worktree reads
		// clean and git removes it without --force.
		cohere := filepath.Join(tree, "cohere")
		if information, err := os.Lstat(cohere); err == nil && information.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(cohere); err == nil {
				_ = os.Mkdir(cohere, 0755)
			}
		}
		remove := exec.Command("git", "worktree", "remove", tree)
		remove.Dir = root
		if output, err := remove.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "bench: git worktree remove %s: %v %s; remove it by hand\n", tree, err, firstLine(string(output)))
		}
	}()
	if err := layCohere(root, tree, commit); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./cmd/adamic")
	build.Dir = tree
	if output, err := step("building adamic at "+commit, build); err != nil {
		return fmt.Errorf("building adamic at %s: %v\n%s", commit, err, output)
	}
	return nil
}

// layCohere fills a fresh worktree's cohere, which go.mod's replace directives build from: a worktree
// leaves a submodule's directory empty (M1's first run, J-fywr4hc, failed to build there). When the
// commit pins the cohere this checkout has laid, the worktree's cohere becomes a link to it; when the
// pins differ the comparison is refused, naming both, since building one commit against another's
// checker would compare the wrong compiler. A cohere already holding files (a tree whose cohere was
// committed as files, as a sealed job's is) is left as it is.
func layCohere(root, tree, commit string) error {
	target := filepath.Join(tree, "cohere")
	entries, err := os.ReadDir(target)
	if err == nil && len(entries) > 0 {
		return nil
	}
	pin := func(revision string) (string, error) {
		command := exec.Command("git", "rev-parse", revision+":cohere")
		command.Dir = root
		output, err := command.Output()
		return strings.TrimSpace(string(output)), err
	}
	wanted, err := pin(commit)
	if err != nil {
		return fmt.Errorf("reading %s's cohere pin: %w", commit, err)
	}
	laid, err := pin("HEAD")
	if err != nil {
		return fmt.Errorf("reading this checkout's cohere pin: %w", err)
	}
	if wanted != laid {
		return fmt.Errorf("%s pins cohere %s, and this checkout has %s laid: lay the commit's cohere to compare it", commit, wanted, laid)
	}
	if _, err := os.Stat(filepath.Join(root, "cohere", "go.mod")); err != nil {
		return fmt.Errorf("this checkout's cohere isn't laid (no cohere/go.mod): %w", err)
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing the worktree's empty cohere: %w", err)
	}
	return os.Symlink(filepath.Join(root, "cohere"), target)
}

// makeInput builds an input's generator natively and runs it once with the size as its argument, its
// standard output becoming the input file.
func makeInput(directory, work, adamic string, input programInput, key string) (string, error) {
	sources := []string{}
	for _, extension := range []string{".ts", ".a"} {
		source := filepath.Join(directory, filepath.FromSlash(input.Generator)+extension)
		if _, err := os.Stat(source); err == nil {
			sources = append(sources, source)
		}
	}
	if len(sources) != 1 {
		return "", fmt.Errorf("input generator %s must be exactly one of bench/%s.ts and bench/%s.a; found %d", input.Generator, input.Generator, input.Generator, len(sources))
	}
	if input.Size < 1 {
		return "", fmt.Errorf("input of %s: size must be positive, not %d", input.Generator, input.Size)
	}
	if err := os.MkdirAll(filepath.Join(work, "inputs"), 0755); err != nil {
		return "", err
	}
	generator := filepath.Join(work, "inputs", key+".generator")
	if output, err := step("building the input generator "+input.Generator, exec.Command(adamic, "build", sources[0], "-o", generator)); err != nil {
		return "", fmt.Errorf("building the input generator %s failed: %v %s", input.Generator, err, firstLine(string(output)))
	}
	path := filepath.Join(work, "inputs", key+".txt")
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var stderr bytes.Buffer
	command := exec.Command(generator, strconv.Itoa(input.Size))
	command.Stdout, command.Stderr = file, &stderr
	if _, err := step(fmt.Sprintf("making the input with %s at size %d", input.Generator, input.Size), command); err != nil {
		return "", fmt.Errorf("making the input with %s at size %d failed: %v %s", input.Generator, input.Size, err, firstLine(stderr.String()))
	}
	return path, nil
}

// stepLimit bounds each step a comparison takes before its first sample: building adamic at a commit,
// building a program, and making an input. A sample's runs have -timeout; these had nothing, so a step
// that never ended held the box until the job's own ceiling (J-6rff648 and J-dj0awxm each spent their
// hour before a first sample, and said nothing of where).
const stepLimit = 20 * time.Minute

// step runs one such step: it says on stderr what it starts and how long it took, so a stalled run
// shows the step it stalled in, and kills it at stepLimit, reported as an error. A command whose output
// isn't already directed gives its combined output back, for the caller's message.
func step(what string, command *exec.Cmd) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "bench: %s\n", what)
	var output bytes.Buffer
	if command.Stdout == nil {
		command.Stdout = &output
	}
	if command.Stderr == nil {
		command.Stderr = &output
	}
	start := time.Now()
	if err := command.Start(); err != nil {
		return nil, err
	}
	timer := time.AfterFunc(stepLimit, func() { _ = command.Process.Kill() })
	err := command.Wait()
	if !timer.Stop() && err != nil {
		fmt.Fprintf(os.Stderr, "bench: %s didn't finish in %s\n", what, stepLimit)
		return output.Bytes(), fmt.Errorf("didn't finish in %s", stepLimit)
	}
	fmt.Fprintf(os.Stderr, "bench: %s took %.1f s\n", what, time.Since(start).Seconds())
	return output.Bytes(), err
}

// combined is one runner's runs in a sample as one measurement: the mean of their wall and of their
// user times, so a sample's time stays comparable to one run of the floor and to the clock's
// resolution, and the larger peak. A run's failure is the sample's, naming the run, and every run must
// print the same answer, or the sample fails naming where they part: a program whose answer changes
// between two runs of the same binary has no time worth judging.
func combined(runs []measurement) measurement {
	var elapsed, user time.Duration
	result := measurement{stdout: runs[0].stdout}
	for index, turn := range runs {
		if turn.failure != "" {
			return measurement{failure: fmt.Sprintf("in run %d of %d %s", index+1, len(runs), turn.failure)}
		}
		if turn.stdout != runs[0].stdout {
			return measurement{failure: fmt.Sprintf("printed %q in run %d of %d and %q in run 1, the same binary giving two answers", firstLine(turn.stdout), index+1, len(runs), firstLine(runs[0].stdout))}
		}
		elapsed += turn.elapsed
		user += turn.user
		result.peak = max(result.peak, turn.peak)
	}
	result.elapsed = elapsed / time.Duration(len(runs))
	result.user = user / time.Duration(len(runs))
	return result
}

// freshPages gives a native binary new pages in the page cache without moving it: a copy beside it,
// renamed over it, is a new file whose pages sit wherever the kernel puts them now. A binary left alone
// keeps the same physical pages for the whole run, and with them the same cache aliasing, a constant
// tied to its side that no number of samples averages away.
func freshPages(path string) error {
	source, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("fresh pages for %s: %w", path, err)
	}
	defer source.Close()
	copied, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".fresh-")
	if err != nil {
		return fmt.Errorf("fresh pages for %s: %w", path, err)
	}
	if _, err := io.Copy(copied, source); err != nil {
		copied.Close()
		os.Remove(copied.Name())
		return fmt.Errorf("fresh pages for %s, copying: %w", path, err)
	}
	if err := copied.Close(); err != nil {
		os.Remove(copied.Name())
		return fmt.Errorf("fresh pages for %s, closing the copy: %w", path, err)
	}
	if err := os.Chmod(copied.Name(), 0755); err != nil {
		os.Remove(copied.Name())
		return fmt.Errorf("fresh pages for %s, making the copy executable: %w", path, err)
	}
	if err := os.Rename(copied.Name(), path); err != nil {
		os.Remove(copied.Name())
		return fmt.Errorf("fresh pages for %s, renaming the copy over it: %w", path, err)
	}
	return nil
}

// padded is a run's settings with ADAMIC_BENCH_PAD added, a value of a random length from 0 to 4095
// drawn for this run alone. The environment's size decides where the stack starts, and every run with
// the same environment starts it at the same alignment, a constant tied to a side; a random size spreads
// it over every alignment on both sides alike. The settings themselves are copied, never appended to.
func padded(random *rand.Rand, settings []string) []string {
	return append(slices.Clone(settings), "ADAMIC_BENCH_PAD="+strings.Repeat("x", random.IntN(4096)))
}

// plantedCopy writes a copy of a source with a plant of milliseconds appended.
func plantedCopy(source, planted string, milliseconds int) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(planted, append(contents, plantLoop(milliseconds)...), 0644)
}

// busyProcesses names every process over 5% CPU, from ps. On Linux the figure is the process's
// share over its life, and on macOS a recent decaying average, so a long-lived daemon that just woke
// can read low; it's what the box was doing, not proof it was quiet.
func busyProcesses() string {
	output, err := exec.Command("ps", "-eo", "pcpu=,comm=").Output()
	if err != nil {
		return "unknown: " + err.Error()
	}
	busy := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		share, err := strconv.ParseFloat(fields[0], 64)
		if err != nil || share <= 5 {
			continue
		}
		busy = append(busy, fmt.Sprintf("%s (%s%%)", strings.Join(fields[1:], " "), fields[0]))
	}
	if len(busy) == 0 {
		return "none"
	}
	return strings.Join(busy, ", ")
}

func ordinal(number int) string {
	suffix := "th"
	if number%100 < 11 || number%100 > 13 {
		switch number % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(number) + suffix
}

// printList writes each line, or "- none" when there are none, so an empty list still reads as looked at.
func printList(out io.Writer, lines []string) {
	if len(lines) == 0 {
		fmt.Fprintln(out, "- none")
		return
	}
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
}
