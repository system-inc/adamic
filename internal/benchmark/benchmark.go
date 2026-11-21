// Package benchmark runs a program to measure it: its wall time from this process's clock, its user and system time
// and its peak resident set from the reaper on Linux (BuildReaper says why) or from the process's own usage elsewhere,
// and what it printed, kept small enough to hold over many runs. bench, its -game included, and cmd/adamic-performance
// both measure through it, so a number one prints is read the way the other's is.
package benchmark

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// reaperSource is reaper/reaper.c, in a directory of its own because a .c file beside this package's Go files would
// make it a cgo package, which the sealed jobs' Go refuses (J-gx6b2hv).
//
//go:embed reaper/reaper.c
var reaperSource []byte

// A Measurement is one run's wall time, user and system time and peak resident memory, or why there isn't one.
type Measurement struct {
	Elapsed time.Duration
	User    time.Duration // CPU time in user mode, summed over every thread
	System  time.Duration // CPU time in the kernel on its behalf, summed over every thread
	Peak    int64         // bytes
	// Stdout is what the run printed, as keptOutput keeps it.
	Stdout  string
	Failure string
}

// BuildReaper writes reaper/reaper.c into work and builds it there, on Linux, for Run to start every program through.
// Go starts a child with clone(CLONE_VFORK|CLONE_VM), and Linux's exec folds the old memory's high-water resident set
// into the new program's maximum (fs/exec.c, exec_mmap), so the measuring process's own resident set was a floor under
// every program's peak: a bench holding 250 MB of inputs read 1321.4 MB for every program, one in C that only prints
// among them (J-wme5jp2). Through the reaper, the old memory is the reaper's, a few hundred kilobytes. The fold is
// Linux's exec, so elsewhere it builds nothing and Run starts the program itself, its peak read by peakResident as
// always. The source is embedded, so a caller needs no checkout to find it: finding it through git would fail where
// the tree is no checkout, as Loom's units are (block 3183).
func BuildReaper(work string) (string, error) {
	if runtime.GOOS != "linux" {
		return "", nil
	}
	// The host's own triple, named as every compile at a runner names its target (#5xsd2ha).
	platform, err := native_runtime.ResolveTarget(native_runtime.Options{})
	if err != nil {
		return "", err
	}
	source := filepath.Join(work, "reaper.c")
	if err := os.WriteFile(source, reaperSource, 0o644); err != nil {
		return "", err
	}
	binary := filepath.Join(work, "reaper")
	if output, err := exec.Command("clang", "--target="+platform.Target, "-O2", "-o", binary, source).CombinedOutput(); err != nil {
		return "", fmt.Errorf("building internal/benchmark/reaper/reaper.c: %v\n%s", err, output)
	}
	return binary, nil
}

// Run runs a command to the end with environment as its whole environment, and measures it. input is a file on its
// standard input, as the Benchmarks Game hands a program its input (`< input.txt`), or empty for none. With a reaper,
// BuildReaper's binary, the command runs through it, and its user and system time and peak come from the line the
// reaper writes to descriptor 3; with none, the program is started directly and they come from its own usage. The wall
// time is always this process's own clock, and a timeout kills the reaper, which takes the program with it. A run that
// failed, didn't finish or printed nothing, or whose input can't be opened, is a Measurement whose Failure says so.
func Run(reaper string, command []string, input string, timeout time.Duration, environment []string) Measurement {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var stdin *os.File
	if input != "" {
		file, err := os.Open(input)
		if err != nil {
			return Measurement{Failure: fmt.Sprintf("its input can't be opened: %v", err)}
		}
		defer file.Close()
		stdin = file
	}
	var usageRead, usageWrite *os.File
	if reaper != "" {
		var err error
		usageRead, usageWrite, err = os.Pipe()
		if err != nil {
			return Measurement{Failure: fmt.Sprintf("failed: a pipe for the reaper: %v", err)}
		}
		defer usageRead.Close()
		command = append([]string{reaper, "3"}, command...)
	}
	process := exec.CommandContext(ctx, command[0], command[1:]...)
	process.Env = environment
	if stdin != nil {
		process.Stdin = stdin
	}
	var stdout, stderr bytes.Buffer
	process.Stdout = &stdout
	process.Stderr = &stderr
	if usageWrite != nil {
		process.ExtraFiles = []*os.File{usageWrite}
	}
	start := time.Now()
	err := process.Start()
	if usageWrite != nil {
		// The reaper has its own copy now; with this one closed, the pipe ends when the reaper does.
		usageWrite.Close()
	}
	if err == nil {
		err = process.Wait()
	}
	elapsed := time.Since(start)
	if ctx.Err() != nil {
		return Measurement{Failure: "did not finish in " + timeout.String()}
	}
	if err != nil {
		return Measurement{Failure: fmt.Sprintf("failed: %v %s", err, firstLine(stderr.String()))}
	}
	if strings.TrimSpace(stdout.String()) == "" {
		return Measurement{Failure: "printed nothing"}
	}
	if usageRead == nil {
		return Measurement{Elapsed: elapsed, User: process.ProcessState.UserTime(), System: process.ProcessState.SystemTime(), Peak: peakResident(process.ProcessState), Stdout: keptOutput(stdout.Bytes())}
	}
	line, err := io.ReadAll(usageRead)
	if err != nil {
		return Measurement{Failure: fmt.Sprintf("failed: reading the reaper's usage line: %v", err)}
	}
	user, system, peak, err := reaped(string(line))
	if err != nil {
		return Measurement{Failure: "failed: " + err.Error()}
	}
	return Measurement{Elapsed: elapsed, User: user, System: system, Peak: peak, Stdout: keptOutput(stdout.Bytes())}
}

// reaped reads the reaper's line, "<user microseconds> <system microseconds> <maxrss kilobytes>\n", as the program's
// user and system time and its peak resident set in bytes. A line missing or cut short is an error, never a zero.
func reaped(line string) (time.Duration, time.Duration, int64, error) {
	if line == "" {
		return 0, 0, 0, errors.New("the reaper wrote no usage line")
	}
	fields := strings.Fields(line)
	if len(fields) != 3 || !strings.HasSuffix(line, "\n") {
		return 0, 0, 0, fmt.Errorf("the reaper's usage line %q isn't three numbers and a newline", line)
	}
	numbers := [3]int64{}
	for index, field := range fields {
		number, err := strconv.ParseInt(field, 10, 64)
		if err != nil || number < 0 {
			return 0, 0, 0, fmt.Errorf("the reaper's usage line %q has %q, which isn't a count", line, field)
		}
		numbers[index] = number
	}
	return time.Duration(numbers[0]) * time.Microsecond, time.Duration(numbers[1]) * time.Microsecond, numbers[2] * 1024, nil
}

// outputKeptWhole is the most of a run's standard output kept as it was printed.
const outputKeptWhole = 64 << 10

// keptOutput is what a Measurement keeps of a run's standard output, which is only ever compared for equality and
// quoted by its first line: the output itself when it's small, and past outputKeptWhole its first line, its length and
// its sha256, which compare equal exactly when the outputs do. bench keeps every sample's measurements, and
// holdout/fasta prints 254 MB a run, so keeping it whole held tens of gigabytes over one comparison.
func keptOutput(output []byte) string {
	if len(output) <= outputKeptWhole {
		return string(output)
	}
	return fmt.Sprintf("%s\n(%d bytes in all, sha256 %x)\n", firstLine(string(output)), len(output), sha256.Sum256(output))
}

// Environment is this process's, without what would change how a runtime runs: BUN_OPTIONS (a shell that sets --smol
// makes Bun trade speed for a smaller heap) and NODE_OPTIONS, and without every variable named in without, which a
// caller sets for its children itself or wants left at its default.
func Environment(without ...string) []string {
	kept := []string{}
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name == "BUN_OPTIONS" || name == "NODE_OPTIONS" || slices.Contains(without, name) {
			continue
		}
		kept = append(kept, variable)
	}
	return kept
}

// NodeCommand runs the program at source with Node: a .a source, or one importing 'adamic' (ImportsAdamic), through
// the repository's oracle runner at root, which resolves 'adamic' to the independent sequential shim.
func NodeCommand(source string, importsAdamic bool, root string) []string {
	if importsAdamic || filepath.Ext(source) == ".a" {
		return []string{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle", "node.mjs"), source}
	}
	return []string{"node", source}
}

// ImportsAdamic is whether a program's source imports from 'adamic', which Node resolves only through the oracle
// runner.
func ImportsAdamic(contents string) bool {
	return strings.Contains(contents, "from 'adamic'") || strings.Contains(contents, `from "adamic"`)
}

// peakResident is the process's maximum resident set in bytes: Linux reports it in kilobytes, and macOS in bytes.
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

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}
