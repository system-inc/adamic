package fuzz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/nodepin"
)

// Checkout is a compiler under test: a checkout of the repository, with its own adamic command built
// and its own runtime compiled once under the sanitizers. The fuzzer drives a checkout through its
// command line rather than in process, so the same fuzzer can run against any commit, an old one
// included, to prove it finds a bug that commit had.
type Checkout struct {
	Root      string
	directory string
	adamic    string
	runtime   string
}

// flags are native.Build's for a sanitized build, what the oracle compiles with. They're this tree's,
// even for a checkout of another commit: the C and the runtime are the checkout's, the flags aren't.
var flags = native.Flags(native.Options{Sanitize: true})

// Prepare builds a checkout's adamic command into directory and gets its cached runtime library.
// The cache is shared with native.Build, while another checkout keeps its own runtime bytes.
func Prepare(root string, directory string) (*Checkout, error) {
	if _, err := nodepin.Check(); err != nil {
		return nil, err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	checkout := &Checkout{Root: root, directory: directory, adamic: filepath.Join(directory, "adamic")}
	build, release := boundedrun.Command(boundedrun.Build, "go", "build", "-o", checkout.adamic, "./cmd/adamic")
	defer release()
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("fuzz: building adamic in %s: %w\n%s", root, err, output)
	}
	checkout.runtime, err = boundedRuntimeLibrary(filepath.Join(root, "internal", "native", "runtime"))
	if err != nil {
		return nil, fmt.Errorf("fuzz: compiling the runtime: %w", err)
	}
	return checkout, nil
}

// Run is one execution's observable behavior, as the oracle compares it.
type Run struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	// Signal is the signal that ended the process, by name ("segmentation fault"), or "" when it
	// exited on its own. ExitCode is then -1.
	Signal   string
	TimedOut bool
}

// Verdict is what came of trying one program.
type Verdict string

const (
	// Crash: native died by a signal or a sanitizer reported a bad access or undefined behavior
	// (AddressSanitizer, UndefinedBehaviorSanitizer, ThreadSanitizer), or the JavaScript backend died by
	// a signal, whatever Node did short of misbehaving itself. It's the most dangerous verdict there
	// is: a type the checker believed that the program's values didn't hold, reaching memory. Node
	// ending cleanly, or throwing an ordinary JavaScript exception at another point or even at the
	// same one, doesn't make it less: Node threw, native corrupted memory. Key says which way crashed
	// and how, as one line ("native AddressSanitizer: heap-use-after-free", "javascript backend
	// signal: abort trap").
	Crash Verdict = "crash"
	// Agreed: all three ways agree, and the sanitizers found nothing.
	Agreed Verdict = "agreed"
	// Invalid: the checker refused the program, so the generator wrote something that isn't Adamic.
	Invalid Verdict = "invalid"
	// NotYet: stage 0 said it can't lower something yet. Not a finding, but counted, since the
	// generator should stay inside what stage 0 lowers.
	NotYet Verdict = "not yet"
	// Checked: native and the JavaScript backend stopped at the same inserted check where the source
	// on Node runs on. That's Adamic meaning what it says, not a bug. The panic line has to be one of
	// the inserted checks' (insertedCheck), not just any stop the two backends share.
	Checked Verdict = "checked"
	// Unfit: the program misbehaved on Node itself (it never ended, or printed megabytes), which is
	// the generator's fault.
	Unfit Verdict = "unfit"
	// Finding: a disagreement, a leak (LeakSanitizer's report, or a leak run that failed), the compiler
	// crashing (its own Go panic, not the program's), or C that clang refused.
	Finding Verdict = "finding"
)

// Verdicts are every verdict, the most severe first: the order a summary lists them in.
var Verdicts = []Verdict{Crash, Finding, Agreed, Checked, NotYet, Invalid, Unfit}

// Outcome is a verdict with what it rests on. Key names the kind of finding, so shrinking can keep a
// candidate only when it still fails the same way.
type Outcome struct {
	Verdict Verdict
	Key     string
	Detail  string
	// Clang is what clang said when it refused the C, whole.
	Clang   string
	Node    Run
	Native  Run
	Backend Run
}

// Try runs a program's source three ways in a directory of its own, the way the oracle does.
func (c *Checkout) Try(source string, directory string) Outcome {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	path := filepath.Join(directory, "program.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	return c.TryFile(path, directory)
}

// TryFile runs a program file three ways, building in directory.
func (c *Checkout) TryFile(path string, directory string) Outcome {
	binary, failed := c.build(path, directory, true)
	if failed != nil {
		return *failed
	}
	oracle := filepath.Join(c.Root, "oracle", "node.mjs")
	outcome := Outcome{
		Node:    execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, path),
		Native:  execute(directory, []string{"ASAN_OPTIONS=detect_leaks=0"}, 20*time.Second, binary),
		Backend: execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, filepath.Join(directory, "program.mjs")),
	}
	return c.judge(outcome, binary, directory)
}

// build makes a program file's C and compiles it with clang, and with javascript makes its
// JavaScript too. It returns the binary, or the outcome a failure comes to.
func (c *Checkout) build(path string, directory string, javascript bool) (string, *Outcome) {
	fail := func(outcome Outcome) (string, *Outcome) { return "", &outcome }
	// The C first: the checker's and stage 0's refusals come from here.
	lowered := execute(directory, nil, 30*time.Second, c.adamic, "c", path)
	if lowered.ExitCode != 0 || lowered.TimedOut {
		return fail(compilerRefusal(lowered))
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), lowered.Stdout, 0o644); err != nil {
		return fail(Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()})
	}
	if javascript {
		emitted := execute(directory, nil, 30*time.Second, c.adamic, "js", path)
		if emitted.ExitCode != 0 || emitted.TimedOut {
			return fail(Outcome{Verdict: Finding, Key: "javascript backend failed", Detail: string(emitted.Stderr)})
		}
		if err := os.WriteFile(filepath.Join(directory, "program.mjs"), emitted.Stdout, 0o644); err != nil {
			return fail(Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()})
		}
	}
	binary := filepath.Join(directory, "program")
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(c.runtime), "-o", binary, filepath.Join(directory, "main.c"))
	arguments = append(arguments, native.RuntimeLinkFlags(c.runtime)...)
	arguments = append(arguments, "-lm")
	clang, release := boundedrun.Command(2*time.Minute, "clang", arguments...)
	defer release()
	if output, err := clang.CombinedOutput(); err != nil {
		key := "clang refused the C"
		if warning := clangWarning.FindSubmatch(output); warning != nil {
			key += ": " + string(warning[1])
		}
		return fail(Outcome{Verdict: Finding, Key: key, Detail: firstLines(string(output), 12), Clang: string(output)})
	}
	return binary, nil
}

// sanitizerReport finds the first report of a bad access or undefined behavior a sanitizer wrote, if
// any: AddressSanitizer's or ThreadSanitizer's headline (ERROR: AddressSanitizer: SEGV on unknown
// address ..., WARNING: ThreadSanitizer: data race ...), or UndefinedBehaviorSanitizer's "runtime
// error" with its words. LeakSanitizer's isn't one: a leak is a leak, not memory corruption, and the
// leak run below makes it a finding.
var sanitizerReport = regexp.MustCompile(`(?:ERROR|WARNING): ((?:Address|Thread|UndefinedBehavior)Sanitizer): ([^\n]*)|runtime error: ([^\n]*)`)

// clangWarning is the warning -Werror made an error, by its flag, so two kinds of bad C stay apart.
var clangWarning = regexp.MustCompile(`\[-Werror,(-W[a-z-]+)\]`)

// insertedCheck is the panic line of a check Adamic inserts where JavaScript doesn't stop: it goes
// on with undefined or null a call put back after a narrowing (ir.Defined, ir.Unwrap), with the
// holes of an array a map's or a find's callback shrank, with an array a write past its end grew or
// a fractional index gave a property, and with an object an as cast doesn't look at (ir.CheckedCast).
// A check whose words are JavaScript's own (TypeError: Cannot read properties of undefined, and the
// rest) is where Node throws too, so Node's run stops there and the two agree; a shared stop on one
// of those, or on anything else, while Node runs on, is a finding.
var insertedCheck = regexp.MustCompile(`^adamic: panic: (?:(?:undefined|null) where the checker narrowed it away: a call since the narrowing put it back|map: the array shrank while it was being mapped|[A-Za-z]+: the array shrank while it was being searched|index [^ ]+ is outside an array of length [0-9]+|cast failed: this .+ is not a .+)\n$`)

var digits = regexp.MustCompile(`[0-9]+|0x[0-9a-f]+`)

// javascriptError is how oracle/adamic.mjs reports an exception the program didn't mean: a panic
// carries the program's own words, and anything else is a JavaScript error with its class's name.
var javascriptError = regexp.MustCompile(`^adamic: panic: (RangeError|TypeError|ReferenceError|SyntaxError|Error|InternalError)\b`)

// sanitized is a run's first sanitizer report as one line that stays the same when the same bug is
// at another address: the sanitizer and what it found, without where ("on address 0x...", "(pid=...)")
// and with its numbers as N. A report Node wrote too is the program's own words, not a sanitizer's.
func sanitized(run Run, node Run) string {
	match := sanitizerReport.FindSubmatch(run.Stderr)
	if match == nil || bytes.Contains(node.Stderr, match[0]) {
		return ""
	}
	if match[1] == nil {
		return "UndefinedBehaviorSanitizer: runtime error: " + digits.ReplaceAllString(string(match[3]), "N")
	}
	what := string(match[2])
	for _, where := range []string{" on ", " ("} {
		if index := strings.Index(what, where); index >= 0 {
			what = what[:index]
		}
	}
	return string(match[1]) + ": " + digits.ReplaceAllString(strings.TrimSpace(what), "N")
}

// crashed is how native, or the JavaScript backend, crashed, as Crash's Key names it, or "" when
// neither did. Native's report comes before its signal, since on macOS ASan aborts after every
// report and the report says more. The JavaScript backend runs on Node, unsanitized, so only a
// signal counts there: sanitizer words in its stderr are the program's. A run stopped at its
// deadline was killed by the harness, which is "never finished", not a crash.
func crashed(outcome Outcome) (string, string) {
	if !outcome.Native.TimedOut {
		if report := sanitized(outcome.Native, outcome.Node); report != "" {
			return "native " + report, firstLines(string(outcome.Native.Stderr), 20)
		}
		if outcome.Native.Signal != "" {
			return "native signal: " + outcome.Native.Signal, firstLines(string(outcome.Native.Stderr), 20)
		}
	}
	if !outcome.Backend.TimedOut && outcome.Backend.Signal != "" {
		return "javascript backend signal: " + outcome.Backend.Signal, firstLines(string(outcome.Backend.Stderr), 20)
	}
	return "", ""
}

func (c *Checkout) judge(outcome Outcome, binary string, directory string) Outcome {
	switch {
	case outcome.Node.TimedOut:
		outcome.Verdict, outcome.Key, outcome.Detail = Unfit, "node never finished", string(outcome.Node.Stderr)
		return outcome
	case len(outcome.Node.Stdout) > 1<<20:
		// A program that prints megabytes isn't one a person can read a difference in, and Node can
		// fail writing that much to a pipe before it's read.
		outcome.Verdict, outcome.Key, outcome.Detail = Unfit, "node printed more than a megabyte", ""
		return outcome
	}
	if key, detail := crashed(outcome); key != "" {
		outcome.Verdict, outcome.Key, outcome.Detail = Crash, key, detail
		return outcome
	}
	var differences []string
	nativeDifference := difference(outcome.Node, outcome.Native)
	backendDifference := difference(outcome.Node, outcome.Backend)
	if javascriptError.Match(outcome.Node.Stderr) {
		// Node threw where the library throws (repeat(-1)), or where the checker's types were wrong (a
		// narrowing a call undid). Adamic panics there too, and the oracle holds it to stdout and the
		// exit code, not to V8's words (docs/0.1.md).
		nativeDifference = differenceBesidesStderr(outcome.Node, outcome.Native)
		backendDifference = differenceBesidesStderr(outcome.Node, outcome.Backend)
	}
	if nativeDifference != "" && backendDifference != "" && difference(outcome.Backend, outcome.Native) == "" && outcome.Native.ExitCode == 70 {
		// Both of Adamic's backends stopped at the same place, where Node didn't stop: an inserted
		// check, as long as the panic says it is one and what was printed before it is what Node
		// printed.
		if insertedCheck.Match(outcome.Native.Stderr) && bytes.HasPrefix(outcome.Node.Stdout, outcome.Native.Stdout) {
			outcome.Verdict, outcome.Key, outcome.Detail = Checked, "inserted check", firstLines(string(outcome.Native.Stderr), 1)
			return outcome
		}
	}
	if nativeDifference != "" {
		differences = append(differences, "native "+nativeDifference)
	}
	if backendDifference != "" {
		differences = append(differences, "javascript backend "+backendDifference)
	}
	if len(differences) > 0 {
		sort.Strings(differences)
		outcome.Verdict, outcome.Key = Finding, strings.Join(differences, ", ")
		outcome.Detail = fmt.Sprintf("node:    exit %d, stdout %q, stderr %q\nnative:  exit %d, stdout %q, stderr %q\nbackend: exit %d, stdout %q, stderr %q",
			outcome.Node.ExitCode, tail(outcome.Node.Stdout), outcome.Node.Stderr,
			outcome.Native.ExitCode, tail(outcome.Native.Stdout), outcome.Native.Stderr,
			outcome.Backend.ExitCode, tail(outcome.Backend.Stdout), outcome.Backend.Stderr)
		return outcome
	}
	// Every program that finishes must let go of everything: the same binary again, leak detection on.
	// LeakSanitizer's report here (on Linux), or any other failure (macOS's ASan aborting because it
	// has no leak detection), is a leak, a finding. A report of a bad access is a Crash like any other.
	if outcome.Node.ExitCode == 0 {
		leaked := execute(directory, []string{"ASAN_OPTIONS=detect_leaks=1"}, 20*time.Second, binary)
		if report := sanitized(leaked, outcome.Node); report != "" {
			outcome.Verdict, outcome.Key, outcome.Detail = Crash, "native "+report, firstLines(string(leaked.Stderr), 20)
			return outcome
		}
		if leaked.ExitCode != 0 {
			outcome.Verdict, outcome.Key, outcome.Detail = Finding, "leak", firstLines(string(leaked.Stderr), 20)
			return outcome
		}
	}
	outcome.Verdict = Agreed
	return outcome
}

// compilerRefusal sorts the compiler's refusal: the checker's (the generator's fault), stage 0's
// "not yet" (counted), or anything else, a crash among them, which is a finding.
func compilerRefusal(lowered Run) Outcome {
	stderr := string(lowered.Stderr)
	switch {
	case lowered.TimedOut:
		return Outcome{Verdict: Finding, Key: "compiler never finished", Detail: stderr}
	case strings.Contains(stderr, "can't lower") && strings.Contains(stderr, "yet"):
		// The where is the file and line; the what is the part worth counting.
		what := stderr
		if index := strings.Index(stderr, "can't lower "); index >= 0 {
			what = strings.TrimSpace(stderr[index+len("can't lower "):])
		}
		return Outcome{Verdict: NotYet, Key: "not yet: " + firstLines(what, 1), Detail: stderr}
	case strings.Contains(stderr, "Adamic 0.1 refuses"):
		return Outcome{Verdict: Invalid, Key: "refused: " + firstLines(stderr, 1), Detail: stderr}
	case strings.Contains(stderr, "panic:") || strings.Contains(stderr, "goroutine "):
		return Outcome{Verdict: Finding, Key: "compiler crashed", Detail: firstLines(stderr, 20)}
	case strings.Contains(stderr, ": error TS") || strings.Contains(stderr, " TS"):
		return Outcome{Verdict: Invalid, Key: "checker", Detail: firstLines(stderr, 4)}
	}
	return Outcome{Verdict: Finding, Key: "compiler failed", Detail: firstLines(stderr, 20)}
}

// differenceBesidesStderr is difference, without comparing what was written to stderr.
func differenceBesidesStderr(expected Run, actual Run) string {
	actual.Stderr = expected.Stderr
	return difference(expected, actual)
}

func difference(expected Run, actual Run) string {
	switch {
	case actual.TimedOut:
		return "never finished"
	case expected.ExitCode != actual.ExitCode:
		return "exit code differs"
	case !bytes.Equal(expected.Stdout, actual.Stdout):
		return "stdout differs"
	case !bytes.Equal(expected.Stderr, actual.Stderr):
		return "stderr differs"
	}
	return ""
}

// execute runs a command in its own process group, killed whole at the deadline, so a program that
// loops is stopped rather than orphaned.
// Retain the established 30s compiler and 20s program limits. The measured
// mini-fixture compile/Node/native maxima were 77ms/84ms/6ms; generated-program
// package checks completed in 10.9s, so these limits retain ample margin.
func execute(directory string, environment []string, limit time.Duration, name string, arguments ...string) Run {
	ctx, cancel := boundedrun.WithTimeout(context.Background(), limit)
	defer cancel()
	command := boundedrun.CommandContext(ctx, name, arguments...)
	defer boundedrun.Kill(command.Cmd)
	command.Dir = directory
	if profile := coverageProfile(); profile != "" {
		environment = append(append([]string{}, environment...), "LLVM_PROFILE_FILE="+profile)
	}
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	run := Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), TimedOut: ctx.Err() != nil}
	if run.TimedOut {
		run.Stderr = append(run.Stderr, []byte(fmt.Sprintf("child %s: deadline exceeded; process group killed\n", name))...)
	}
	var exitError *exec.ExitError
	switch {
	case err == nil || errors.As(err, &exitError):
		run.ExitCode = command.ProcessState.ExitCode()
		if status, ok := command.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			run.Signal = status.Signal().String()
		}
	default:
		run.ExitCode = -1
		run.Stderr = append(run.Stderr, []byte(err.Error())...)
	}
	return run
}

// profiles numbers the runs of a coverage measurement, so each one writes its own .profraw.
var profiles atomic.Uint64

// coverageProfile is where the next run writes its clang coverage profile, when
// verify/coverage/measure.sh asks for one (native.Options.Coverage): a path no other run of this
// process or another uses, under ADAMIC_C_COVERAGE_DIRECTORY. Only a binary built with coverage
// writes it; Node and the compiler ignore it. Empty when nothing is being measured.
func coverageProfile() string {
	directory := os.Getenv("ADAMIC_C_COVERAGE_DIRECTORY")
	if directory == "" || !native.CoverageRequested() {
		return ""
	}
	return filepath.Join(directory, fmt.Sprintf("fuzz-%d-%d-%%p.profraw", os.Getpid(), profiles.Add(1)))
}

func firstLines(text string, count int) string {
	lines := strings.SplitN(text, "\n", count+1)
	if len(lines) > count {
		lines = lines[:count]
	}
	return strings.Join(lines, "\n")
}

// tail is the end of a long output, which is where two runs that drift apart usually differ.
func tail(output []byte) []byte {
	if len(output) > 400 {
		return output[len(output)-400:]
	}
	return output
}
