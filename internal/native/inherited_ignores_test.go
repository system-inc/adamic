package native

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// inheritedIgnoresMutants each break one promise of inherited_signals_impl.h's Linux road, and name the case
// that must catch it.
var inheritedIgnoresMutants = []struct{ name, before, after, red string }{
	// A signal the mask names is left ignored.
	{"skip_masked", "if ((mask >> (signal_number - 1) & 1) != 0 && signal_number != SIGKILL", "if ((mask >> (signal_number - 1) & 1) != 0 && signal_number != SIGALRM && signal_number != SIGKILL", "proc"},
	// A file that can't be opened reads as nothing ignored.
	{"open_failure_as_none", "if (descriptor < 0) {\n\t\treturn false;", "if (descriptor < 0) {\n\t\t*mask = 0;\n\t\treturn true;", "missing"},
	// A file with no SigIgn line reads as nothing ignored.
	{"missing_line_as_none", "if (line == NULL) {\n\t\treturn false;", "if (line == NULL) {\n\t\t*mask = 0;\n\t\treturn true;", "no_line"},
	// A mask is read up to its first character that isn't hex, or its sixteenth digit.
	{"partial_hex", "if (nibble < 0 || digits == 16) {\n\t\t\treturn false;", "if (nibble < 0 || digits == 16) {\n\t\t\tbreak;", "not_hex"},
	// A mask that can't be read resets nothing.
	{"fallback_none", "\tinherited_ignores_sweep(to_default);\n}\n", "\t(void)to_default;\n}\n", "missing"},
}

// inheritedIgnoresBuild compiles testdata/inherited_ignores.c with inherited_signals_impl.h's text in place of
// its include, after the mutant's change when one is named.
func inheritedIgnoresBuild(t *testing.T, directory, name, before, after string, options native_runtime.Options) string {
	t.Helper()
	header, err := fs.ReadFile(native_runtime.Sources, "c/inherited_signals_impl.h")
	if err != nil {
		t.Fatal(err)
	}
	harness, err := os.ReadFile(filepath.Join("testdata", "inherited_ignores.c"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(header)
	if before != "" {
		if strings.Count(text, before) != 1 {
			t.Fatalf("%s: the mutant's seam isn't in inherited_signals_impl.h once", name)
		}
		text = strings.Replace(text, before, after, 1)
	}
	const include = `#include "inherited_signals_impl.h"`
	if strings.Count(string(harness), include) != 1 {
		t.Fatal("inherited_ignores.c must include inherited_signals_impl.h once")
	}
	binary := filepath.Join(directory, name)
	if err := Build(strings.Replace(string(harness), include, text, 1), binary, options); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return binary
}

type inheritedIgnoresResult struct {
	before, after []int
	calls         int
}

func inheritedIgnoresRun(t *testing.T, binary, status string) inheritedIgnoresResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, binary, status).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", filepath.Base(binary), status, err, output)
	}
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "ignored before:") || !strings.HasPrefix(lines[1], "ignored after:") || !strings.HasPrefix(lines[2], "calls ") {
		t.Fatalf("%s: unexpected output %q", filepath.Base(binary), output)
	}
	numbers := func(line, prefix string) []int {
		var result []int
		for _, field := range strings.Fields(strings.TrimPrefix(line, prefix)) {
			number, err := strconv.Atoi(field)
			if err != nil {
				t.Fatalf("%q: %v", line, err)
			}
			result = append(result, number)
		}
		return result
	}
	calls, err := strconv.Atoi(strings.TrimPrefix(lines[2], "calls "))
	if err != nil {
		t.Fatal(err)
	}
	return inheritedIgnoresResult{numbers(lines[0], "ignored before:"), numbers(lines[1], "ignored after:"), calls}
}

// inheritedIgnoresMask is the SigIgn mask naming signals, as the kernel prints it.
func inheritedIgnoresMask(signals []int) string {
	var mask uint64
	for _, signal := range signals {
		mask |= 1 << (signal - 1)
	}
	return fmt.Sprintf("%016x", mask)
}

// inheritedIgnoresCase is one status file the reset is handed, and what it must do with it: reset exactly
// the signals left (the mask road, one sigaction each), or ask every signal and reset all four (the sweep).
type inheritedIgnoresCase struct {
	name   string
	status func(t *testing.T, directory string, before []int) string
	// named are the signals the mask road resets, by their place in before; nil means the sweep.
	named []int
}

func inheritedIgnoresCases() []inheritedIgnoresCase {
	file := func(name, contents string) func(*testing.T, string, []int) string {
		return func(t *testing.T, directory string, _ []int) string {
			path := filepath.Join(directory, name)
			if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
			return path
		}
	}
	status := func(sigIgn string) string {
		return "Name:\tinherited\nState:\tR (running)\nSigPnd:\t0000000000000000\nSigBlk:\t0000000000000000\nSigIgn:\t" + sigIgn + "SigCgt:\t0000000000000000\n"
	}
	masked := func(name string, places []int, upper bool) func(*testing.T, string, []int) string {
		return func(t *testing.T, directory string, before []int) string {
			var signals []int
			for _, place := range places {
				signals = append(signals, before[place])
			}
			mask := inheritedIgnoresMask(signals)
			if upper {
				mask = strings.ToUpper(mask)
			}
			return file(name, status(mask+"\n"))(t, directory, before)
		}
	}
	return []inheritedIgnoresCase{
		// The process's own status, as adamic.c reads it.
		{"proc", func(*testing.T, string, []int) string { return "/proc/self/status" }, []int{0, 1, 2, 3}},
		// A mask is followed, not swept past: what it doesn't name stays as it was, the real-time signal's high bit included.
		{"partial", masked("partial", []int{0, 3}, false), []int{0, 3}},
		{"upper", masked("upper", []int{0, 1, 2, 3}, true), []int{0, 1, 2, 3}},
		// Every status that can't be read whole falls back to asking each signal.
		{"missing", func(_ *testing.T, directory string, _ []int) string { return filepath.Join(directory, "absent") }, nil},
		{"unreadable", func(_ *testing.T, directory string, _ []int) string { return directory }, nil},
		{"no_line", file("no_line", "Name:\tinherited\nSigCgt:\t0000000000000000\n"), nil},
		{"not_hex", file("not_hex", status("00000000000g1000\n")), nil},
		{"empty_mask", file("empty_mask", status("\n")), nil},
		{"too_long", file("too_long", status("00000000000000000\n")), nil},
		{"unterminated", file("unterminated", "Name:\tinherited\nSigIgn:\t0000000000001000"), nil},
	}
}

// inheritedIgnoresHolds says what the reset did wrong with one case, or nothing. Off Linux every case is the
// sweep, since only Linux has the mask.
func inheritedIgnoresHolds(c inheritedIgnoresCase, result inheritedIgnoresResult) string {
	if len(result.before) != 4 {
		return fmt.Sprintf("the harness ignored %v before the reset, not four signals", result.before)
	}
	named := c.named
	if goruntime.GOOS != "linux" {
		named = nil
	}
	if named == nil {
		if len(result.after) != 0 {
			return fmt.Sprintf("the sweep left %v ignored", result.after)
		}
		if result.calls < 31 {
			return fmt.Sprintf("a status that can't be read made %d sigaction calls, fewer than one for each signal", result.calls)
		}
		return ""
	}
	reset := map[int]bool{}
	for _, place := range named {
		reset[result.before[place]] = true
	}
	var left []int
	for _, signal := range result.before {
		if !reset[signal] {
			left = append(left, signal)
		}
	}
	if fmt.Sprint(result.after) != fmt.Sprint(left) {
		return fmt.Sprintf("the mask road left %v ignored, want %v", result.after, left)
	}
	if result.calls != len(named) {
		return fmt.Sprintf("the mask road made %d sigaction calls, want %d, one for each signal it names", result.calls, len(named))
	}
	return ""
}

// A program resets what its parent ignored from one read of /proc/self/status's SigIgn mask, a sigaction only
// for each signal it names, and asks every signal whenever the mask can't be read whole (#7z36qn6). The oracle's
// TestSignal*Inherited* rows hold the reset end to end; this holds the mask road's call count, its refusals,
// and the fallback, which no program's run can reach while /proc is there.
func TestInheritedIgnoresReadTheMaskOnce(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for _, build := range []struct {
		name    string
		options native_runtime.Options
	}{{"release", native_runtime.Options{}}, {"asan", native_runtime.Options{Sanitize: true}}} {
		binary := inheritedIgnoresBuild(t, directory, "control-"+build.name, "", "", build.options)
		for _, c := range inheritedIgnoresCases() {
			t.Run(build.name+"/"+c.name, func(t *testing.T) {
				cases := t.TempDir()
				result := inheritedIgnoresRun(t, binary, c.status(t, cases, inheritedIgnoresRun(t, binary, "/proc/self/status").before))
				if failure := inheritedIgnoresHolds(c, result); failure != "" {
					t.Fatal(failure)
				}
			})
		}
	}
}

// Each mutant of the Linux road goes red on the case it names.
func TestInheritedIgnoresMutants(t *testing.T) {
	t.Parallel()
	if goruntime.GOOS != "linux" {
		t.Skip("the mask road is Linux's; elsewhere the reset is the sweep, which TestInheritedIgnoresReadTheMaskOnce holds")
	}
	directory := t.TempDir()
	for _, mutant := range inheritedIgnoresMutants {
		t.Run(mutant.name, func(t *testing.T) {
			binary := inheritedIgnoresBuild(t, directory, mutant.name, mutant.before, mutant.after, native_runtime.Options{})
			for _, c := range inheritedIgnoresCases() {
				if c.name != mutant.red {
					continue
				}
				cases := t.TempDir()
				result := inheritedIgnoresRun(t, binary, c.status(t, cases, inheritedIgnoresRun(t, binary, "/proc/self/status").before))
				failure := inheritedIgnoresHolds(c, result)
				if failure == "" {
					t.Fatalf("%s escaped case %s", mutant.name, c.name)
				}
				t.Logf("caught %s on case %s: %s", mutant.name, c.name, failure)
				return
			}
			t.Fatalf("%s names case %s, which isn't one", mutant.name, mutant.red)
		})
	}
}
