// Command bench times Adamic's native binaries against Node and Bun on the same programs.
//
//	go run ./bench [-rounds 5] [-timeout 120s] [-only nbody,trees]
//
// Every program in bench/ is a fair one: the same source, as it stands, runs on all three, and all
// three must print the same answer or the row says so. Each is built natively the way a user
// builds it (adamic build: clang -O2, no sanitizers), and once more counted (adamic build --count),
// run once untimed, so every row carries the allocations and the retain and release traffic that
// explain it. The runs are interleaved, a round at a time (native, Node, Bun, native, ...), so a
// machine that slows down part way through slows all three alike, and the best of the rounds is
// reported, the run least disturbed by everything else on the machine. Peak memory is the
// process's maximum resident set, from the operating system. The machine and its load, before and
// after, are printed with the numbers, because numbers without them aren't worth reading.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// A runtime is one way to run a program.
type runtimeKind struct {
	name    string
	command func(program program) []string
}

// A program is one benchmark: its source, and the native binaries built from it.
type program struct {
	name    string
	source  string
	native  string
	counted string
}

// A measurement is one run's wall time and peak resident memory, or why there isn't one.
type measurement struct {
	elapsed time.Duration
	peak    int64 // bytes
	stdout  string
	failure string
}

var countsLine = regexp.MustCompile(`adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)`)

func main() {
	rounds := flag.Int("rounds", 5, "how many interleaved rounds; the best of them is reported")
	timeout := flag.Duration("timeout", 120*time.Second, "how long one run may take before it's reported as not finishing")
	only := flag.String("only", "", "comma-separated benchmark names to run (default: all)")
	flag.Parse()

	directory, err := benchDirectory()
	if err != nil {
		fail(err)
	}
	work, err := os.MkdirTemp("", "adamic-bench-")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(work)

	programs, err := findPrograms(directory, *only)
	if err != nil {
		fail(err)
	}
	runtimes := availableRuntimes()

	describeMachine(runtimes)
	fmt.Printf("load before: %s\n\n", loadAverage())

	adamic := filepath.Join(work, "adamic")
	if output, err := exec.Command("go", "build", "-o", adamic, "./cmd/adamic").CombinedOutput(); err != nil {
		fail(fmt.Errorf("building adamic: %v\n%s", err, output))
	}
	for index := range programs {
		program := &programs[index]
		program.native = filepath.Join(work, program.name)
		program.counted = filepath.Join(work, program.name+".counted")
		for _, build := range [][]string{{"build", program.source, "-o", program.native}, {"build", program.source, "-o", program.counted, "--count"}} {
			if output, err := exec.Command(adamic, build...).CombinedOutput(); err != nil {
				fail(fmt.Errorf("adamic %s: %v\n%s", strings.Join(build, " "), err, output))
			}
		}
	}

	// results[program][runtime] holds every round's measurement.
	results := make([][][]measurement, len(programs))
	for index := range results {
		results[index] = make([][]measurement, len(runtimes))
	}
	for round := 0; round < *rounds; round++ {
		for programIndex, program := range programs {
			for runtimeIndex, kind := range runtimes {
				previous := results[programIndex][runtimeIndex]
				// A runtime that didn't finish once won't in later rounds either; don't spend them.
				if len(previous) > 0 && previous[0].failure != "" {
					continue
				}
				result := run(kind.command(program), *timeout)
				results[programIndex][runtimeIndex] = append(previous, result)
				fmt.Fprintf(os.Stderr, "round %d %s %s: %s\n", round+1, program.name, kind.name, short(result))
			}
		}
	}

	fmt.Printf("best of %d interleaved rounds; time is wall clock, memory is peak resident\n\n", *rounds)
	header := []string{"benchmark"}
	for _, kind := range runtimes {
		header = append(header, kind.name+" time", kind.name+" memory")
	}
	header = append(header, "native vs node", "same answer")
	fmt.Println("| " + strings.Join(header, " | ") + " |")
	fmt.Println("|" + strings.Repeat("---|", len(header)))
	for programIndex, program := range programs {
		row := []string{program.name}
		best := make([]measurement, len(runtimes))
		for runtimeIndex := range runtimes {
			best[runtimeIndex] = bestOf(results[programIndex][runtimeIndex])
			if best[runtimeIndex].failure != "" {
				row = append(row, best[runtimeIndex].failure, "")
				continue
			}
			row = append(row, seconds(best[runtimeIndex].elapsed), megabytes(best[runtimeIndex].peak))
		}
		row = append(row, ratio(best[0], best[1]), sameAnswer(best))
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

	fmt.Println()
	fmt.Println("counted native build, run once (adamic build --count): heap values and reference-count calls")
	fmt.Println()
	fmt.Println("| benchmark | allocations | frees | retains | releases | peak live | in regions |")
	fmt.Println("|---|---:|---:|---:|---:|---:|---:|")
	for _, program := range programs {
		fmt.Println(counts(program, *timeout))
	}
	fmt.Printf("\nload after: %s\n", loadAverage())
}

// benchDirectory is bench/ in the repository, from wherever go run was started inside it.
func benchDirectory() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("bench runs inside the adamic repository: %w", err)
	}
	root := strings.TrimSpace(string(output))
	if err := os.Chdir(root); err != nil {
		return "", err
	}
	return filepath.Join(root, "bench"), nil
}

func findPrograms(directory string, only string) ([]program, error) {
	sources, err := filepath.Glob(filepath.Join(directory, "*.ts"))
	if err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, name := range strings.Split(only, ",") {
		if name != "" {
			wanted[name] = true
		}
	}
	programs := []program{}
	for _, source := range sources {
		name := strings.TrimSuffix(filepath.Base(source), ".ts")
		if len(wanted) == 0 || wanted[name] {
			programs = append(programs, program{name: name, source: source})
		}
	}
	if len(programs) == 0 {
		return nil, errors.New("no benchmarks to run")
	}
	return programs, nil
}

// availableRuntimes is native and Node always, and Bun when it's installed.
func availableRuntimes() []runtimeKind {
	runtimes := []runtimeKind{
		{"native", func(program program) []string { return []string{program.native} }},
		{"node", func(program program) []string { return []string{"node", program.source} }},
	}
	if _, err := exec.LookPath("bun"); err == nil {
		runtimes = append(runtimes, runtimeKind{"bun", func(program program) []string { return []string{"bun", program.source} }})
	}
	return runtimes
}

// run runs a command to the end, and measures it.
func run(command []string, timeout time.Duration) measurement {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	process := exec.CommandContext(ctx, command[0], command[1:]...)
	process.Env = environment()
	var stdout, stderr bytes.Buffer
	process.Stdout = &stdout
	process.Stderr = &stderr
	start := time.Now()
	err := process.Run()
	elapsed := time.Since(start)
	if ctx.Err() != nil {
		return measurement{failure: "did not finish in " + timeout.String()}
	}
	if err != nil {
		return measurement{failure: fmt.Sprintf("failed: %v %s", err, firstLine(stderr.String()))}
	}
	return measurement{elapsed: elapsed, peak: peakResident(process.ProcessState), stdout: stdout.String()}
}

// environment is this process's, without what would change how a runtime runs: BUN_OPTIONS
// (a shell that sets --smol makes Bun trade speed for a smaller heap) and NODE_OPTIONS.
func environment() []string {
	kept := []string{}
	for _, variable := range os.Environ() {
		if strings.HasPrefix(variable, "BUN_OPTIONS=") || strings.HasPrefix(variable, "NODE_OPTIONS=") {
			continue
		}
		kept = append(kept, variable)
	}
	return kept
}

// peakResident is the process's maximum resident set in bytes: Linux reports it in kilobytes, and
// macOS in bytes.
func peakResident(state *os.ProcessState) int64 {
	usage, isUsage := state.SysUsage().(*syscall.Rusage)
	if !isUsage {
		return 0
	}
	if runtime.GOOS == "darwin" {
		return usage.Maxrss
	}
	return usage.Maxrss * 1024
}

// bestOf is the fastest finished run, or the failure when none finished.
func bestOf(runs []measurement) measurement {
	finished := []measurement{}
	for _, run := range runs {
		if run.failure == "" {
			finished = append(finished, run)
		}
	}
	if len(finished) == 0 {
		if len(runs) == 0 {
			return measurement{failure: "not run"}
		}
		return runs[0]
	}
	return slices.MinFunc(finished, func(left, right measurement) int { return int(left.elapsed - right.elapsed) })
}

// counts runs the counted build once and reads the line it writes last.
func counts(program program, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	process := exec.CommandContext(ctx, program.counted)
	var stderr bytes.Buffer
	process.Stderr = &stderr
	process.Stdout = nil
	_ = process.Run()
	if ctx.Err() != nil {
		return fmt.Sprintf("| %s | did not finish in %s | | | | | |", program.name, timeout)
	}
	match := countsLine.FindStringSubmatch(stderr.String())
	if match == nil {
		return fmt.Sprintf("| %s | no counts written | | | | | |", program.name)
	}
	cells := []string{program.name}
	for _, number := range match[1:] {
		value, _ := strconv.ParseInt(number, 10, 64)
		cells = append(cells, thousands(value))
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

func sameAnswer(best []measurement) string {
	answer := ""
	for _, result := range best {
		if result.failure != "" {
			continue
		}
		if answer == "" {
			answer = result.stdout
		} else if result.stdout != answer {
			return "DIFFERS"
		}
	}
	return "yes"
}

// ratio says how native compares with Node: below 1 is native faster.
func ratio(native measurement, node measurement) string {
	if native.failure != "" || node.failure != "" {
		return "n/a"
	}
	return fmt.Sprintf("%.2fx", native.elapsed.Seconds()/node.elapsed.Seconds())
}

func describeMachine(runtimes []runtimeKind) {
	fmt.Printf("date: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("machine: %s/%s, %d logical CPUs, %s\n", runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), cpuModel())
	if uname, err := exec.Command("uname", "-srm").Output(); err == nil {
		fmt.Printf("kernel: %s", uname)
	}
	if commit, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		fmt.Printf("adamic: %s", commit)
	}
	for _, version := range [][]string{{"clang", "--version"}, {"node", "--version"}, {"bun", "--version"}} {
		if output, err := exec.Command(version[0], version[1:]...).Output(); err == nil {
			fmt.Printf("%s: %s\n", version[0], firstLine(string(output)))
		}
	}
	names := []string{}
	for _, kind := range runtimes {
		names = append(names, kind.name)
	}
	fmt.Printf("runtimes: %s\n", strings.Join(names, ", "))
}

func cpuModel() string {
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	if contents, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(contents), "\n") {
			if name, found := strings.CutPrefix(line, "model name"); found {
				return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(name), ":"))
			}
		}
	}
	return "unknown processor"
}

func loadAverage() string {
	if contents, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(contents))
		if len(fields) >= 3 {
			return strings.Join(fields[:3], " ") + " (1, 5, 15 minutes)"
		}
	}
	if output, err := exec.Command("sysctl", "-n", "vm.loadavg").Output(); err == nil {
		return strings.Trim(strings.TrimSpace(string(output)), "{} ") + " (1, 5, 15 minutes)"
	}
	return "unknown"
}

func seconds(elapsed time.Duration) string {
	return fmt.Sprintf("%.3f s", elapsed.Seconds())
}

func megabytes(bytes int64) string {
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
}

func thousands(value int64) string {
	text := strconv.FormatInt(value, 10)
	for index := len(text) - 3; index > 0; index -= 3 {
		text = text[:index] + "," + text[index:]
	}
	return text
}

func short(result measurement) string {
	if result.failure != "" {
		return result.failure
	}
	return seconds(result.elapsed) + ", " + megabytes(result.peak)
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bench:", err)
	os.Exit(1)
}
