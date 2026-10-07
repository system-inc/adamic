// adamic-steer rewrites the breadth scene's weight table (internal/fuzz/breadth_weights.txt) from
// coverage. Each round generates a batch of programs with the current weights, runs each the way
// adamic-fuzz does with the compiler built for Go coverage and the runtime for clang's, and reads
// what each program reached on its own. Taken in seed order, a program's novelty is what it reached
// that nothing before it had: neither the baseline (a measure.sh run of the generator) nor an earlier
// program. A region the oracle fixtures don't reach either counts four times one they do, so the
// weights steer first toward what neither line of defense executes. Every construct and variant a
// program chose shares its novelty; after the round each one that took part gains weight if its
// programs lit anything new and loses weight if they only revisited, and the weights are scaled back
// to a mean of 1.
//
//	adamic-steer -baseline /tmp/adamic-coverage -seed 100001 -batch 120 -rounds 4
//
// Same baseline, seed, batch, rounds and starting table, same table out: nothing here depends on
// timing, only on what the programs reached, which they reach the same way every run.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/system-inc/adamic/internal/fuzz"
	"github.com/system-inc/adamic/internal/native"
)

// coverpkg is what measure.sh measures, with cmd/adamic so the toolchain links the coverage writer.
const coverpkg = "./cmd/adamic,./internal/lower/...,./internal/native/...,./internal/ir/...,./internal/flow/..."

// neitherWeight is what a region neither the generator's baseline nor the fixtures reach is worth,
// against 1 for one only the fixtures reach.
const neitherWeight = 4

func main() {
	os.Exit(run())
}

func run() int {
	root := flag.String("root", ".", "the checkout whose compiler and runtime are measured")
	baseline := flag.String("baseline", "", "a verify/coverage/measure.sh output directory: generator/go.txt, generator/c.profdata, fixtures/go.txt, fixtures/c.profdata")
	seed := flag.Uint64("seed", 100001, "the first program's seed; keep it clear of the seeds coverage is measured on")
	batch := flag.Int("batch", 120, "programs per round")
	rounds := flag.Int("rounds", 4, "rounds")
	parallel := flag.Int("parallel", 6, "programs at once")
	weightsPath := flag.String("weights", "internal/fuzz/breadth_weights.txt", "the starting table")
	output := flag.String("o", "", "where the new table goes (default: over -weights)")
	work := flag.String("work", filepath.Join(os.TempDir(), "adamic-steer"), "where programs are built and run")
	raise := flag.Float64("raise", 1.5, "a weight's factor when its programs lit something new")
	lower := flag.Float64("lower", 0.7, "a weight's factor when its programs only revisited")
	flag.Parse()
	if *baseline == "" {
		fmt.Fprintln(os.Stderr, "adamic-steer: -baseline is required: a measure.sh output directory")
		return 2
	}
	if *output == "" {
		*output = *weightsPath
	}
	// The fuzz package reads its clang flags as it starts, so the coverage build has to be asked for
	// in the environment this process starts with.
	if !native.CoverageRequested() {
		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		environment := append(os.Environ(), "ADAMIC_C_COVERAGE=1", "ADAMIC_COVERAGE_PER_PROGRAM=1", "GOFLAGS=-trimpath -cover -coverpkg="+coverpkg)
		if err := syscall.Exec(executable, os.Args, environment); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	text, err := os.ReadFile(*weightsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	weights, err := fuzz.ParseWeights(string(text))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, name := range fuzz.BreadthChoices() {
		if _, known := weights[name]; !known {
			weights[name] = 1
		}
	}

	if err := os.MkdirAll(*work, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	checkout, err := fuzz.Prepare(*root, filepath.Join(*work, "checkout"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	library, err := native.RuntimeLibrary(filepath.Join(checkout.Root, "internal", "native", "runtime"), native.Options{Sanitize: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	runtimeNames, err := runtimeFunctions(filepath.Dir(library))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// What's covered already, and what the fixtures reach, which sets what a new region is worth.
	covered, err := baselineKeys(*baseline, "generator", runtimeNames)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fixtures, err := baselineKeys(*baseline, "fixtures", runtimeNames)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("baseline: %d keys the generator reaches, %d the fixtures do\n", len(covered), len(fixtures))

	for round := range *rounds {
		first := *seed + uint64(round**batch)
		reached := make([]map[string]bool, *batch)
		chosen := make([][]string, *batch)
		var next sync.Mutex
		index := 0
		var group sync.WaitGroup
		for range *parallel {
			group.Add(1)
			go func() {
				defer group.Done()
				for {
					next.Lock()
					mine := index
					index++
					next.Unlock()
					if mine >= *batch {
						return
					}
					programSeed := first + uint64(mine)
					program, choices := fuzz.GenerateWeighted(programSeed, nil, nil, weights)
					chosen[mine] = choices
					directory := filepath.Join(*work, "programs", strconv.FormatUint(programSeed, 10))
					_ = os.RemoveAll(directory)
					checkout.Try(program.Source(), directory)
					keys, err := programKeys(filepath.Join(directory, fuzz.ProgramCoverage), runtimeNames)
					if err != nil {
						fmt.Fprintf(os.Stderr, "seed %d: %v\n", programSeed, err)
					}
					reached[mine] = keys
					_ = os.RemoveAll(directory)
				}
			}()
		}
		group.Wait()

		gain := map[string]float64{}
		uses := map[string]int{}
		var newNeither, newFixtures int
		for program := range *batch {
			score := 0.0
			for _, key := range sortedKeys(reached[program]) {
				if covered[key] {
					continue
				}
				covered[key] = true
				if fixtures[key] {
					score++
					newFixtures++
				} else {
					score += neitherWeight
					newNeither++
				}
			}
			for _, name := range chosen[program] {
				gain[name] += score
				uses[name]++
			}
		}
		for name := range uses {
			if gain[name] > 0 {
				weights[name] *= *raise
			} else {
				weights[name] *= *lower
			}
		}
		normalize(weights)
		fmt.Printf("round %d, seeds %d to %d: %d new keys neither line reached, %d new the fixtures reached\n", round+1, first, first+uint64(*batch)-1, newNeither, newFixtures)
		names := fuzz.BreadthChoices()
		sort.SliceStable(names, func(left, right int) bool { return gain[names[left]] > gain[names[right]] })
		for _, name := range names[:min(8, len(names))] {
			fmt.Printf("  %-28s gain %6.0f over %3d programs, weight %.3f\n", name, gain[name], uses[name], weights[name])
		}
	}

	header := fmt.Sprintf("The breadth scene's weights (internal/fuzz/breadth.go), written by cmd/adamic-steer:\n-baseline %s -seed %d -batch %d -rounds %d -raise %g -lower %g, from %s.\nA construct's weight picks it among constructs; construct/variant picks the variant within it.",
		filepath.Base(*baseline), *seed, *batch, *rounds, *raise, *lower, filepath.Base(*weightsPath))
	if err := os.WriteFile(*output, []byte(fuzz.FormatWeights(weights, header)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("wrote %s\n", *output)
	return 0
}

// normalize scales the constructs' weights to a mean of 1, and each construct's variants to a mean
// of 1 among themselves, so a round that lowers everything changes nothing but the ratios.
func normalize(weights map[string]float64) {
	groups := map[string][]string{}
	for _, name := range fuzz.BreadthChoices() {
		group := ""
		if slash := strings.IndexByte(name, '/'); slash >= 0 {
			group = name[:slash]
		}
		groups[group] = append(groups[group], name)
	}
	for _, names := range groups {
		total := 0.0
		for _, name := range names {
			total += weights[name]
		}
		mean := total / float64(len(names))
		for _, name := range names {
			weights[name] = math.Max(0.02, math.Min(50, weights[name]/mean))
		}
	}
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// runtimeFunctions are the names the runtime's functions have in a clang profile: an external
// function by its own name, a static one as file.c:name. A program's own functions (main.c) aren't
// among them.
func runtimeFunctions(directory string) (map[string]bool, error) {
	objects, err := filepath.Glob(filepath.Join(directory, "*.o"))
	if err != nil || len(objects) == 0 {
		return nil, fmt.Errorf("adamic-steer: no runtime objects in %s", directory)
	}
	names := map[string]bool{}
	for _, object := range objects {
		output, err := exec.Command("nm", "-gU", object).Output()
		if err != nil {
			return nil, fmt.Errorf("adamic-steer: nm %s: %w", object, err)
		}
		for _, line := range strings.Split(string(output), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			name := fields[2]
			if runtime.GOOS == "darwin" {
				name = strings.TrimPrefix(name, "_")
			}
			names[name] = true
		}
		names["file:"+strings.TrimSuffix(filepath.Base(object), ".o")+".c"] = true
	}
	return names, nil
}

func isRuntime(function string, names map[string]bool) bool {
	if colon := strings.IndexByte(function, ':'); colon > 0 {
		return names["file:"+function[:colon]]
	}
	return names[function]
}

// baselineKeys is what a side of a measure.sh run reached: Go blocks as file:position, and runtime
// counters as function#hash#counter.
func baselineKeys(directory string, side string, runtimeNames map[string]bool) (map[string]bool, error) {
	keys := map[string]bool{}
	if err := goKeys(filepath.Join(directory, side, "go.txt"), keys); err != nil {
		return nil, err
	}
	if err := clangKeys([]string{filepath.Join(directory, side, "c.profdata")}, runtimeNames, keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// programKeys is what one program's runs reached, from its coverage directory.
func programKeys(directory string, runtimeNames map[string]bool) (map[string]bool, error) {
	keys := map[string]bool{}
	golang := filepath.Join(directory, "go")
	if entries, _ := os.ReadDir(golang); len(entries) > 0 {
		text := filepath.Join(directory, "go.txt")
		if output, err := exec.Command("go", "tool", "covdata", "textfmt", "-i="+golang, "-o="+text).CombinedOutput(); err != nil {
			return keys, fmt.Errorf("covdata: %w\n%s", err, output)
		}
		if err := goKeys(text, keys); err != nil {
			return keys, err
		}
	}
	profiles, _ := filepath.Glob(filepath.Join(directory, "c", "*.profraw"))
	if len(profiles) > 0 {
		if err := clangKeys(profiles, runtimeNames, keys); err != nil {
			return keys, err
		}
	}
	return keys, nil
}

// goKeys adds the blocks a Go text profile ran, in the packages measure.sh reads.
func goKeys(path string, keys map[string]bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		space := strings.LastIndexByte(line, ' ')
		if space < 0 || strings.HasPrefix(line, "mode:") || line[space+1:] == "0" {
			continue
		}
		block := line[:strings.LastIndexByte(line[:space], ' ')]
		if strings.Contains(block, "/internal/") && !strings.Contains(block, "/cmd/") {
			keys["go "+block] = true
		}
	}
	return scanner.Err()
}

// clangKeys adds the runtime counters that a set of clang profiles counted above zero.
func clangKeys(profiles []string, runtimeNames map[string]bool, keys map[string]bool) error {
	arguments := append([]string{"llvm-profdata", "merge", "-text", "-o", "-"}, profiles...)
	command := exec.Command(arguments[0], arguments[1:]...)
	if runtime.GOOS == "darwin" {
		command = exec.Command("xcrun", arguments...)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("llvm-profdata: %w\n%s", err, stderr.String())
	}
	lines := strings.Split(string(output), "\n")
	for index := 0; index < len(lines); index++ {
		if lines[index] != "# Func Hash:" || index == 0 || index+4 >= len(lines) {
			continue
		}
		function, hash := lines[index-1], lines[index+1]
		counters, err := strconv.Atoi(lines[index+3])
		if err != nil {
			continue
		}
		if isRuntime(function, runtimeNames) {
			for counter := range counters {
				if index+5+counter < len(lines) && lines[index+5+counter] != "0" {
					keys[fmt.Sprintf("c %s#%s#%d", function, hash, counter)] = true
				}
			}
		}
		index += 4 + counters
	}
	return nil
}
