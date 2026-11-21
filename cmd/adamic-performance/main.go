// Command adamic-performance writes Performance.jsonc, Adamic's performance at one commit as a single snapshot:
//
//	adamic-performance -adamic <adamic> -out <path> [-previous <path>] -commit <40-hex sha> [-rounds 5] [-held-quiet]
//	adamic-performance -out <path> -previous <path> -commit <40-hex sha> -unmeasured "<why>"
//
// The line runs it from the repository's root in every landing commit and commits what it writes, so git's history of
// the file is the series and each landing's diff shows what moved (#hed1mkb). It measures the compile half: a cold
// build of hello, then warm builds and adamic c of four programs, interleaved over the rounds, each written as the
// median of the rounds with their min and max. The execute half, native programs against Node, is carried from
// -previous as it was. Every child gets a temporary directory's caches, so the cold build is cold and nothing touches
// this machine's.
//
// The note compares against -previous: a number moved only when its new range doesn't overlap the old one, and a
// different machine, tool or round count is named, never compared (compareNote). The second form measures nothing: it
// keeps -previous's numbers under the new commit and says why (unmeasuredSnapshot).
//
// Nothing that builds or tests reads the file, which is written after the gate; the noreader package holds that.
package main

import (
	"bytes"
	"debug/buildinfo"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// fullCommit is a commit as -commit takes it: all 40 hex digits, so the file names one commit forever.
var fullCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)

// printedLines is how many of a failed child's first lines its error carries.
const printedLines = 12

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(arguments []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("adamic-performance", flag.ContinueOnError)
	flags.SetOutput(stderr)
	adamic := flags.String("adamic", "", "the built adamic binary to measure")
	out := flags.String("out", "", "where to write the snapshot")
	previousPath := flags.String("previous", "", "the snapshot before this one, to compare against and to carry the execute half from")
	commit := flags.String("commit", "", "the 40-hex commit the snapshot is of")
	rounds := flags.Int("rounds", 5, "how many interleaved rounds of warm builds and of adamic c")
	heldQuiet := flags.Bool("held-quiet", false, "the machine was held quiet for the measure")
	unmeasured := flags.String("unmeasured", "", "measure nothing, keep -previous's numbers, and say why")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	unmeasuredGiven := false
	flags.Visit(func(set *flag.Flag) {
		if set.Name == "unmeasured" {
			unmeasuredGiven = true
		}
	})
	if err := write(writeRequest{adamic: *adamic, out: *out, previous: *previousPath, commit: *commit, rounds: *rounds, heldQuiet: *heldQuiet, unmeasured: *unmeasured, unmeasuredGiven: unmeasuredGiven}); err != nil {
		fmt.Fprintln(stderr, "adamic-performance:", err)
		return 1
	}
	return 0
}

// A writeRequest is the command's flags.
type writeRequest struct {
	adamic          string
	out             string
	previous        string
	commit          string
	rounds          int
	heldQuiet       bool
	unmeasured      string
	unmeasuredGiven bool
}

// write makes the snapshot the request asks for and writes it to its out, or writes nothing: a file with a number
// missing is never written.
func write(request writeRequest) error {
	if request.out == "" {
		return errors.New("-out names where to write the snapshot")
	}
	if !fullCommit.MatchString(request.commit) {
		return fmt.Errorf("-commit takes a full 40-hex commit, and %q isn't one", request.commit)
	}
	var previous *snapshot
	if request.previous != "" {
		data, err := os.ReadFile(request.previous)
		if err != nil {
			return err
		}
		read, err := decodeSnapshot(data)
		if err != nil {
			return fmt.Errorf("-previous %s: %w", request.previous, err)
		}
		previous = &read
	}
	var written snapshot
	if request.unmeasuredGiven {
		if strings.TrimSpace(request.unmeasured) == "" {
			return errors.New("-unmeasured takes why the commit wasn't measured")
		}
		if previous == nil {
			return errors.New("-unmeasured keeps -previous's numbers, so it needs -previous")
		}
		written = unmeasuredSnapshot(*previous, request.commit, request.unmeasured)
	} else {
		measured, err := measureSnapshot(request, previous)
		if err != nil {
			return err
		}
		written = measured
	}
	data, err := encodeSnapshot(written)
	if err != nil {
		return err
	}
	return os.WriteFile(request.out, data, 0o644)
}

// measureSnapshot measures this commit from the repository's root, the working directory, and notes it against
// previous.
func measureSnapshot(request writeRequest, previous *snapshot) (snapshot, error) {
	if request.adamic == "" {
		return snapshot{}, errors.New("-adamic names the built adamic binary to measure")
	}
	if request.rounds < 1 {
		return snapshot{}, fmt.Errorf("-rounds is at least 1, and it's %d", request.rounds)
	}
	// Every child runs in its program's own directory, so a relative binary would name another file there.
	adamic, err := filepath.Abs(request.adamic)
	if err != nil {
		return snapshot{}, err
	}
	root, err := os.Getwd()
	if err != nil {
		return snapshot{}, err
	}
	if module, err := os.ReadFile(filepath.Join(root, "go.mod")); err != nil || !strings.Contains(string(module), "module github.com/system-inc/adamic\n") {
		return snapshot{}, fmt.Errorf("adamic-performance runs from the adamic repository's root, and %s isn't it", root)
	}
	// The tools are read first, so a toolchain that can't answer fails before minutes of measuring.
	versions, err := readTools(adamic)
	if err != nil {
		return snapshot{}, err
	}
	name, err := os.Hostname()
	if err != nil {
		return snapshot{}, err
	}
	work, err := os.MkdirTemp("", "adamic-performance-")
	if err != nil {
		return snapshot{}, err
	}
	defer os.RemoveAll(work)

	started := time.Now()
	loadBefore, err := loadAverage()
	if err != nil {
		return snapshot{}, err
	}
	measured, err := measureCompile(compileRequest{root: root, work: work, rounds: request.rounds}, adamicRunner(adamic))
	if err != nil {
		return snapshot{}, err
	}
	loadAfter, err := loadAverage()
	if err != nil {
		return snapshot{}, err
	}

	current := snapshot{
		Schema:         schemaVersion,
		Commit:         request.commit,
		MeasuredCommit: request.commit,
		MeasuredAt:     started.UTC().Format(time.RFC3339),
		Machine:        machine{Name: name, CPU: cpuModel(), Cores: runtime.NumCPU(), LoadBefore: loadBefore, LoadAfter: loadAfter, HeldQuiet: request.heldQuiet},
		Tools:          versions,
		Compile:        measured,
	}
	if previous != nil {
		current.Execute = previous.Execute
	}
	current.Note = compareNote(previous, current)
	return current, nil
}

// readTools reads the versions the numbers depend on. clang's is the one native_runtime keys the runtime archive by,
// read through it, so this command names no C compiler of its own (internal/compilercensus); the adamic binary's Go is
// from its build information; Node's is left out where there is none on PATH.
func readTools(adamic string) (tools, error) {
	information, err := buildinfo.ReadFile(adamic)
	if err != nil {
		return tools{}, fmt.Errorf("reading %s's build information: %w", adamic, err)
	}
	_, _, _, clangVersion, err := native_runtime.RuntimeInputs("", native_runtime.Options{}, nil)
	if err != nil {
		return tools{}, err
	}
	versions := tools{Adamic: information.GoVersion, Clang: firstLine(clangVersion)}
	if node, err := exec.LookPath("node"); err == nil {
		output, err := exec.Command(node, "--version").Output()
		if err != nil {
			return tools{}, fmt.Errorf("%s --version: %w", node, err)
		}
		versions.Node = firstLine(string(output))
	}
	return versions, nil
}

// adamicRunner runs children of the adamic binary at adamic, each clocked by wall time around its process. A child's
// environment is this process's, without the settings the measure gives its own, so a log or a cache set where this
// runs never reaches a child the measure didn't hand one.
func adamicRunner(adamic string) runner {
	return func(child child) (childResult, error) {
		command := exec.Command(adamic, child.arguments...)
		command.Dir = child.directory
		command.Env = append(environment(), child.environment...)
		var stdout, stderr bytes.Buffer
		command.Stderr = &stderr
		if child.keepOutput {
			command.Stdout = &stdout
		} else {
			// The null device is a file, so the child writes to it directly and nothing copies its output.
			discard, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
			if err != nil {
				return childResult{}, err
			}
			defer discard.Close()
			command.Stdout = discard
		}
		started := time.Now()
		err := command.Run()
		elapsed := time.Since(started)
		if err != nil {
			return childResult{}, fmt.Errorf("adamic %s in %s: %v\n%s", strings.Join(child.arguments, " "), child.directory, err, firstLines(stderr.String()+stdout.String(), printedLines))
		}
		return childResult{elapsed: elapsed, stdout: stdout.Bytes()}, nil
	}
}

// environment is this process's, without the variables the measure sets for its children.
func environment() []string {
	kept := []string{}
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name == "ADAMIC_BUILD_LOG" || name == "ADAMIC_BUILD_CACHE_DIR" || name == "XDG_CACHE_HOME" {
			continue
		}
		kept = append(kept, variable)
	}
	return kept
}

// loadAverage is the machine's 1, 5 and 15 minute load, from /proc/loadavg on Linux and sysctl on macOS.
func loadAverage() ([3]float64, error) {
	text := ""
	if contents, err := os.ReadFile("/proc/loadavg"); err == nil {
		text = string(contents)
	} else if output, err := exec.Command("sysctl", "-n", "vm.loadavg").Output(); err == nil {
		text = strings.Trim(strings.TrimSpace(string(output)), "{} ")
	} else {
		return [3]float64{}, errors.New("no load average: neither /proc/loadavg nor sysctl vm.loadavg answered")
	}
	fields := strings.Fields(text)
	if len(fields) < 3 {
		return [3]float64{}, fmt.Errorf("the load average %q isn't three numbers", text)
	}
	var load [3]float64
	for index := range load {
		value, err := strconv.ParseFloat(fields[index], 64)
		if err != nil {
			return [3]float64{}, fmt.Errorf("the load average %q isn't three numbers", text)
		}
		load[index] = value
	}
	return load, nil
}

// cpuModel is the processor's name, as bench/run.go's cpuModel reads it.
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

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

// firstLines is text's first count lines, so a failed child's error says what it printed without all of it.
func firstLines(text string, count int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.Join(lines[:min(count, len(lines))], "\n")
}
