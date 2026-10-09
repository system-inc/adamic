package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"syscall"
	"time"

	"github.com/system-inc/cohere/internal/release/dispatch"
	"github.com/system-inc/cohere/internal/release/packaging"
)

// swiftPassThroughSwitches are the boolean flags the Swift engine reads with the meaning the contract
// gives them, forwarded only when the caller typed them. The rest are refused by name, by the engine
// or below, because a flag accepted and ignored reads as a run that did what was asked.
var swiftPassThroughSwitches = []string{
	"no-fix", "fix", "types", "lint", "format", "format-all", "single-threaded",
	"rules", "rules-enabled", "version", "unused", "unused-all", "unused-deep", "timing",
}

// swiftFrontDoorFlags are the flags the front door answers for a Swift run itself: where the package is,
// the lint config and fix passes it forwards with their values, how the report is printed, and the profile
// of this process. Any flag given that is in neither list is refused by name (see swiftEngineArguments).
// It used to be dropped, and `--format-only` dropped meant a format gate that type-checked and linted the
// whole package, minutes of compiling under a verdict that said formatting (#nqb3mjv).
var swiftFrontDoorFlags = []string{
	"directory", "lint-config", "fix-passes", "verbose", "json", "phases", "coverage", "profile",
}

// swiftRefusals say why a flag the engine cannot honor yet is refused, where the general sentence would
// not say enough.
var swiftRefusals = map[string]string{
	"format-only": "--format-only is not implemented for Swift yet: the engine has no format-only mode, and running " +
		"it would type-check and lint the package as well (#nqb3mjv), so nothing was checked",
	"no-format": "--no-format is not implemented for Swift yet: the engine formats on every run (swift/Contract.md), " +
		"so its format findings would still count (#nqb3mjv), and nothing was checked",
}

// runSwiftEngine checks a Swift package by running cohere-swift on it and rendering its records.
//
// A released cohere runs the engine its package ships beside it. A development cohere resolves one the
// way the launcher resolves cohere: from this module's `swift/` sources, rebuilt when they or the
// toolchain change. See resolveSwiftEngineBinary for the order.
func runSwiftEngine(location projectLocation, given map[string]bool, positionals []string) (int, error) {
	if given["tsconfig"] {
		return 1, fmt.Errorf(
			"--tsconfig names a TypeScript program, and %s is a Swift package (it holds a %s), so there is no tsconfig to use",
			location.Root, swiftProjectMarker)
	}

	arguments, mode, err := swiftEngineArguments(location, given, flagValue, positionals)
	if err != nil {
		return 1, err
	}

	// Building the engine is a step the launcher's progress shows: in full under --verbose, otherwise one
	// line on a terminal, cleared before the report (#ytqqv8v).
	dispatch.Report.SetVerbose(given["verbose"] && flagValue("verbose") == "true")
	binaryPath, sourceCommit, err := resolveSwiftEngineBinary(location, release.Current())
	dispatch.Report.Clear()
	if err != nil {
		return 1, err
	}

	run := newSwiftRun(os.Stdout, mode, location.rootNote(), processStart)
	run.details = given["coverage"] && flagValue("coverage") == "true"
	run.engineSourceCommit = sourceCommit
	run.requested = swiftRequestedPhases(given, flagValue)
	return runEngineBinary(binaryPath, arguments, run, os.Stderr)
}

// swiftEnginePlatform is the platform this cohere runs on. A variable so a test can run as a platform
// the machine running it is not.
var swiftEnginePlatform = release.Target{GoOperatingSystem: runtime.GOOS, GoArchitecture: runtime.GOARCH}

// resolveSwiftEngineBinary picks the engine, in this order:
//
//  1. The one the caller named with COHERE_SWIFT_ENGINE, on any platform and any build.
//  2. None, on a platform the engine does not ship for, refused by name before anything is looked for.
//     A Linux checkout would otherwise try to build the engine and bury the reason in a toolchain error.
//  3. For a released cohere, the engine its package ships beside it, and nothing else: see
//     shippedSwiftEngine.
//  4. For a development cohere, one built from this module's checkout, below.
//
// A development cohere never takes an engine from beside itself. A `go build -o` next to a stale
// cohere-swift would run that engine silently, and the contract version cannot catch rules that drifted
// inside one version. So each kind of build has one source: a release runs its sibling, development runs
// the checkout.
//
// A cohere binary that names its commit and has nothing uncommitted in it, which is what the launcher
// builds for every gate, gets an engine built from that same commit. Any other cohere, a `--dev` build
// from a dirty tree included, gets one built from the working tree, so a development run checks
// development rules in both engines and says so through both provenance records.
//
// The second value is the commit whose Swift sources the engine is guaranteed to match, or empty when
// nothing guarantees it: an engine named by the override, or one built from the working tree. The
// provenance is this cohere's, passed in rather than read here so a test can state a clean named commit,
// which a test binary never has, and prove the override still vouches for none.
func resolveSwiftEngineBinary(location projectLocation, provenance release.Provenance) (string, string, error) {
	if named := os.Getenv(dispatch.SwiftEngineOverrideVariable); named != "" {
		if !isRegularFile(named) {
			return "", "", fmt.Errorf("%s names %s, which is not a file, so nothing was checked", dispatch.SwiftEngineOverrideVariable, named)
		}
		fmt.Fprintf(os.Stderr, "cohere: running the Swift engine %s named by %s, not one built from the sources on disk\n",
			named, dispatch.SwiftEngineOverrideVariable)
		return named, "", nil
	}

	if !release.ShipsSwiftEngine(swiftEnginePlatform.GoOperatingSystem) {
		return "", "", fmt.Errorf("Swift is not available on this platform (%s): %s ships only for macOS, so nothing in %s was checked",
			swiftEnginePlatform, release.SwiftEngineFileName, location.Root)
	}

	if !provenance.IsDevelopment() {
		binaryPath, err := shippedSwiftEngine(provenance)
		return binaryPath, "", err
	}

	moduleDirectory, err := dispatch.FindModuleDirectory()
	if err != nil {
		return "", "", fmt.Errorf("%s is a Swift package, and the Swift engine is built from a cohere checkout: %w", location.Root, err)
	}
	commit := committedSwiftSource(provenance)
	binaryPath, _, err := dispatch.ResolveSwiftEngine(dispatch.DefaultPaths(moduleDirectory), commit, swiftContractVersion)
	if err != nil {
		return "", "", err
	}
	return binaryPath, commit, nil
}

// shippedSwiftEngine is the engine a released cohere's package ships beside it: release.SwiftEngineFileName
// in the directory of this executable, symlinks resolved, which is the package's `bin/`, since the npm
// launcher execs the platform binary directly.
//
// It never falls back to a checkout. A release run inside a cohere clone would otherwise build whatever
// rules were on disk there and report them under the release's version. A missing sibling is a broken
// install, and it says so rather than checking nothing. It vouches for no commit: the release stamps
// none into cohere, and the engine's own provenance record names what it was built from.
func shippedSwiftEngine(provenance release.Provenance) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding this cohere's own executable, beside which its Swift engine ships: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	binaryPath := filepath.Join(filepath.Dir(executable), release.SwiftEngineFileName)
	if !isRegularFile(binaryPath) {
		return "", fmt.Errorf("this cohere %s install has no Swift engine: expected %s, which the macOS package ships beside cohere, so nothing was checked; reinstall cohere",
			provenance.Version, binaryPath)
	}
	return binaryPath, nil
}

// committedSwiftSource is the commit the Swift engine is built from, or empty for the working tree.
//
// Only a cohere that names its commit and was built from a tree with nothing uncommitted in it has a
// commit to offer: anything else could not be reproduced from one, so neither could an engine built to
// match it. That empty answer is also what keeps `cohere --version` from claiming a commit for an engine
// nothing vouches for.
func committedSwiftSource(provenance release.Provenance) string {
	if provenance.SelfCommit != "" && !provenance.SourceTreeModified {
		return provenance.SelfCommit
	}
	return ""
}

// flagValue reads a parsed flag's value by name.
func flagValue(name string) string {
	if found := flag.Lookup(name); found != nil {
		return found.Value.String()
	}
	return ""
}

// swiftEngineArguments builds the engine's command line, `--contract 1 --root <root> [flags] [paths]`,
// and says which question the run answers.
//
// Every path is made absolute against where the caller typed it, so the engine never has to know what
// `--directory` meant. Takes the flag reader as a parameter so a test can state the flags it means
// without touching the process's own.
func swiftEngineArguments(
	location projectLocation,
	given map[string]bool,
	value func(string) string,
	positionals []string,
) ([]string, swiftMode, error) {
	for _, refused := range []string{"explain"} {
		if given[refused] {
			// Refused here rather than forwarded for the engine to refuse, because its argument is a
			// path and the engine's sentence would name the path rather than the flag.
			return nil, "", fmt.Errorf("--%s is not implemented for Swift yet, so nothing was checked", refused)
		}
	}
	// Sorted, so a command line naming two of them is refused for the same one every time.
	for _, name := range slices.Sorted(maps.Keys(given)) {
		// A switch turned off asks for nothing, so there is nothing to honor or refuse.
		if slices.Contains(swiftPassThroughSwitches, name) || slices.Contains(swiftFrontDoorFlags, name) ||
			name == "tsconfig" || value(name) == "false" {
			continue
		}
		if reason, explained := swiftRefusals[name]; explained {
			return nil, "", errors.New(reason)
		}
		return nil, "", fmt.Errorf("--%s is not implemented for Swift yet, so the run was refused rather than run without it, and nothing was checked", name)
	}

	arguments := []string{"--contract", fmt.Sprint(swiftContractVersion), "--root", location.Root}
	for _, name := range swiftPassThroughSwitches {
		if given[name] && value(name) == "true" {
			arguments = append(arguments, "--"+name)
		}
	}
	if given["lint-config"] {
		arguments = append(arguments, "--lint-config", location.LintConfigFileName)
	}
	if given["fix-passes"] {
		arguments = append(arguments, "--fix-passes", value("fix-passes"))
	}
	for _, positional := range positionals {
		arguments = append(arguments, absoluteFrom(location.ArgumentBase, positional))
	}

	// The same precedence the TypeScript run gives these, so one command line asks one question
	// whichever engine answers it.
	mode := swiftModeCheck
	switch {
	case given["rules"] && value("rules") == "true":
		mode = swiftModeRules
	case given["rules-enabled"] && value("rules-enabled") == "true":
		mode = swiftModeRulesEnabled
	case given["version"] && value("version") == "true":
		mode = swiftModeVersion
	}
	return arguments, mode, nil
}

// engineWaitDelay bounds how long Wait may wait on the engine's pipes after the engine has ended. Killing
// the engine's process group closes them at once; this is the backstop for a descendant that left the
// group, so a stray holding a pipe open can never hold the front door open with it.
const engineWaitDelay = 2 * time.Second

// runEngineBinary runs an engine, renders its stdout record by record as it arrives, passes its
// stderr through untouched, and returns the exit code the front door owns.
//
// Records are rendered as they stream so a slow types phase shows its fix line first, as a TypeScript
// run does. A record that breaks the contract stops the run there: the engine is killed, the phase
// line says what was not reached, and the error says the engine is broken.
func runEngineBinary(binaryPath string, arguments []string, run *swiftRun, standardError io.Writer) (int, error) {
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupts)
	return runEngineBinaryUntil(binaryPath, arguments, run, standardError, interrupts)
}

// runEngineBinaryUntil is runEngineBinary with the signals to forward handed in, so a test can deliver
// one without signalling its own process.
//
// The engine runs in a process group of its own, so ending it ends everything it started: a swift build
// or a compiler it spawned would otherwise outlive it, still running and still holding its output pipes
// open. A group of its own is also out of the terminal's reach, so an interrupt the terminal sends to
// cohere is passed on to the engine's group here, every time it arrives, rather than left to orphan it.
func runEngineBinaryUntil(
	binaryPath string,
	arguments []string,
	run *swiftRun,
	standardError io.Writer,
	interrupts <-chan os.Signal,
) (int, error) {
	command := exec.Command(binaryPath, arguments...)
	command.Stderr = standardError
	startInOwnProcessGroup(command)
	command.WaitDelay = engineWaitDelay
	standardOutput, err := command.StdoutPipe()
	if err != nil {
		return 1, fmt.Errorf("connecting to the Swift engine's output: %w", err)
	}
	if err := command.Start(); err != nil {
		return 1, fmt.Errorf("starting the Swift engine %s: %w", binaryPath, err)
	}

	// With Setpgid the engine leads its group, so the group's id is the engine's process id.
	engineGroup := command.Process.Pid
	runEnded := make(chan struct{})
	defer close(runEnded)
	go func() {
		for {
			select {
			case received := <-interrupts:
				signalProcessGroup(engineGroup, received)
			case <-runEnded:
				return
			}
		}
	}()

	scanner := bufio.NewScanner(standardOutput)
	// A finding's message is bounded by what a rule writes, not by a line length, so the limit is far
	// past any real record and a line past it is still refused loudly rather than split.
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)

	var protocolError error
	for scanner.Scan() {
		if protocolError = run.accept(scanner.Bytes()); protocolError != nil {
			break
		}
	}
	if protocolError == nil && scanner.Err() != nil {
		protocolError = fmt.Errorf("reading its output: %w", scanner.Err())
	}

	if protocolError != nil {
		// Killed rather than drained: nothing it writes after breaking the contract can be trusted, and
		// waiting for it to finish would make the reader wait for output that will be discarded. The
		// whole group, because a child it started would keep running and keep the pipes open.
		killEngineGroup(engineGroup, command)
		run.writeUnfinished("the Swift engine broke its contract")
		return 1, fmt.Errorf("the Swift engine is broken, so nothing it reported can be trusted: %w", protocolError)
	}

	waitError := command.Wait()
	exitCode, ended := describeExit(command.ProcessState, waitError)
	return run.finish(exitCode, ended)
}

// killEngineGroup kills every process in the engine's group and waits for the engine.
//
// One kill is not enough: a child the engine forks at the instant the group is signalled can miss the
// signal, and it would outlive the engine holding its pipes. So the group is signalled again until the
// engine is reaped and its pipes are closed. Signalling it again is safe, because a process id is never
// reused while a process group of that id still has members.
func killEngineGroup(engineGroup int, command *exec.Cmd) {
	killProcessGroup(engineGroup)
	waited := make(chan struct{})
	go func() {
		_ = command.Wait()
		close(waited)
	}()
	for {
		select {
		case <-waited:
			return
		case <-time.After(20 * time.Millisecond):
			killProcessGroup(engineGroup)
		}
	}
}

// describeExit turns how a process ended into an exit code and the words for it.
//
// A process killed by a signal has no exit code, and Go reports -1 for it. Printing "exited -1" would
// state a number the process never chose, so a signal is named as a signal.
func describeExit(state *os.ProcessState, waitError error) (int, string) {
	if state == nil {
		return -1, fmt.Sprintf("could not be waited for (%v)", waitError)
	}
	if status, isUnix := state.Sys().(syscall.WaitStatus); isUnix && status.Signaled() {
		return -1, fmt.Sprintf("was killed by %s", status.Signal())
	}
	var exitError *exec.ExitError
	if waitError != nil && !errors.As(waitError, &exitError) {
		return -1, fmt.Sprintf("could not be waited for (%v)", waitError)
	}
	return state.ExitCode(), fmt.Sprintf("exited %d", state.ExitCode())
}
