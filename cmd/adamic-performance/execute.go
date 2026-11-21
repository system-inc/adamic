package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/benchmark"
)

// executeTimeout is how long one run of an execute program may take, bench's own -timeout default, past which the run
// is a failure and nothing is written.
const executeTimeout = 120 * time.Second

// An executeSource is one program the execute half runs: its name in the file, its path from the repository's root,
// and the measure it is judged on, user or wall.
type executeSource struct {
	name  string
	path  string
	judge string
}

// executeSources are the execute half's programs, in the file's order: bench/suites.json's everyday programs with its
// judges, and primes_parallel, Kirk's parallel demo (#sp4m5ex), judged on wall as parallel_files is, since more threads
// spend more user time by design. TestPerformanceExecuteProgramsAreTheEverydaySuite holds the list to suites.json, so the
// two never drift. tokenizer and trees are the compile half's too, so the binaries its warm builds make are the ones
// this half runs.
var executeSources = []executeSource{
	{name: "devirtualize", path: "bench/devirtualize.a", judge: "user"},
	{name: "nbody", path: "bench/nbody.ts", judge: "user"},
	{name: "parallel_files", path: "bench/parallel_files.a", judge: "wall"},
	{name: "primes_parallel", path: "bench/primes_parallel.a", judge: "wall"},
	{name: "sort", path: "bench/sort.ts", judge: "user"},
	{name: "spectral_norm", path: "bench/spectral_norm.ts", judge: "user"},
	{name: "tokenizer", path: "bench/tokenizer.ts", judge: "user"},
	{name: "trees", path: "bench/trees.ts", judge: "user"},
	{name: "word_count", path: "bench/word_count.ts", judge: "user"},
}

// The two runtimes every execute program runs on, by their place in a round's order.
const (
	nativeSide = iota
	nodeSide
)

var sideNames = [...]string{"native", "node"}

// A timer runs one command to its end and measures it, as benchmark.Run does. The measure takes one so a test can run
// it whole with no program, no Node and no reaper.
type timer func(command []string) benchmark.Measurement

// reaperTimer runs every command through reaper, benchmark.BuildReaper's binary or empty where there is none, with
// environment as its whole environment and bench's timeout.
func reaperTimer(reaper string, environment []string) timer {
	return func(command []string) benchmark.Measurement {
		return benchmark.Run(reaper, command, executeTimeout, environment)
	}
}

// executeEnvironment is every execute run's environment: this process's as bench keeps it (benchmark.Environment),
// without the settings the measure gives its builds, and without ADAMIC_THREADS, so a native run gets the pool's
// default threads, as bench's do unless -threads names others.
func executeEnvironment() []string {
	return benchmark.Environment(append(slices.Clone(measureSettings), "ADAMIC_THREADS")...)
}

// An executeRequest is where the execute half reads and writes: the compile half's request, whose root every source is
// read from and whose work holds every binary and build log, and how many rounds it runs.
type executeRequest struct {
	compile compileRequest
	rounds  int
}

// An executeTarget is a source as the execute half runs it: its absolute path, the binary built from it, and whether
// it imports 'adamic', which Node resolves only through the oracle runner.
type executeTarget struct {
	source        executeSource
	path          string
	binary        string
	importsAdamic bool
}

// measureExecute measures the execute half on the work directory the compile half measured in, after it, and says how
// many runtime archives its builds made. A program the compile half builds runs the binary its warm builds left; every
// other one is built once, untimed, on the same caches (buildForExecute). Then rounds rounds, each running every program
// natively and then on Node, every run through measure, and on odd rounds, counting from 0, both orders mirrored
// (executeOrder). Each program runs as bench runs it: from the repository's root, with no arguments, and Node through
// benchmark.NodeCommand. Every run's output must be Node's first run's, or the measure is an error: a native time for
// another answer is no time at all.
func measureExecute(request executeRequest, build runner, measure timer) (execute, int, error) {
	if err := os.MkdirAll(filepath.Join(request.compile.work, "logs"), 0o755); err != nil {
		return execute{}, 0, err
	}
	targets := []executeTarget{}
	archivesBuilt := 0
	for _, source := range executeSources {
		path := filepath.Join(request.compile.root, filepath.FromSlash(source.path))
		contents, err := os.ReadFile(path)
		if err != nil {
			return execute{}, 0, err
		}
		target := executeTarget{source: source, path: path, binary: request.compile.output(source.name), importsAdamic: benchmark.ImportsAdamic(string(contents))}
		compiled := slices.IndexFunc(sources, func(built programSource) bool { return built.path == source.path })
		if compiled >= 0 {
			target.binary = request.compile.output(sources[compiled].name)
		} else {
			built, err := buildForExecute(request.compile, target, build)
			if err != nil {
				return execute{}, 0, err
			}
			if built {
				archivesBuilt++
			}
		}
		targets = append(targets, target)
	}

	// runs[program][side] holds every round's measurement, in round order.
	runs := make([][2][]benchmark.Measurement, len(targets))
	for round := range request.rounds {
		programOrder, sideOrder := executeOrder(len(targets), round)
		for _, index := range programOrder {
			target := targets[index]
			commands := [2][]string{{target.binary}, benchmark.NodeCommand(target.path, target.importsAdamic, request.compile.root)}
			for _, side := range sideOrder {
				result := measure(commands[side])
				if result.Failure != "" {
					return execute{}, 0, fmt.Errorf("%s: %s's run in round %d %s", target.source.name, sideNames[side], round+1, result.Failure)
				}
				runs[index][side] = append(runs[index][side], result)
			}
			answer := runs[index][nodeSide][0].Stdout
			for _, side := range sideOrder {
				if printed := runs[index][side][round].Stdout; printed != answer {
					return execute{}, 0, fmt.Errorf("%s: %s's run in round %d printed another answer than Node's in round 1, at %s", target.source.name, sideNames[side], round+1, firstDifference(printed, answer))
				}
			}
		}
	}

	measured := execute{Rounds: request.rounds, Programs: executions{}}
	for index, target := range targets {
		entry := execution{Judge: target.source.judge, Native: runtimeNumbersOf(runs[index][nativeSide]), Node: runtimeNumbersOf(runs[index][nodeSide])}
		node := entry.Node.judged(entry.Judge).Median
		if node <= 0 {
			return execute{}, 0, fmt.Errorf("%s: Node's %s time read %d ms, so there is no ratio to it", target.source.name, entry.Judge, node)
		}
		entry.Ratio = ratioOf(entry.Native.judged(entry.Judge).Median, node)
		measured.Programs = append(measured.Programs, namedExecution{Name: target.source.name, Execution: entry})
	}
	return measured, archivesBuilt, nil
}

// buildForExecute builds target's binary once, untimed, on the compile half's caches, and says whether its runtime
// archive was built rather than found. A program's runtime features key its archive (native_runtime.FeatureFlags), so
// one whose features differ from hello's misses the archive hello's cold build made and builds its own. A miss costs
// no timed number, since no execute run is a build, so it is counted and never refused; a build whose log has no
// cArchive line at all never reached the archive, and is an error.
func buildForExecute(request compileRequest, target executeTarget, build runner) (bool, error) {
	logPath := filepath.Join(request.work, "logs", target.source.name+"-execute.log")
	if _, err := build(child{
		directory:   filepath.Dir(target.path),
		arguments:   []string{"build", filepath.Base(target.path), "-o", target.binary},
		environment: append(request.cacheSettings(), "ADAMIC_BUILD_LOG="+logPath),
	}); err != nil {
		return false, fmt.Errorf("%s: %w", target.source.name, err)
	}
	text, err := os.ReadFile(logPath)
	if err != nil {
		return false, fmt.Errorf("%s: its build's log: %w", target.source.name, err)
	}
	log, err := parseBuildLog(string(text))
	if err != nil {
		return false, fmt.Errorf("%s: its build's log: %w", target.source.name, err)
	}
	if log.archives == 0 {
		return false, fmt.Errorf("%s: its build logged no cArchive line, so it never reached the runtime archive", target.source.name)
	}
	return !log.archiveHit, nil
}

// executeOrder is a round's order: the programs by their index in executeSources, and the two sides, native before
// Node. On odd rounds, counting from 0, both are mirrored, as bench -compare swaps its sides on odd samples
// (sampleRuns), so a program's place in the round, and which runtime runs first, even out over the rounds.
func executeOrder(count, round int) ([]int, [2]int) {
	programOrder := make([]int, count)
	for index := range programOrder {
		programOrder[index] = index
	}
	sideOrder := [2]int{nativeSide, nodeSide}
	if round%2 == 1 {
		slices.Reverse(programOrder)
		sideOrder = [2]int{nodeSide, nativeSide}
	}
	return programOrder, sideOrder
}

// runtimeNumbersOf is one runtime's runs of a program as the file writes them: wall and user time in whole
// milliseconds, and peak resident set in whole KiB, each the median of the rounds (spreadOf).
func runtimeNumbersOf(runs []benchmark.Measurement) runtimeNumbers {
	walls, users, peaks := []int64{}, []int64{}, []int64{}
	for _, run := range runs {
		walls = append(walls, run.Elapsed.Milliseconds())
		users = append(users, run.User.Milliseconds())
		peaks = append(peaks, run.Peak/1024)
	}
	return runtimeNumbers{Wall: spreadOf(walls), User: spreadOf(users), PeakKiB: spreadOf(peaks).Median}
}

// ratioOf is native's judged median over Node's, rounded to two decimals, below 1 meaning native is faster. It's read
// from the medians the file writes, so a reader can check it from them. node is never 0: measureExecute refuses that.
func ratioOf(native, node int64) float64 {
	return math.Round(float64(native)/float64(node)*100) / 100
}

// firstDifference is where two outputs part: the first line that differs, each side's, or their lengths when every
// line is the same and only an end differs. It names what differs without printing either output whole.
func firstDifference(printed, answer string) string {
	printedLines, answerLines := strings.Split(printed, "\n"), strings.Split(answer, "\n")
	for index := range min(len(printedLines), len(answerLines)) {
		if printedLines[index] != answerLines[index] {
			return fmt.Sprintf("line %d, %q against %q", index+1, printedLines[index], answerLines[index])
		}
	}
	return fmt.Sprintf("its end, %d bytes against %d", len(printed), len(answer))
}
