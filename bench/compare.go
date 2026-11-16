// Compare times one suite of bench/suites.json at two commits of Adamic, and says whether the
// candidate is faster, slower, or no different at 95% confidence, with the variation that verdict
// rests on.
//
//	go run ./bench -compare <baseline sha> <candidate sha> -suite everyday [-samples 15] [-only trees] [-plant candidate]
//
// cmd/adamic is built at each commit in a detached worktree, and every program in the suite is built
// with each of the two compilers from the same source, the one in this checkout, so the compiler is
// the only thing that differs. Each sample runs the baseline, the candidate and Node once each, in an
// order that rotates every sample, so no runner always goes first on a warm or a cold machine.
//
// The statistic is paired: per sample, the candidate's time over the baseline's on the program's
// judge, user time unless the suite says wall (a parallel program spends more user time with more
// threads by design). The median of those ratios is reported with the sign test's 95% interval, the
// k-th smallest and k-th largest ratio, which assumes nothing about how the times are distributed. A
// verdict needs the whole interval on one side of 1.00; a program whose answer differs from Node's at
// either commit gets none, and so does a run refused the way run.go refuses one (at or below the
// clock's resolution, or within its runtime's floor).
//
// -plant candidate proves the comparison can fail: it builds each candidate program from a copy of
// its source with a fixed 25 ms busy loop appended, so an A/A comparison with the plant must read
// slower. The plant lives here, in the harness, and never in the compiler or the runtime.
package main

import (
	"encoding/json"
	"fmt"
	"math/big"
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
	only      string
	timeout   time.Duration
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
	Branch   string   `json:"branch"`
	Note     string   `json:"note"`
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
	baseline    string
	candidate   string
	failure     string // why it can't be compared, from a build
}

// The three runners of every sample, in the order the first sample runs them.
const (
	baselineRunner = iota
	candidateRunner
	nodeRunner
)

var runnerNames = [...]string{"baseline", "candidate", "node"}

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
}

const noChange = "no change at this confidence"

// plantLoop is the plant: a fixed 25 ms of busy CPU, the same whatever the program's size, appended to
// a copy of a candidate program's source, so it runs inside the timed process at the end of the
// program's synchronous code (before any timer the program set fires). It prints nothing, so the
// planted program still gives Node's answer, and only its time can tell it apart.
const plantLoop = `
// Planted by go run ./bench -compare -plant candidate (bench/compare.go): a fixed 25 ms of busy CPU.
const adamicBenchPlantStart = Date.now();
while (Date.now() - adamicBenchPlantStart < 25) {
	// Spins, printing nothing.
}
`

func compareCommits(request compareRequest) {
	if _, err := signTestRank(request.samples); err != nil {
		fail(err)
	}
	if request.plant != "" && request.plant != "candidate" {
		fail(fmt.Errorf("-plant takes only candidate, not %q", request.plant))
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

	describeMachine([]runtimeKind{{name: runnerNames[baselineRunner]}, {name: runnerNames[candidateRunner]}, {name: runnerNames[nodeRunner]}})
	fmt.Printf("baseline: %s\ncandidate: %s\n", baselineCommit, candidateCommit)
	plantDescription := "none"
	if request.plant == "candidate" {
		plantDescription = "candidate, a fixed 25 ms busy loop appended to every candidate program"
	}
	fmt.Printf("suite: %s, %d rows; samples: %d; plant: %s\n", request.suite, len(entries), request.samples, plantDescription)
	fmt.Printf("load before: %s\n", loadAverage())
	fmt.Printf("processes over 5%% CPU at start: %s\n\n", busyProcesses())

	baselineAdamic := filepath.Join(work, "adamic-baseline")
	if err := buildAdamicAt(root, baselineCommit, baselineAdamic); err != nil {
		fail(err)
	}
	candidateAdamic := filepath.Join(work, "adamic-candidate")
	if err := buildAdamicAt(root, candidateCommit, candidateAdamic); err != nil {
		fail(err)
	}

	floorSource := filepath.Join(work, "floor.ts")
	if err := os.WriteFile(floorSource, []byte("console.log(\"0\");\n"), 0644); err != nil {
		fail(err)
	}
	floor := program{name: "floor", source: floorSource}
	floorBinaries := [2]string{filepath.Join(work, "floor-baseline"), filepath.Join(work, "floor-candidate")}
	for index, adamic := range []string{baselineAdamic, candidateAdamic} {
		if output, err := exec.Command(adamic, "build", floor.source, "-o", floorBinaries[index]).CombinedOutput(); err != nil {
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
		entry.baseline = filepath.Join(work, base+".baseline")
		entry.candidate = filepath.Join(work, base+".candidate")
		candidateSource := entry.program.source
		if request.plant == "candidate" {
			candidateSource = filepath.Join(plantedDirectory, base+filepath.Ext(entry.program.source))
			err = plantedCopy(entry.program.source, candidateSource)
			if err != nil {
				fail(err)
			}
		}
		if output, err := exec.Command(baselineAdamic, "build", entry.program.source, "-o", entry.baseline).CombinedOutput(); err != nil {
			entry.failure = fmt.Sprintf("the baseline's build failed: %v %s", err, firstLine(string(output)))
		} else if output, err := exec.Command(candidateAdamic, "build", candidateSource, "-o", entry.candidate).CombinedOutput(); err != nil {
			entry.failure = fmt.Sprintf("the candidate's build failed: %v %s", err, firstLine(string(output)))
		}
		built[entry.program.source] = *entry
	}

	// samples[entry] holds each sample's three measurements, by runner.
	samples := make([][][3]measurement, len(entries))
	// floors[runner] holds every sample's run of the floor program.
	floors := [3][]measurement{}
	floorCommands := [3][]string{{floorBinaries[0]}, {floorBinaries[1]}, nodeCommand(floor, root)}
	stopped := make([]bool, len(entries))
	for sample := 0; sample < request.samples; sample++ {
		order := [3]int{sample % 3, (sample + 1) % 3, (sample + 2) % 3}
		for _, runner := range order {
			result := run(floorCommands[runner], request.timeout)
			floors[runner] = append(floors[runner], result)
			fmt.Fprintf(os.Stderr, "sample %d floor %s: %s\n", sample+1, runnerNames[runner], short(result))
		}
		for index, entry := range entries {
			// A row that failed to build, failed a run, or gave another answer than Node's gets no
			// verdict however many samples follow; don't spend them.
			if entry.failure != "" || stopped[index] {
				continue
			}
			commands := [3][]string{{entry.baseline}, {entry.candidate}, nodeCommand(entry.program, root)}
			settings := [3][]string{entry.environment, entry.environment, nil}
			taken := [3]measurement{}
			for _, runner := range order {
				taken[runner] = run(commands[runner], request.timeout, settings[runner]...)
				fmt.Fprintf(os.Stderr, "sample %d %s %s: %s\n", sample+1, entry.program.name, runnerNames[runner], short(taken[runner]))
			}
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
	fmt.Printf("wall clock resolution: %s; floor, a program that only prints, fastest to slowest of %d: %s\n", resolution, request.samples, strings.Join(floorCells, ", "))
	fmt.Printf("%d interleaved samples, the runners' order rotating each sample. The ratio is the candidate's time over the baseline's in the same sample, on the program's judge: user, CPU time in user mode across every thread, or wall. The interval is the sign test's 95%% interval for the median ratio, the %s and %s smallest of %d ratios, which assumes nothing about how times are distributed. faster needs the interval's top below 1.00, slower its bottom above 1.00.\n\n", request.samples, ordinal(rank), ordinal(request.samples+1-rank), request.samples)
	fmt.Println("| program | judge | candidate / baseline, median | 95% interval | verdict | candidate vs node, median |")
	fmt.Println("|---|---|---|---|---|---|")
	results := make([]programComparison, len(entries))
	for index, entry := range entries {
		if entry.failure != "" {
			results[index] = programComparison{failure: entry.failure}
		} else {
			results[index] = compareSamples(samples[index], entry.judge, floorSlowest, resolution)
		}
		result := results[index]
		if result.failure != "" {
			fmt.Printf("| %s | %s | | | %s | |\n", entry.program.name, entry.judge, result.failure)
			continue
		}
		fmt.Printf("| %s | %s | %.3f | %.3f to %.3f | %s | %.2fx |\n", entry.program.name, entry.judge, result.median, result.low, result.high, result.verdict, result.candidateMedian.Seconds()/result.nodeMedian.Seconds())
	}

	// Losses are listed first and beside the wins, so a comparison that improves one row can't hide
	// that native still trails Node on another.
	losses, wins := []string{}, []string{}
	for index, entry := range entries {
		result := results[index]
		if result.failure != "" {
			continue
		}
		line := fmt.Sprintf("- %s: %s against Node's %s on %s, %.2fx", entry.program.name, seconds(result.candidateMedian), seconds(result.nodeMedian), entry.judge, result.candidateMedian.Seconds()/result.nodeMedian.Seconds())
		if result.candidateMedian > result.nodeMedian {
			losses = append(losses, line)
		} else {
			wins = append(wins, line)
		}
	}
	fmt.Println()
	fmt.Println("losses, where native at the candidate is slower than Node (medians, on each program's judge):")
	printList(losses)
	fmt.Println("wins, where it is no slower:")
	printList(wins)
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
		rows := []compareEntry{{program: program{name: member.Name, source: sources[0], adamic: importsAdamic(string(contents))}, judge: judgeOn}}
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
	ratios := []float64{}
	candidates, nodes := []time.Duration{}, []time.Duration{}
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
		ratios = append(ratios, float64(judgedTime(sample[candidateRunner], judgeOn))/float64(baselineTime))
		candidates = append(candidates, judgedTime(sample[candidateRunner], judgeOn))
		nodes = append(nodes, judgedTime(sample[nodeRunner], judgeOn))
	}
	median, low, high, err := signTestInterval(ratios)
	if err != nil {
		return programComparison{failure: err.Error()}
	}
	return programComparison{median: median, low: low, high: high, verdict: verdictOf(low, high), candidateMedian: medianDuration(candidates), nodeMedian: medianDuration(nodes)}
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
		remove := exec.Command("git", "worktree", "remove", tree)
		remove.Dir = root
		if output, err := remove.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "bench: git worktree remove %s: %v %s; remove it by hand\n", tree, err, firstLine(string(output)))
		}
	}()
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./cmd/adamic")
	build.Dir = tree
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("building adamic at %s: %v\n%s", commit, err, output)
	}
	return nil
}

// plantedCopy writes a copy of a source with the plant appended.
func plantedCopy(source, planted string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(planted, append(contents, plantLoop...), 0644)
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

func printList(lines []string) {
	if len(lines) == 0 {
		fmt.Println("- none")
		return
	}
	for _, line := range lines {
		fmt.Println(line)
	}
}
