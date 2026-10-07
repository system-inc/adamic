package metamorphic

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/fuzz"
)

// Checkout is a compiler under test, driven through its own command line the way the oracle builds a
// fixture (TestNativeAgreesWithNode): the source on Node, native under the sanitizers, native as a
// user's release build, the JavaScript backend on Node, and, for a program that finishes, the leak
// check of its platform. Every build is the checkout's own adamic build, with the checkout's runtime
// and flags, so a fault anywhere in them is the checkout's to show.
type Checkout struct {
	// Root is the checkout, whose oracle/node.mjs runs the source and the JavaScript backend.
	Root string
	// Adamic is the adamic command built from Root.
	Adamic string
}

// Deadlines: the oracle gives every command a minute.
const (
	programLimit = time.Minute
	buildLimit   = 3 * time.Minute
)

// Result is what one program came to on a checkout: fuzz's verdicts, judged by fuzz's judge, with
// the oracle's two further checks, the release build against the sanitized one and the leak check.
type Result struct {
	Verdict fuzz.Verdict `json:"verdict"`
	Key     string       `json:"key,omitempty"`
	Detail  string       `json:"detail,omitempty"`
	// Line is a disagreement's first line: the first line of output that differs, or the exit code or
	// stderr when the output doesn't.
	Line string `json:"line,omitempty"`
}

// Node runs a program's source on Node: the truth.
func (c Checkout) Node(path string, directory string) fuzz.Run {
	return fuzz.Execute(directory, nil, programLimit, "node", "--disable-warning=ExperimentalWarning", filepath.Join(c.Root, "oracle", "node.mjs"), path)
}

// Run builds and runs a program every way the oracle does, in directory, and judges it.
func (c Checkout) Run(path string, directory string) Result {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Result{Verdict: fuzz.Finding, Key: "harness", Detail: err.Error()}
	}
	// The JavaScript first: the checker's and stage 0's refusals come from here.
	emitted := fuzz.Execute(directory, nil, buildLimit, c.Adamic, "js", path)
	if emitted.ExitCode != 0 || emitted.TimedOut {
		refusal := fuzz.Refusal(emitted)
		return Result{Verdict: refusal.Verdict, Key: refusal.Key, Detail: refusal.Detail}
	}
	backendPath := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(backendPath, emitted.Stdout, 0o644); err != nil {
		return Result{Verdict: fuzz.Finding, Key: "harness", Detail: err.Error()}
	}
	sanitizedBinary, releaseBinary, countedBinary := filepath.Join(directory, "sanitized"), filepath.Join(directory, "release"), filepath.Join(directory, "counted")
	for _, build := range []struct {
		name      string
		arguments []string
	}{
		{"sanitized", []string{"build", path, "-o", sanitizedBinary, "--sanitize"}},
		{"release", []string{"build", path, "-o", releaseBinary}},
		{"counted", []string{"build", path, "-o", countedBinary, "--count"}},
	} {
		built := fuzz.Execute(directory, nil, buildLimit, c.Adamic, build.arguments...)
		if built.ExitCode != 0 || built.TimedOut {
			key := "the " + build.name + " build failed"
			if refused := clangError.FindSubmatch(built.Stderr); refused != nil {
				key += ": clang: " + string(numbers.ReplaceAll(refused[1], []byte("N")))
			}
			return Result{Verdict: fuzz.Finding, Key: key, Detail: firstLines(string(built.Stderr), 12)}
		}
	}
	var sanitizedEnvironment []string
	if runtime.GOOS == "linux" {
		sanitizedEnvironment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	runs := fuzz.Outcome{
		Node:    c.Node(path, directory),
		Native:  fuzz.Execute(directory, sanitizedEnvironment, programLimit, sanitizedBinary),
		Backend: fuzz.Execute(directory, nil, programLimit, "node", "--disable-warning=ExperimentalWarning", filepath.Join(c.Root, "oracle", "node.mjs"), backendPath),
	}
	nativeStdout := runs.Native.Stdout
	backendStdout, nodeStdout := runs.Backend.Stdout, runs.Node.Stdout
	if len(runs.Node.Stdout) > 1<<20 {
		// The judge calls a program that prints more than a megabyte unfit, a rule for generated
		// programs; a fixture that prints that much (a sweep) is held to every byte of it by the oracle.
		// Each run's stdout is judged by its digest instead, which is equal exactly when the bytes are.
		runs.Node.Stdout, runs.Native.Stdout, runs.Backend.Stdout = digest(runs.Node.Stdout), digest(runs.Native.Stdout), digest(runs.Backend.Stdout)
	}
	outcome := fuzz.Compare(runs)
	outcome.Node.Stdout, outcome.Native.Stdout, outcome.Backend.Stdout = nodeStdout, nativeStdout, backendStdout
	if outcome.Verdict != fuzz.Agreed && outcome.Verdict != fuzz.Checked {
		result := Result{Verdict: outcome.Verdict, Key: outcome.Key, Detail: outcome.Detail}
		if outcome.Verdict == fuzz.Finding {
			result.Line = firstDifference(outcome.Node, outcome.Native, "native")
			if result.Line == "" {
				result.Line = firstDifference(outcome.Node, outcome.Backend, "javascript backend")
			}
		}
		return result
	}
	// The build a user gets must say exactly what the sanitized one did.
	release := fuzz.Execute(directory, nil, programLimit, releaseBinary)
	if difference := disagreement(outcome.Native, release); difference != "" {
		return Result{Verdict: fuzz.Finding, Key: "release build " + difference, Line: firstDifference(outcome.Native, release, "release"), Detail: fmt.Sprintf("sanitized: exit %d, stdout %q, stderr %q\nrelease:   exit %d, stdout %q, stderr %q",
			outcome.Native.ExitCode, tail(outcome.Native.Stdout), firstLines(string(outcome.Native.Stderr), 4), release.ExitCode, tail(release.Stdout), firstLines(string(release.Stderr), 4))}
	}
	if outcome.Verdict == fuzz.Agreed && outcome.Node.ExitCode == 0 {
		if leaked := c.leaks(directory, sanitizedBinary, countedBinary); leaked != "" {
			return Result{Verdict: fuzz.Finding, Key: "leak", Detail: leaked}
		}
	}
	return Result{Verdict: outcome.Verdict, Key: outcome.Key}
}

// clangError is clang's first error in a failed build, without where it was; numbers names the
// numbers in it, which a program's other temporaries shift.
var (
	clangError = regexp.MustCompile(`(?m)^[^ ]+:[0-9]+:[0-9]+: error: ([^\n]*)`)
	numbers    = regexp.MustCompile(`[0-9]+`)
)

// countsLine is the line a counted build writes last to stderr (runtime/count.c).
var countsLine = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n\z`)

// leaks is the oracle's leak check: on macOS the counted build's counts and then leaks --atExit on
// it, on Linux LeakSanitizer on the sanitized binary. It returns a report, or "".
func (c Checkout) leaks(directory string, sanitized string, counted string) string {
	if runtime.GOOS == "linux" {
		report := fuzz.Execute(directory, []string{"ASAN_OPTIONS=detect_leaks=1"}, programLimit, sanitized)
		if report.ExitCode == 0 {
			return ""
		}
		return fmt.Sprintf("LeakSanitizer: exit %d\n%s", report.ExitCode, firstLines(string(report.Stderr), 20))
	}
	run := fuzz.Execute(directory, nil, programLimit, counted)
	match := countsLine.FindSubmatch(run.Stderr)
	if run.ExitCode != 0 || match == nil {
		return fmt.Sprintf("the counted build didn't finish with its counts: exit %d, stderr %q", run.ExitCode, tail(run.Stderr))
	}
	allocations, _ := strconv.ParseInt(string(match[1]), 10, 64)
	frees, _ := strconv.ParseInt(string(match[2]), 10, 64)
	regions, _ := strconv.ParseInt(string(match[6]), 10, 64)
	if allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d, in regions %d)", allocations-frees-regions, allocations, frees, regions)
	}
	report := fuzz.Execute(directory, nil, programLimit, "leaks", "--atExit", "--", counted)
	if report.ExitCode != 0 {
		return "leaks --atExit:\n" + firstLines(string(report.Stdout), 30)
	}
	return ""
}

// disagreement says how two runs differ, as the oracle says it, or "".
func disagreement(expected fuzz.Run, actual fuzz.Run) string {
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

// Same says whether two runs on Node printed exactly the same things and ended the same way: the test
// a variant has to pass to be kept.
func Same(original fuzz.Run, variant fuzz.Run) bool {
	return !original.TimedOut && !variant.TimedOut && disagreement(original, variant) == ""
}

func firstLines(text string, count int) string {
	lines := bytes.SplitN([]byte(text), []byte("\n"), count+1)
	if len(lines) > count {
		lines = lines[:count]
	}
	return string(bytes.Join(lines, []byte("\n")))
}

func tail(output []byte) []byte {
	if len(output) > 400 {
		return output[len(output)-400:]
	}
	return output
}

// digest is an output's length and SHA-256, as text.
func digest(output []byte) []byte {
	return fmt.Appendf(nil, "%d bytes, sha256 %x\n", len(output), sha256.Sum256(output))
}

// firstDifference is the first thing a run differs from the expected one on, as one line: an output
// line by its number, then the exit code, then stderr's first line.
func firstDifference(expected fuzz.Run, actual fuzz.Run, name string) string {
	want, got := strings.Split(string(expected.Stdout), "\n"), strings.Split(string(actual.Stdout), "\n")
	for index := range max(len(want), len(got)) {
		var wanted, gotten string
		if index < len(want) {
			wanted = want[index]
		}
		if index < len(got) {
			gotten = got[index]
		}
		if index >= len(want) || index >= len(got) || wanted != gotten {
			return fmt.Sprintf("%s stdout line %d: expected %q, %s %q", name, index+1, clip(wanted), name, clip(gotten))
		}
	}
	switch {
	case actual.TimedOut:
		return name + " never finished"
	case expected.ExitCode != actual.ExitCode:
		return fmt.Sprintf("%s exit: expected %d, %s %d: %s", name, expected.ExitCode, name, actual.ExitCode, firstLines(string(actual.Stderr), 1))
	case !bytes.Equal(expected.Stderr, actual.Stderr):
		return fmt.Sprintf("%s stderr: expected %q, %s %q", name, firstLines(string(expected.Stderr), 1), name, firstLines(string(actual.Stderr), 1))
	}
	return ""
}

func clip(line string) string {
	if len(line) > 160 {
		return line[:160] + "..."
	}
	return line
}
