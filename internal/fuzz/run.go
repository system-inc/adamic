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
	"sync"
	"syscall"
	"time"

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
	// Unfit: the program misbehaved on Node itself (it never ended, or printed megabytes), which is
	// the generator's fault.
	Unfit Verdict = "unfit"
	// Flaked: an execution failure under load disappeared on one exclusive rerun.
	Flaked Verdict = "flaked under load"
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
	retry   bool // execution failures that need one exclusive confirmation
}

// programRuns covers all checkouts in this process, including shrinking and
// reduction. A waiting confirmation blocks new attempts until it has run alone.
var programRuns sync.RWMutex

func confirm(run func() Outcome) Outcome {
	first := func() Outcome {
		programRuns.RLock()
		defer programRuns.RUnlock()
		return run()
	}()
	if first.Verdict != Finding || !first.retry {
		return first
	}
	programRuns.Lock()
	defer programRuns.Unlock()
	second := run()
	if second.Verdict == Agreed {
		first.Verdict = Flaked
		return first
	}
	// The exclusive attempt is authoritative; retain both diagnostics when it
	// produces a different failure rather than losing the initial evidence.
	second.Detail = fmt.Sprintf("first attempt: %s %s\n%s\nexclusive rerun: %s %s\n%s",
		first.Verdict, first.Key, first.Detail, second.Verdict, second.Key, second.Detail)
	return second
}

func runnerDied(run Run) bool { return run.ExitCode < 0 || run.TimedOut }

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
	return confirm(func() Outcome { return c.tryFile(path, directory) })
}

// tryFile performs exactly one attempt; callers hold programRuns throughout it.
func (c *Checkout) tryFile(path string, directory string) Outcome {
	// The C first: the checker's and stage 0's refusals come from here.
	lowered := execute(directory, nil, 30*time.Second, c.adamic, "c", path)
	if lowered.ExitCode != 0 || lowered.TimedOut {
		outcome := compilerRefusal(lowered)
		outcome.retry = outcome.Verdict == Finding
		return outcome
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), lowered.Stdout, 0o644); err != nil {
		return Outcome{Verdict: Finding, Key: "fuzzer", Detail: err.Error()}
	}
	javascript := execute(directory, nil, 30*time.Second, c.adamic, "js", path)
	if javascript.ExitCode != 0 || javascript.TimedOut {
		return Outcome{Verdict: Finding, Key: "javascript backend failed", Detail: string(javascript.Stderr), retry: true}
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
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			detail = err.Error()
		}
		return Outcome{Verdict: Finding, Key: key, Detail: detail, retry: true}
	}

	oracle := filepath.Join(c.Root, "oracle", "node.mjs")
	outcome := Outcome{
		Node:    execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, path),
		Native:  execute(directory, []string{"ASAN_OPTIONS=detect_leaks=0"}, 20*time.Second, binary),
		Backend: execute(directory, nil, 20*time.Second, "node", "--disable-warning=ExperimentalWarning", oracle, filepath.Join(directory, "program.mjs")),
	}
	return c.judge(outcome, binary, directory)
}

// sanitizerReport finds the kind of report a sanitizer wrote, if any: ASan's error name, or UBSan's
// "runtime error" with its words, numbers taken out so the same bug at another address is the same.
var sanitizerReport = regexp.MustCompile(`ERROR: AddressSanitizer: ([a-z-]+)|runtime error: ([^\n]*)|ERROR: LeakSanitizer`)

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
	outcome.retry = runnerDied(outcome.Node) || runnerDied(outcome.Native) || runnerDied(outcome.Backend)
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
	if match := sanitizerReport.FindSubmatch(outcome.Native.Stderr); match != nil {
		kind := string(match[1])
		if kind == "" && match[2] != nil {
			kind = "undefined behavior: " + digits.ReplaceAllString(string(match[2]), "N")
		}
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
	// Every program that finishes must let go of everything: the same binary again, leak detection on.
	if outcome.Node.ExitCode == 0 {
		leaked := execute(directory, []string{"ASAN_OPTIONS=detect_leaks=1"}, 20*time.Second, binary)
		if leaked.ExitCode != 0 || leaked.TimedOut {
			outcome.retry = outcome.retry || runnerDied(leaked)
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

// execute limits child CPU time, with a wall backstop for blocked children.
// Cancellation kills the whole process group so descendants are not orphaned.
func execute(directory string, environment []string, limit time.Duration, name string, arguments ...string) Run {
	path, err := exec.LookPath(name)
	if err != nil {
		return Run{ExitCode: -1, Stderr: []byte(err.Error())}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// Set both limits so children that ignore SIGXCPU (including Go programs)
	// are still stopped by the kernel. Some kernels send SIGKILL at the hard
	// limit; distinguish that CPU death using the child's recorded CPU usage.
	// The portable shell interface accepts whole seconds, rounded up.
	seconds := limit / time.Second
	if limit%time.Second != 0 {
		seconds++
	}
	// The soft limit is set first (a hard limit below the current soft one is refused), and the hard
	// limit sits one second past it, so a child that ignores SIGXCPU is killed a
	// full second after its budget, and the SIGKILL check below compares against the soft limit:
	// rusage is tick-sampled and can read a hair under the hard limit the kernel enforced
	// (TestExecuteCPULimit/infinite on 6f16a169's whole gate, "TimedOut:false").
	script := fmt.Sprintf(`ulimit -S -t %d && ulimit -H -t %d || exit; exec "$0" "$@"`, seconds, seconds+1)
	command := exec.CommandContext(ctx, "/bin/sh", append([]string{"-c", script, path}, arguments...)...)
	command.Dir = directory
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	run := Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), TimedOut: ctx.Err() != nil}
	var exitError *exec.ExitError
	switch {
	case err == nil || errors.As(err, &exitError):
		run.ExitCode = command.ProcessState.ExitCode()
		if status, ok := command.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
			if status.Signal() == syscall.SIGXCPU || (status.Signal() == syscall.SIGKILL && cpu >= seconds*time.Second) {
				run.TimedOut = true
			}
		}
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
