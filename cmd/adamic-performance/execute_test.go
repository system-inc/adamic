package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/benchmark"
)

// fakeOffsets are each run's milliseconds over its runtime's base, by how many runs of that program the runtime has had:
// out of order, so a spread that kept the first or the last round reads wrong, and with a median (12) that isn't their
// mean (14), so a mean reads wrong too.
var fakeOffsets = []int64{30, 0, 12, 40}

// A fakeClock stands in for the reaper and both runtimes: it records every command it's handed, and measures each a
// fixed function of the runtime and the run, printing Node's answer but where wrong or failing names the run, as
// "<side> <program> <run>", that prints another or fails.
type fakeClock struct {
	commands [][]string
	runs     []string
	counts   map[string]int
	wrong    string
	failing  string
}

func (fake *fakeClock) run(command []string) benchmark.Measurement {
	fake.commands = append(fake.commands, command)
	side, program := "native", filepath.Base(command[0])
	if command[0] == "node" {
		source := command[len(command)-1]
		side, program = "node", strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	}
	fake.counts[side+" "+program]++
	run := fake.counts[side+" "+program]
	fake.runs = append(fake.runs, side+" "+program)
	named := fmt.Sprintf("%s %s %d", side, program, run)
	if named == fake.failing {
		return benchmark.Measurement{Failure: "failed: exit status 1 out of memory"}
	}
	stdout := "primes: 664579\ntotal: 42\n"
	if named == fake.wrong {
		stdout = "primes: 664579\ntotal: 41\n"
	}
	offset := fakeOffsets[run-1]
	if side == "native" {
		return benchmark.Measurement{Elapsed: time.Duration(100+offset) * time.Millisecond, User: time.Duration(90+offset) * time.Millisecond, Peak: (2000 + offset) * 1024, Stdout: stdout}
	}
	return benchmark.Measurement{Elapsed: time.Duration(200+offset) * time.Millisecond, User: time.Duration(210+offset) * time.Millisecond, Peak: (70000 + offset) * 1024, Stdout: stdout}
}

// executeTree is a stand-in repository holding every execute program, word_count importing 'adamic', so Node runs it
// through the oracle runner as it runs every .a source.
func executeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, source := range executeSources {
		path := filepath.Join(root, filepath.FromSlash(source.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		contents := "console.log(42);\n"
		if source.name == "word_count" {
			contents = "import { parallelMap } from 'adamic';\n" + contents
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// The whole execute half through a fake runner and a fake clock: the seven programs the compile half doesn't build are
// built once each on its caches, tokenizer and trees never, since its warm builds left their binaries. Then the rounds:
// in round 1 and round 3 every program in order, native before Node, and in round 2 the programs backwards, Node
// before native. Native runs its binary alone, and Node a .a or an 'adamic' import through the oracle runner and any
// other source directly. Each number is the median of the rounds with their ends, the peak the median of the rounds'
// peaks in KiB, and the ratio native's judged median over Node's: user for trees, wall for primes_parallel. Mutants:
// the odd round left unmirrored, or mirrored in its programs alone; tokenizer rebuilt; the median taken as the mean,
// or as the last round's; the ratio taken on user time for a wall-judged program.
func TestPerformanceExecuteInterleavesAndMirrorsItsRounds(t *testing.T) {
	t.Parallel()
	root, work := executeTree(t), t.TempDir()
	build := &fakeAdamic{work: work, warmBuilds: map[string]int{}}
	clock := &fakeClock{counts: map[string]int{}}
	measured, archivesBuilt, err := measureExecute(executeRequest{compile: compileRequest{root: root, work: work, rounds: 5}, rounds: 3}, build.run, clock.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(archivesBuilt) != 0 || archivesLine(archivesBuilt) != "" {
		t.Errorf("%v built runtime archives, and every build found hello's", archivesBuilt)
	}
	wantBuilds := []string{"warm devirtualize.a 1", "warm nbody.ts 1", "warm parallel_files.a 1", "warm primes_parallel.a 1", "warm sort.ts 1", "warm spectral_norm.ts 1", "warm word_count.ts 1"}
	if !slices.Equal(build.children, wantBuilds) {
		t.Errorf("builds ran as\n%s\nwant\n%s", strings.Join(build.children, "\n"), strings.Join(wantBuilds, "\n"))
	}
	var forward []string
	for _, source := range executeSources {
		forward = append(forward, source.name)
	}
	var want []string
	for round := range 3 {
		if round == 1 {
			for _, name := range slices.Backward(forward) {
				want = append(want, "node "+name, "native "+name)
			}
			continue
		}
		for _, name := range forward {
			want = append(want, "native "+name, "node "+name)
		}
	}
	if !slices.Equal(clock.runs, want) {
		t.Errorf("runs went\n%s\nwant\n%s", strings.Join(clock.runs, "\n"), strings.Join(want, "\n"))
	}
	for _, command := range clock.commands {
		if command[0] == "node" {
			source := command[len(command)-1]
			throughOracle := strings.HasSuffix(source, ".a") || strings.HasSuffix(source, "word_count.ts")
			if wantCommand := benchmark.NodeCommand(source, throughOracle, root); !slices.Equal(command, wantCommand) {
				t.Errorf("Node ran %q, want %q", command, wantCommand)
			}
			continue
		}
		if len(command) != 1 || filepath.Dir(command[0]) != filepath.Join(work, "out") {
			t.Errorf("native ran %q, want its binary in %s alone", command, filepath.Join(work, "out"))
		}
	}

	if measured.Rounds != 3 || len(measured.Programs) != len(executeSources) {
		t.Fatalf("rounds %d and %d programs", measured.Rounds, len(measured.Programs))
	}
	for index, entry := range measured.Programs {
		if entry.Name != executeSources[index].name || entry.Execution.Judge != executeSources[index].judge {
			t.Errorf("program %d is %s judged on %s, want %s on %s", index, entry.Name, entry.Execution.Judge, executeSources[index].name, executeSources[index].judge)
		}
	}
	native := runtimeNumbers{Wall: spread{Median: 112, Min: 100, Max: 130}, User: spread{Median: 102, Min: 90, Max: 120}, PeakKiB: 2012}
	node := runtimeNumbers{Wall: spread{Median: 212, Min: 200, Max: 230}, User: spread{Median: 222, Min: 210, Max: 240}, PeakKiB: 70012}
	for name, wantEntry := range map[string]execution{
		"trees":           {Judge: "user", Native: native, Node: node, Ratio: 0.46},
		"primes_parallel": {Judge: "wall", Native: native, Node: node, Ratio: 0.53},
	} {
		if got, _ := measured.Programs.lookup(name); got != wantEntry {
			t.Errorf("%s is %+v, want %+v", name, got, wantEntry)
		}
	}
}

// The ratio is native's judged median over Node's, rounded to the nearest hundredth either way, below 1 meaning native
// is faster. Mutants: truncated to two decimals (2 over 3 reads 0.66), or Node over native.
func TestPerformanceExecuteRatioRoundsToTwoDecimals(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		native, node int64
		want         float64
	}{
		{native: 1608, node: 1504, want: 1.07},
		{native: 2, node: 3, want: 0.67},
		{native: 1301, node: 2032, want: 0.64},
		{native: 1004, node: 1000, want: 1},
		{native: 1006, node: 1000, want: 1.01},
		{native: 198, node: 948, want: 0.21},
	} {
		if got := ratioOf(each.native, each.node); got != each.want {
			t.Errorf("ratioOf(%d, %d) is %v, want %v", each.native, each.node, got, each.want)
		}
	}
}

// A run whose answer isn't Node's first, natively or on Node itself, is an error naming the program, the runtime, the
// round and the line they part at, and so is a run that failed: nothing is measured, so measureSnapshot returns the
// error and write writes no file. Mutants: the agreement check dropped, and the measure returns numbers; native held to
// Node's answer and Node not held to its own.
func TestPerformanceExecuteDisagreementIsAnError(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		wrong, failing string
		want           string
	}{
		{wrong: "native primes_parallel 2", want: `primes_parallel: native's run in round 2 printed another answer than Node's in round 1, at line 2, "total: 41" against "total: 42"`},
		{wrong: "node trees 3", want: "trees: node's run in round 3 printed another answer than Node's in round 1"},
		{failing: "node sort 1", want: "sort: node's run in round 1 failed: exit status 1"},
	} {
		root, work := executeTree(t), t.TempDir()
		build := &fakeAdamic{work: work, warmBuilds: map[string]int{}}
		clock := &fakeClock{counts: map[string]int{}, wrong: each.wrong, failing: each.failing}
		measured, _, err := measureExecute(executeRequest{compile: compileRequest{root: root, work: work, rounds: 5}, rounds: 3}, build.run, clock.run)
		if err == nil || !strings.Contains(err.Error(), each.want) {
			t.Errorf("error %v, want it to name %q", err, each.want)
		}
		if measured.Rounds != 0 || len(measured.Programs) != 0 {
			t.Errorf("with %q wrong and %q failing, the measure returned %+v", each.wrong, each.failing, measured)
		}
	}
}

// A build for the execute half whose runtime archive misses is named and measured on, since a program's features key
// its archive and no execute run is a build, and the line beside the file counts and names them in executeSources'
// order; one whose log has no cArchive line never reached the archive, and is an error naming the program. Mutants: a
// miss refused, as a warm build's is, and the measure fails; the hits named instead; the missing line taken as a miss.
func TestPerformanceExecuteBuildsCountTheArchivesTheyBuild(t *testing.T) {
	t.Parallel()
	root, work := executeTree(t), t.TempDir()
	build := &fakeAdamic{work: work, warmBuilds: map[string]int{}, archiveMisses: []string{"parallel_files.a 1", "sort.ts 1", "primes_parallel.a 1"}}
	clock := &fakeClock{counts: map[string]int{}}
	measured, archivesBuilt, err := measureExecute(executeRequest{compile: compileRequest{root: root, work: work, rounds: 5}, rounds: 3}, build.run, clock.run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(archivesBuilt, []string{"parallel_files", "primes_parallel", "sort"}) || len(measured.Programs) != len(executeSources) {
		t.Errorf("%v built runtime archives and %d programs were measured, want parallel_files, primes_parallel and sort, and %d", archivesBuilt, len(measured.Programs), len(executeSources))
	}
	if got, want := archivesLine(archivesBuilt), "adamic-performance: execute builds: 3 runtime archives built (their feature sets differ from hello's): parallel_files, primes_parallel, sort\n"; got != want {
		t.Errorf("line %q, want %q", got, want)
	}

	root, work = executeTree(t), t.TempDir()
	build = &fakeAdamic{work: work, warmBuilds: map[string]int{}, missingHit: "nbody.ts 1"}
	clock = &fakeClock{counts: map[string]int{}}
	_, _, err = measureExecute(executeRequest{compile: compileRequest{root: root, work: work, rounds: 5}, rounds: 3}, build.run, clock.run)
	if err == nil || !strings.Contains(err.Error(), "nbody: its build logged no cArchive line") {
		t.Fatalf("error %v, want nbody's build named", err)
	}
	if len(clock.runs) != 0 {
		t.Errorf("%d runs were timed after a build failed", len(clock.runs))
	}
}

// The execute programs are bench/suites.json's everyday programs, plus primes_parallel judged on wall, each with its
// suite's judge (user where it names none), each at bench/<name>.ts or bench/<name>.a and not both, as bench -compare
// finds them. A member with threads or an input runs otherwise than this half runs it, so none may have one. Mutants:
// a program added to everyday and not here; parallel_files judged on user.
func TestPerformanceExecuteProgramsAreTheEverydaySuite(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	contents, err := os.ReadFile(filepath.Join(root, "bench", "suites.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Suites map[string]struct {
			Includes []string `json:"includes"`
			Programs []struct {
				Name    string          `json:"name"`
				Kind    string          `json:"kind"`
				Judge   string          `json:"judge"`
				Threads []int           `json:"threads"`
				Input   json.RawMessage `json:"input"`
			} `json:"programs"`
		} `json:"suites"`
	}
	if err := json.Unmarshal(contents, &file); err != nil {
		t.Fatal(err)
	}
	everyday, found := file.Suites["everyday"]
	if !found || len(everyday.Includes) > 0 {
		t.Fatalf("bench/suites.json's everyday is missing or includes %v, and this test reads its own programs alone", everyday.Includes)
	}
	want := []executeSource{}
	for _, member := range everyday.Programs {
		if member.Kind != "" || len(member.Threads) > 0 || member.Input != nil {
			t.Errorf("everyday's %s has a kind, threads or an input, so it runs otherwise than the execute half runs it", member.Name)
		}
		judge := member.Judge
		if judge == "" {
			judge = "user"
		}
		want = append(want, executeSource{name: member.Name, judge: judge})
	}
	want = append(want, executeSource{name: "primes_parallel", judge: "wall"})
	got := []executeSource{}
	for _, source := range executeSources {
		found := 0
		for _, extension := range []string{".ts", ".a"} {
			if _, err := os.Stat(filepath.Join(root, "bench", source.name+extension)); err == nil {
				found++
				if source.path != "bench/"+source.name+extension {
					t.Errorf("%s's path is %s, and its source is bench/%s%s", source.name, source.path, source.name, extension)
				}
			}
		}
		if found != 1 {
			t.Errorf("%s must be exactly one of bench/%s.ts and bench/%s.a; found %d", source.name, source.name, source.name, found)
		}
		got = append(got, executeSource{name: source.name, judge: source.judge})
	}
	byName := func(left, right executeSource) int { return strings.Compare(left.name, right.name) }
	slices.SortFunc(want, byName)
	slices.SortFunc(got, byName)
	if !slices.Equal(got, want) {
		t.Errorf("the execute programs are %+v, want %+v", got, want)
	}
}
