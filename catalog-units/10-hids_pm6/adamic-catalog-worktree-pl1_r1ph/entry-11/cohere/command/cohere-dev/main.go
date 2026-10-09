// Command cohere-dev runs the house's heavy Go work within the machine's budget (#3kr3x59, #qhg0ntb).
//
//	go run ./command/cohere-dev test ./...            every package: the landing gate
//	go run ./command/cohere-dev test ./internal/edit  one package
//	go run ./command/cohere-dev test --fast           the edit loop: all but the landing gate's own (fast.go)
//	go run ./command/cohere-dev build|vet [packages]  go build or go vet
//	go run ./command/cohere-dev exec -- <command>     any other heavy command: a bench, a cohere run, clang
//	go run ./command/cohere-dev land                  merge main, gate and fast-forward main, one landing at a time
//	go run ./command/cohere-dev status                the pool: its tokens, their holders, the line, the cache
//	go run ./command/cohere-dev quiet on|off          a quiet window: the cache trim holds off until off (quiet.go)
//
// Every one of them takes a token from the machine's pool, in arrival order, and runs with Go held to the
// token's share of the cores (pool.go). A dozen members each running go on all sixteen cores held the machine
// at load 110 to 180, and made every run slower, not only later. When a token's run ends, the Go build
// cache is looked at, and trimmed under its cap with the pool held (cache.go).
//
// A whole-module run never passes -count=1. Go's test cache skips every package whose inputs did not
// change, which on a one-file change is most of the module, and -count=1 throws that away. A flaky test
// is fixed, not re-run around; name its package to re-run it alone.
//
// Every run is nice 10, and so is everything it starts, so heavy work uses the cores no one else is using
// and yields to a person typing. --full-priority keeps the priority the run was started with, for a
// measurement.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "status":
		os.Exit(status())
	case trimCacheVerb:
		os.Exit(trimCache())
	case "quiet":
		os.Exit(quiet(os.Args[2:]))
	case "test", "build", "vet", "exec", "land":
		os.Exit(run(os.Args[1], os.Args[2:]))
	}
	usage()
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cohere-dev test [--fast] [--full-priority] [go test flags] [packages]\n"+
		"       cohere-dev build|vet [--full-priority] [go flags] [packages]\n"+
		"       cohere-dev exec [--full-priority] -- <command> [arguments]\n"+
		"       cohere-dev land\n"+
		"       cohere-dev status\n"+
		"       cohere-dev quiet [on [--for <duration>] [why] | off]")
	os.Exit(2)
}

// fullPriorityFlag keeps a run at the priority it was started with, for a timed measurement. Without it a
// run lowers itself to testNiceness first.
const fullPriorityFlag = "--full-priority"

// testNiceness is how nice a run makes itself, and so every go command and test binary it starts. Tests
// compiling and running on every core held the machine at load 168 on 16 cores and made Kirk's typing lag
// (2026-10-05); at nice 10 they still take every idle cycle and yield to what a person is doing.
const testNiceness = 10

// nicenessVariable sets another niceness than testNiceness, from 0 to 19, for whoever wants their tests
// to yield more, and for the test that shows a run is niced whatever priority it was started at.
const nicenessVariable = "COHERE_DEV_NICENESS"

// niceness is how nice a run makes itself: testNiceness, or what nicenessVariable says.
func niceness() (int, error) {
	value := os.Getenv(nicenessVariable)
	if value == "" {
		return testNiceness, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > 19 {
		return 0, fmt.Errorf("%s is %q, want a whole number from 0 to 19", nicenessVariable, value)
	}
	return parsed, nil
}

// run runs a heavy command, niced, through a token: go test, go build, go vet, or any command with exec.
func run(verb string, arguments []string) int {
	arguments, fullPriority := withoutFlag(arguments, fullPriorityFlag)
	if !fullPriority {
		level, err := niceness()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
			return 2
		}
		if err := lowerPriority(level); err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: lowering this run's priority: %v\n", err)
			return 1
		}
	}

	switch verb {
	case "land":
		return land(arguments)
	case "exec":
		if len(arguments) > 0 && arguments[0] == "--" {
			arguments = arguments[1:]
		}
		if len(arguments) == 0 {
			usage()
		}
		return withToken(strings.Join(arguments, " "), func(environment []string) int {
			return runCommand(environment, arguments[0], arguments[1:])
		})
	case "build", "vet":
		return withToken("go "+verb+" "+strings.Join(arguments, " "), func(environment []string) int {
			return runCommand(environment, "go", append([]string{verb}, arguments...))
		})
	}

	fast := len(arguments) > 0 && arguments[0] == "--fast"
	if fast {
		arguments = arguments[1:]
	}
	arguments = withEveryVetCheck(arguments)
	if fast {
		// The fast tier rests on Go's test cache, and runs nearly every package, so -count=1 here is the
		// costliest run there is on a loaded machine (@system_cohere_build's catch).
		if countsOnce(arguments) {
			refuseCountOnce()
			return 2
		}
		return withToken("go test --fast "+strings.Join(arguments, " "), func(environment []string) int {
			return fastTest(environment, arguments)
		})
	}
	if wholeModule(arguments) && countsOnce(arguments) {
		refuseCountOnce()
		return 2
	}
	// A whole-module run from the root covers what the landing gate covers: every module go.work names
	// outside a submodule (workspace.go).
	if wholeModule(arguments) {
		expanded, err := withWorkspaceModules(arguments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: reading the modules to test: %v\n", err)
			return 2
		}
		arguments = expanded
	}
	return withToken("go test "+strings.Join(arguments, " "), func(environment []string) int {
		return goTest(environment, arguments)
	})
}

// withEveryVetCheck runs every check go vet runs, the landing gate's set, in a test run's vet, unless the caller
// chose a -vet of their own. go test runs a curated subset by default, which leaves out composites: 130
// unkeyed literals of another package's struct passed cohere-dev test and failed plain go vet (#kr2yp54).
func withEveryVetCheck(arguments []string) []string {
	for _, argument := range arguments {
		if argument == "-vet" || argument == "--vet" || strings.HasPrefix(argument, "-vet=") || strings.HasPrefix(argument, "--vet=") {
			return arguments
		}
	}
	return append([]string{"-vet=all"}, arguments...)
}

// withoutFlag returns the arguments with every copy of flag removed, and whether there was one, so a flag of
// this wrapper's never reaches go test.
func withoutFlag(arguments []string, flag string) ([]string, bool) {
	kept := make([]string, 0, len(arguments))
	found := false
	for _, argument := range arguments {
		if argument == flag {
			found = true
			continue
		}
		kept = append(kept, argument)
	}
	return kept, found
}

// wholeModule reports whether the arguments name every package, the run that takes a slot.
func wholeModule(arguments []string) bool {
	for _, argument := range arguments {
		if argument == "./..." || argument == "..." {
			return true
		}
	}
	return false
}

// withWorkspaceModules replaces a ./... run from the repository's root with the patterns of every module
// the landing gate covers, saying which. From anywhere else ./... means that directory's tree, and stays.
func withWorkspaceModules(arguments []string) ([]string, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(root, "go.work")); err != nil {
		return arguments, nil
	}
	covered, err := readWorkspace(root)
	if err != nil {
		return nil, err
	}
	covered.report(os.Stderr)
	expanded := make([]string, 0, len(arguments)+len(covered.gated))
	for _, argument := range arguments {
		if argument == "./..." {
			expanded = append(expanded, covered.patterns()...)
			continue
		}
		expanded = append(expanded, argument)
	}
	return expanded, nil
}

// refuseCountOnce says why a run of nearly every package, through the slot or the fast tier, never passes
// -count=1.
func refuseCountOnce() {
	fmt.Fprintln(os.Stderr, "cohere-dev: a whole-module run never passes -count=1: it re-runs every package whose "+
		"inputs did not change, which Go's test cache would have skipped. Fix a flaky test rather than re-run "+
		"around it, and name a package to re-run it alone: go test -count=1 ./its/package")
}

// countsOnce reports whether the arguments turn Go's test cache off with -count=1.
func countsOnce(arguments []string) bool {
	for index, argument := range arguments {
		switch {
		case argument == "-count=1" || argument == "--count=1":
			return true
		case (argument == "-count" || argument == "--count") && index+1 < len(arguments) && arguments[index+1] == "1":
			return true
		}
	}
	return false
}

// goTest runs go test with the caller's terminal and returns its exit code.
func goTest(base []string, arguments []string) int {
	environment, uncovered, err := testEnvironment(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 2
	}
	command := exec.Command("go", append([]string{"test"}, arguments...)...)
	command.Env = environment
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	defer printUncovered(os.Stderr, uncovered)
	if err := command.Run(); err != nil {
		if exited, ok := err.(*exec.ExitError); ok {
			return exited.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "cohere-dev: running go test: %v\n", err)
		return 1
	}
	return 0
}

// runCommand runs a command with the caller's terminal in the environment given, and returns its exit code.
func runCommand(environment []string, name string, arguments []string) int {
	command := exec.Command(name, arguments...)
	command.Env = environment
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		if exited, ok := err.(*exec.ExitError); ok {
			return exited.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "cohere-dev: running %s: %v\n", name, err)
		return 1
	}
	return 0
}

// holderLine describes the run holding a token: who, where, since when and what it runs.
func holderLine(what string) string {
	directory, _ := os.Getwd()
	return fmt.Sprintf("pid %d in %s since %s: %s", os.Getpid(), directory, time.Now().Format("15:04:05"), what)
}

// holders reads who holds each taken slot. A slot that is free says nothing.
func holders(directory string, slots int) []string {
	lines := []string{}
	for slot := 1; slot <= slots; slot++ {
		contents, err := os.ReadFile(holderPath(directory, slot))
		if err == nil && len(contents) > 0 {
			lines = append(lines, strings.TrimSpace(string(contents)))
		}
	}
	return lines
}

func lockPath(directory string, slot int) string {
	return filepath.Join(directory, fmt.Sprintf("slot-%d.lock", slot))
}

func holderPath(directory string, slot int) string {
	return filepath.Join(directory, fmt.Sprintf("slot-%d.holder", slot))
}

// heldSlot is a token this process holds until release, or until it exits.
type heldSlot struct {
	number int
	lock   *os.File
	holder string
}

func (held *heldSlot) release() {
	os.Remove(held.holder)
	unlockFile(held.lock)
	held.lock.Close()
}

// takeSlot takes the first free token, or returns nil when every token is held.
func takeSlot(directory string, slots int, what string) (*heldSlot, error) {
	for slot := 1; slot <= slots; slot++ {
		lock, err := os.OpenFile(lockPath(directory, slot), os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return nil, err
		}
		taken, err := tryLockFile(lock)
		if err != nil {
			lock.Close()
			return nil, err
		}
		if !taken {
			lock.Close()
			continue
		}
		holder := holderPath(directory, slot)
		if err := os.WriteFile(holder, []byte(holderLine(what)+"\n"), 0o644); err != nil {
			unlockFile(lock)
			lock.Close()
			return nil, err
		}
		return &heldSlot{number: slot, lock: lock, holder: holder}, nil
	}
	return nil, nil
}
