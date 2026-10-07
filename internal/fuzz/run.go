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
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
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
	// tsanRuntime is set when this machine can build a ThreadSanitizer runtime. Empty means the
	// runner still compares one thread with the default, and says why TSan is missing in TSan.
	tsanRuntime string
	// TSan is "ready", or the reason the ThreadSanitizer runtime could not be built.
	TSan string
}

// flags are native.Build's for a sanitized build, what the oracle compiles with. They're this tree's,
// even for a checkout of another commit: the C and the runtime are the checkout's, the flags aren't.
var flags = native.Flags(native.Options{Sanitize: true})

// Prepare builds a checkout's adamic command into directory and gets its cached runtime library.
// The cache is shared with native.Build, while another checkout keeps its own runtime bytes.
func Prepare(root string, directory string) (*Checkout, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	checkout := &Checkout{Root: root, directory: directory, adamic: filepath.Join(directory, "adamic")}
	build := exec.Command("go", "build", "-o", checkout.adamic, "./cmd/adamic")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("fuzz: building adamic in %s: %w\n%s", root, err, output)
	}
	checkout.runtime, err = native.RuntimeLibrary(filepath.Join(root, "internal", "native", "runtime"), native.Options{Sanitize: true})
	if err != nil {
		return nil, fmt.Errorf("fuzz: compiling the runtime: %w", err)
	}
	// TSan is a separate build. A machine without it still runs; the report says so.
	tsanRuntime, tsanErr := native.RuntimeLibrary(filepath.Join(root, "internal", "native", "runtime"), native.Options{ThreadSanitize: true})
	if tsanErr != nil {
		checkout.TSan = tsanErr.Error()
	} else {
		checkout.tsanRuntime = tsanRuntime
		checkout.TSan = "ready"
	}
	return checkout, nil
}

// Run is one execution's observable behavior, as the oracle compares it.
type Run struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	TimedOut bool
}

// Verdict is what came of trying one program.
type Verdict string

const (
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
	// Refused: the program was built to break a parallel proof, and the compiler named that path.
	Refused Verdict = "refused"
	// Unfit: the program misbehaved on Node itself (it never ended, or printed megabytes), which is
	// the generator's fault.
	Unfit Verdict = "unfit"
	// Finding: a disagreement, a sanitizer report, a leak, a compiler crash, or C that clang refused.
	Finding Verdict = "finding"
)

// Outcome is a verdict with what it rests on. Key names the kind of finding, so shrinking can keep a
// candidate only when it still fails the same way.
type Outcome struct {
	Verdict Verdict
	Key     string
	Detail  string
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
	// The refusal a program expects, and whether it reaches a task boundary, are in its comments.
	contents, err := os.ReadFile(path)
	if err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	source := string(contents)
	// The C first: the checker's and stage 0's refusals come from here.
	lowered := execute(directory, nil, 30*time.Second, c.adamic, "c", path)
	expected := parallelRefusal(source)
	moveWhat, moveFix := moveRefusal(source)
	if (moveWhat == "") != (moveFix == "") || (moveWhat != "" && moveFix != "return it through the results" && moveFix != "don't use it after the parallelMap") {
		return Outcome{Verdict: Finding, Key: "invalid move expectation"}
	}
	if lowered.ExitCode != 0 || lowered.TimedOut {
		refusal := compilerRefusal(lowered)
		if moveWhat != "" {
			return judgeMoveRefusal(moveWhat, moveFix, lowered)
		}
		if expected == "" {
			return refusal
		}
		return judgeParallelRefusal(expected, refusal, lowered)
	}
	if expected != "" || moveWhat != "" {
		if moveWhat != "" {
			expected = moveWhat + "; " + moveFix
		}
		return Outcome{Verdict: Finding, Key: "compiler accepted a program it must refuse", Detail: "expected a refusal containing " + expected}
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), lowered.Stdout, 0o644); err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	javascript := execute(directory, nil, 30*time.Second, c.adamic, "js", path)
	if javascript.ExitCode != 0 || javascript.TimedOut {
		return Outcome{Verdict: Finding, Key: "javascript backend failed", Detail: string(javascript.Stderr)}
	}
	if err := os.WriteFile(filepath.Join(directory, "program.mjs"), javascript.Stdout, 0o644); err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	binary := filepath.Join(directory, "program")
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(c.runtime), "-o", binary, filepath.Join(directory, "main.c"))
	arguments = append(arguments, native.RuntimeLinkFlags(c.runtime)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		key := "clang refused the C"
		if warning := clangWarning.FindSubmatch(output); warning != nil {
			key += ": " + string(warning[1])
		}
		return Outcome{Verdict: Finding, Key: key, Detail: firstLines(string(output), 12)}
	}

	oracle := filepath.Join(c.Root, "oracle", "node.mjs")
	outcome := Outcome{
		Node:    execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, path),
		Native:  executeIsolated(directory, nativeEnvironment(""), 20*time.Second, binary),
		Backend: execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, filepath.Join(directory, "program.mjs")),
	}
	judged := c.judge(outcome, binary, directory)
	if judged.Verdict != Agreed && judged.Verdict != Checked {
		return judged
	}
	if !strings.Contains(source, "parallelMap(") {
		return judged
	}
	return c.parallelRuns(judged, binary, directory)
}

// sanitizerReport finds the kind of report a sanitizer wrote, if any: ASan's error name, UBSan's
// "runtime error" with its words, TSan's race, or a leak. Numbers are taken out by sanitizerKind so
// the same bug at another address is the same.
var sanitizerReport = regexp.MustCompile(`ERROR: AddressSanitizer: ([a-z-]+)|runtime error: ([^\n]*)|WARNING: ThreadSanitizer: ([^\n]+)|ERROR: ThreadSanitizer: ([^\n]+)|ERROR: LeakSanitizer`)

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

func (c *Checkout) judge(outcome Outcome, binary string, directory string) Outcome {
	switch {
	case outcome.Node.TimedOut:
		outcome.Verdict, outcome.Key, outcome.Detail = Unfit, "node never finished", ""
		return outcome
	case len(outcome.Node.Stdout) > 1<<20:
		// A program that prints megabytes isn't one a person can read a difference in, and Node can
		// fail writing that much to a pipe before it's read.
		outcome.Verdict, outcome.Key, outcome.Detail = Unfit, "node printed more than a megabyte", ""
		return outcome
	}
	if kind := sanitizerKind(outcome.Native.Stderr); kind != "" {
		outcome.Verdict, outcome.Key, outcome.Detail = Finding, "sanitizer: "+kind, firstLines(string(outcome.Native.Stderr), 20)
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
	// Every program that finishes must let go of everything through the shared platform check.
	// Threads stay at the default here; the one-thread leak check is parallelRuns' job.
	if outcome.Node.ExitCode == 0 {
		report, err := c.leaks(binary, directory, "")
		if err != nil {
			report = err.Error()
		}
		if report != "" {
			outcome.Verdict, outcome.Key, outcome.Detail = Finding, "leak", firstLines(report, 20)
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
		return Outcome{Verdict: Finding, Key: "compiler never finished"}
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

// judgeParallelRefusal checks that a program built to be illegal was refused, and that the message
// names the path the generator broke. Acceptance is a finding: the proof did not fire.
func judgeParallelRefusal(expected string, refusal Outcome, lowered Run) Outcome {
	stderr := string(lowered.Stderr)
	if lowered.TimedOut || strings.Contains(stderr, "panic:") || strings.Contains(stderr, "goroutine ") {
		return refusal
	}
	if strings.Contains(stderr, "can't lower") && strings.Contains(stderr, "yet") {
		return Outcome{Verdict: Finding, Key: "expected a refusal, stage 0 said not yet", Detail: firstLines(stderr, 6)}
	}
	if strings.Contains(stderr, expected) {
		return Outcome{Verdict: Refused, Key: expected, Detail: firstLines(stderr, 4)}
	}
	return Outcome{Verdict: Finding, Key: "refusal did not name the path", Detail: "expected " + expected + "\n" + firstLines(stderr, 8)}
}

// parallelRuns holds an accepted parallel program to one thread, the default pool, and ThreadSanitizer
// when that runtime built. Any of those disagreeing with Node, or with each other, is a finding.
func (c *Checkout) parallelRuns(outcome Outcome, binary string, directory string) Outcome {
	one := executeIsolated(directory, nativeEnvironment("1"), 20*time.Second, binary)
	if kind := sanitizerKind(one.Stderr); kind != "" {
		return Outcome{Verdict: Finding, Key: "sanitizer: " + kind, Detail: firstLines(string(one.Stderr), 20)}
	}
	reference := outcome.Node
	if outcome.Verdict == Checked {
		// An inserted check already disagrees with Node. The thread counts still have to agree with each other.
		reference = outcome.Native
	} else if diff := difference(outcome.Node, one); diff != "" {
		return threadFinding("ADAMIC_THREADS=1 "+diff, outcome.Node, one)
	}
	if diff := difference(outcome.Native, one); diff != "" {
		return threadFinding("threads differ: "+diff, outcome.Native, one)
	}
	if outcome.Verdict == Agreed && outcome.Node.ExitCode == 0 {
		report, err := c.leaks(binary, directory, "1")
		if err != nil {
			report = err.Error()
		}
		if report != "" {
			return Outcome{Verdict: Finding, Key: "leak at ADAMIC_THREADS=1", Detail: firstLines(report, 20)}
		}

	}
	if c.tsanRuntime == "" {
		return outcome
	}
	tsanBinary := filepath.Join(directory, "program-tsan")
	if err := c.link(tsanBinary, directory, native.Options{ThreadSanitize: true}, c.tsanRuntime); err != nil {
		return Outcome{Verdict: Finding, Key: "clang refused the TSan build", Detail: err.Error()}
	}
	for _, threads := range []string{"", "1"} {
		label := "default"
		if threads == "1" {
			label = "ADAMIC_THREADS=1"
		}
		observed := executeIsolated(directory, tsanEnvironment(threads), 45*time.Second, tsanBinary)
		if kind := sanitizerKind(observed.Stderr); kind != "" {
			return Outcome{Verdict: Finding, Key: "sanitizer: " + kind + " (" + label + ")", Detail: firstLines(string(observed.Stderr), 20)}
		}
		if observed.TimedOut {
			return Outcome{Verdict: Finding, Key: "thread sanitizer never finished (" + label + ")"}
		}
		if reference.ExitCode != observed.ExitCode || !bytes.Equal(reference.Stdout, observed.Stdout) {
			return threadFinding("thread sanitizer "+label+" disagrees", reference, observed)
		}
		if len(observed.Stderr) != 0 && !bytes.Equal(reference.Stderr, observed.Stderr) {
			return threadFinding("thread sanitizer "+label+" stderr differs", reference, observed)
		}
	}
	return outcome
}

func (c *Checkout) link(binary string, directory string, options native.Options, runtime string) error {
	arguments := append(append([]string{}, native.Flags(options)...), "-I", filepath.Dir(runtime), "-o", binary, filepath.Join(directory, "main.c"))
	arguments = append(arguments, native.RuntimeLinkFlags(runtime)...)
	arguments = append(arguments, "-lm")
	output, err := exec.Command("clang", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", firstLines(string(output), 12))
	}
	return nil
}

func threadFinding(key string, expected Run, actual Run) Outcome {
	return Outcome{
		Verdict: Finding,
		Key:     key,
		Detail: fmt.Sprintf("expected: exit %d, stdout %q, stderr %q\nactual:   exit %d, stdout %q, stderr %q",
			expected.ExitCode, tail(expected.Stdout), expected.Stderr,
			actual.ExitCode, tail(actual.Stdout), actual.Stderr),
	}
}

func sanitizerKind(stderr []byte) string {
	match := sanitizerReport.FindSubmatch(stderr)
	if match == nil {
		return ""
	}
	switch {
	case len(match) > 1 && len(match[1]) > 0:
		return string(match[1])
	case len(match) > 2 && len(match[2]) > 0:
		return "undefined behavior: " + digits.ReplaceAllString(string(match[2]), "N")
	case len(match) > 3 && len(match[3]) > 0:
		return "thread: " + strings.TrimSpace(digits.ReplaceAllString(string(match[3]), "N"))
	case len(match) > 4 && len(match[4]) > 0:
		return "thread: " + strings.TrimSpace(digits.ReplaceAllString(string(match[4]), "N"))
	default:
		return "leak"
	}
}

// Native comparison disables Linux leak detection; the shared check runs it separately.
func nativeEnvironment(threads string) []string {
	var extra []string
	if runtime.GOOS == "linux" {
		extra = append(extra, "ASAN_OPTIONS=detect_leaks=0")
	}
	if threads != "" {
		extra = append(extra, "ADAMIC_THREADS="+threads)
	}
	return isolatedEnvironment(extra)
}

// Preserve the fuzzer's isolated environment, deadline and one-worker witness in every leak run.
func (c *Checkout) leaks(binary, directory, threads string) (string, error) {
	code, err := os.ReadFile(filepath.Join(directory, "main.c"))
	if err != nil {
		return "", err
	}
	timedOut := false
	report, err := leakcheck.Check(leakcheck.Program{
		C: string(code), Sanitized: binary, Counted: filepath.Join(directory, "program-counted"),
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			extra := append([]string{}, environment...)
			if threads != "" {
				extra = append(extra, "ADAMIC_THREADS="+threads)
			}
			result := executeIsolated(directory, isolatedEnvironment(extra), 20*time.Second, name, arguments...)
			timedOut = timedOut || result.TimedOut
			return leakcheck.Run{Stdout: result.Stdout, Stderr: result.Stderr, ExitCode: result.ExitCode}
		},
	})
	if timedOut {
		return "", fmt.Errorf("leak check never finished")
	}
	return report, err
}

func tsanEnvironment(threads string) []string {
	extra := []string{"TSAN_OPTIONS=halt_on_error=1"}
	if threads != "" {
		extra = append(extra, "ADAMIC_THREADS="+threads)
	}
	return isolatedEnvironment(extra)
}

// isolatedEnvironment drops a parent shell's thread and sanitizer settings, then adds extra.
// The default pool is the unset variable, not whatever the fuzzer itself was launched with.
func isolatedEnvironment(extra []string) []string {
	var environment []string
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ADAMIC_THREADS=") || strings.HasPrefix(entry, "ASAN_OPTIONS=") || strings.HasPrefix(entry, "TSAN_OPTIONS=") {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment, extra...)
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
// loops is stopped rather than orphaned. environment is appended to the parent environment.
func execute(directory string, environment []string, limit time.Duration, name string, arguments ...string) Run {
	return executeIsolated(directory, append(os.Environ(), environment...), limit, name, arguments...)
}

// executeIsolated is execute with exactly the environment it is given.
func executeIsolated(directory string, environment []string, limit time.Duration, name string, arguments ...string) Run {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = directory
	command.Env = environment
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	run := Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), TimedOut: ctx.Err() != nil}
	var exitError *exec.ExitError
	switch {
	case err == nil || errors.As(err, &exitError):
		run.ExitCode = command.ProcessState.ExitCode()
	default:
		run.ExitCode = -1
		run.Stderr = append(run.Stderr, []byte(err.Error())...)
	}
	return run
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
