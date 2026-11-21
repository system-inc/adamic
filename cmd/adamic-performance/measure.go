package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// phaseNames are the compiler's phases a build log times, in the order they run.
var phaseNames = []string{"load", "lower", "emit"}

// generatedPath is the path written for a program this command writes itself rather than reads from the tree.
const generatedPath = "<generated>"

// helloSource is hello, the floor under every build: process start, the front end and the runtime's link, with
// almost no program of its own.
const helloSource = "console.log(\"hello, world\");\n"

// A programSource is one program the compile half builds, by its name in the file and its path from the repository's
// root, empty for hello, which is written into the temporary directory.
type programSource struct {
	name string
	path string
}

// sources are the compile half's programs, in the file's order. Parser.a is real code from the TypeScript compiler,
// and tokenizer and trees are the execute half's too, so the binaries their warm builds make are the ones it runs.
var sources = []programSource{
	{name: "hello"},
	{name: "Parser", path: "stage3/namespaces/shapes/Parser.a"},
	{name: "tokenizer", path: "bench/tokenizer.ts"},
	{name: "trees", path: "bench/trees.ts"},
}

// A child is one run of the adamic binary: in directory, as cd <directory> would, with arguments, and environment's
// settings over the measuring process's own. keepOutput has its standard output returned, which is otherwise
// discarded, so a front end's time isn't spent filling a buffer.
type child struct {
	directory   string
	arguments   []string
	environment []string
	keepOutput  bool
}

// A childResult is a child that exited 0: its wall time, clocked around the process, and its standard output when
// kept.
type childResult struct {
	elapsed time.Duration
	stdout  []byte
}

// A runner runs one child to its end. An exit other than 0 is its error, with the first lines the child printed. The
// measure takes one so a test can run it whole with no adamic and no compiler.
type runner func(child child) (childResult, error)

// A compileRequest is where the compile half reads and writes: root is the repository's, which every source path is
// read from, and work the temporary directory, which holds hello.a, every output, every build log and both caches.
type compileRequest struct {
	root   string
	work   string
	rounds int
}

// cacheSettings put both of the build's caches in work, so the cold build is cold and nothing reaches this machine's.
func (request compileRequest) cacheSettings() []string {
	return []string{
		"XDG_CACHE_HOME=" + filepath.Join(request.work, "cache"),
		"ADAMIC_BUILD_CACHE_DIR=" + filepath.Join(request.work, "build-cache"),
	}
}

// output is where a program's build writes its binary, the same path every round.
func (request compileRequest) output(name string) string {
	return filepath.Join(request.work, "out", name)
}

// A buildTarget is a source as the measure builds it: the directory its build runs in, the base name it builds there,
// and how many lines the source holds.
type buildTarget struct {
	source    programSource
	directory string
	base      string
	lines     int
}

// measureCompile measures the compile half. One build of hello comes first with every cache empty, timed as its cold
// build; one untimed build of every other program then caches its own objects. Then rounds rounds, each building every
// program once in order, and rounds more, each running adamic c on every one: interleaved, so a machine that slows part
// way through slows every program alike. A warm build's own build log gives its C compile and link, and one with no
// cArchive hit is an error, since its cache didn't hold and its time would be a cold one's.
func measureCompile(request compileRequest, run runner) (compile, error) {
	for _, directory := range []string{"out", "logs"} {
		if err := os.MkdirAll(filepath.Join(request.work, directory), 0o755); err != nil {
			return compile{}, err
		}
	}
	targets := []buildTarget{}
	for _, source := range sources {
		path := filepath.Join(request.root, filepath.FromSlash(source.path))
		if source.path == "" {
			path = filepath.Join(request.work, source.name+".a")
			if err := os.WriteFile(path, []byte(helloSource), 0o644); err != nil {
				return compile{}, err
			}
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return compile{}, err
		}
		targets = append(targets, buildTarget{source: source, directory: filepath.Dir(path), base: filepath.Base(path), lines: countLines(contents)})
	}
	build := func(target buildTarget, settings ...string) (childResult, error) {
		result, err := run(child{
			directory:   target.directory,
			arguments:   []string{"build", target.base, "-o", request.output(target.source.name)},
			environment: append(request.cacheSettings(), settings...),
		})
		if err != nil {
			return result, fmt.Errorf("%s: %w", target.source.name, err)
		}
		return result, nil
	}

	cold, err := build(targets[0])
	if err != nil {
		return compile{}, err
	}
	for _, target := range targets[1:] {
		if _, err := build(target); err != nil {
			return compile{}, err
		}
	}

	warm := make([][]int64, len(targets))
	compiles := make([][]int64, len(targets))
	links := make([][]int64, len(targets))
	phases := make([]map[string][]int64, len(targets))
	for index := range phases {
		phases[index] = map[string][]int64{}
	}
	for round := 1; round <= request.rounds; round++ {
		for index, target := range targets {
			logPath := filepath.Join(request.work, "logs", fmt.Sprintf("%s-%d.log", target.source.name, round))
			result, err := build(target, "ADAMIC_BUILD_LOG="+logPath)
			if err != nil {
				return compile{}, err
			}
			text, err := os.ReadFile(logPath)
			if err != nil {
				return compile{}, fmt.Errorf("%s: warm build %d's log: %w", target.source.name, round, err)
			}
			log, err := parseBuildLog(string(text))
			if err != nil {
				return compile{}, fmt.Errorf("%s: warm build %d's log: %w", target.source.name, round, err)
			}
			if !log.archiveHit {
				return compile{}, fmt.Errorf("%s: warm build %d logged no cArchive hit: the runtime archive's cache didn't hold, so its time would be a cold build's", target.source.name, round)
			}
			if log.compiles == 0 || log.links == 0 {
				return compile{}, fmt.Errorf("%s: warm build %d logged %d cCompile and %d cLink lines, and its C compile and link need one of each", target.source.name, round, log.compiles, log.links)
			}
			for _, name := range phaseNames {
				seconds, found := log.phases[name]
				if !found {
					return compile{}, fmt.Errorf("%s: warm build %d logged no %s phase", target.source.name, round, name)
				}
				phases[index][name] = append(phases[index][name], milliseconds(seconds))
			}
			warm[index] = append(warm[index], result.elapsed.Milliseconds())
			compiles[index] = append(compiles[index], milliseconds(log.compile))
			links[index] = append(links[index], milliseconds(log.link))
		}
	}

	frontEnd := make([][]int64, len(targets))
	cLines := make([]int, len(targets))
	for round := 1; round <= request.rounds; round++ {
		for index, target := range targets {
			// The first round's C is kept to count, and every other round's discarded.
			result, err := run(child{directory: target.directory, arguments: []string{"c", target.base}, environment: request.cacheSettings(), keepOutput: round == 1})
			if err != nil {
				return compile{}, fmt.Errorf("%s: %w", target.source.name, err)
			}
			if round == 1 {
				cLines[index] = countLines(result.stdout)
			}
			frontEnd[index] = append(frontEnd[index], result.elapsed.Milliseconds())
		}
	}

	measured := compile{Rounds: request.rounds, Programs: programs{}}
	for index, target := range targets {
		path := target.source.path
		if path == "" {
			path = generatedPath
		}
		entry := program{
			Path:      path,
			Lines:     target.lines,
			CLines:    cLines[index],
			WarmBuild: spreadOf(warm[index]),
			FrontEnd:  spreadOf(frontEnd[index]),
			Clang:     clangTimes{Compile: spreadOf(compiles[index]).Median, Link: spreadOf(links[index]).Median},
			Phases:    phaseTimes{Load: spreadOf(phases[index]["load"]).Median, Lower: spreadOf(phases[index]["lower"]).Median, Emit: spreadOf(phases[index]["emit"]).Median},
		}
		if index == 0 {
			coldBuild := cold.elapsed.Milliseconds()
			entry.ColdBuild = &coldBuild
		}
		measured.Programs = append(measured.Programs, namedProgram{Name: target.source.name, Program: entry})
	}
	return measured, nil
}

// A buildLog is what one build's $ADAMIC_BUILD_LOG says: the seconds clang ran for its compiles and its links, summed,
// how many lines of each, how many cArchive lines and whether one was a cache hit, and the seconds of each phase.
type buildLog struct {
	compile  float64
	link     float64
	compiles int
	links    int
	// archives counts the cArchive lines, hit or miss: a build with none never reached the runtime archive.
	archives   int
	archiveHit bool
	phases     map[string]float64
}

// parseBuildLog reads a build log in internal/native_runtime's BuildLog grammar, '<kind> <name> <key12>
// hit|miss|off <seconds>', and internal/buildlog's phase lines, 'phase <name> - off <seconds>'. Every other line is
// another reader's, buildcache's census among them, and passes by; a line of a kind read here that doesn't parse is an
// error, never a zero.
func parseBuildLog(text string) (buildLog, error) {
	log := buildLog{phases: map[string]float64{}}
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		kind := fields[0]
		if kind != "cCompile" && kind != "cLink" && kind != "cArchive" && kind != "phase" {
			continue
		}
		if len(fields) != 5 {
			return buildLog{}, fmt.Errorf("the line %q isn't five fields", line)
		}
		seconds, err := strconv.ParseFloat(fields[4], 64)
		if err != nil || math.IsNaN(seconds) || seconds < 0 || math.IsInf(seconds, 0) {
			return buildLog{}, fmt.Errorf("the line %q ends in %q, which isn't seconds", line, fields[4])
		}
		switch kind {
		case "cCompile":
			log.compile += seconds
			log.compiles++
		case "cLink":
			log.link += seconds
			log.links++
		case "cArchive":
			log.archives++
			if fields[3] == "hit" {
				log.archiveHit = true
			}
		case "phase":
			log.phases[fields[1]] += seconds
		}
	}
	return log, nil
}

// milliseconds is a build log's seconds, written to the millisecond, in whole milliseconds.
func milliseconds(seconds float64) int64 {
	return int64(math.Round(seconds * 1000))
}

// spreadOf is the median of values with their min and max. The median of an even count is the lower of the middle
// two, so it is always a round that was measured, never an average no round took. values is never empty: rounds is at
// least 1.
func spreadOf(values []int64) spread {
	sorted := slices.Sorted(slices.Values(values))
	return spread{Median: sorted[(len(sorted)-1)/2], Min: sorted[0], Max: sorted[len(sorted)-1]}
}

// countLines is how many lines text holds, a last one with no newline counted.
func countLines(text []byte) int {
	lines := bytes.Count(text, []byte("\n"))
	if len(text) > 0 && text[len(text)-1] != '\n' {
		lines++
	}
	return lines
}
