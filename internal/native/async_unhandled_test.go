package native_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

var asyncUnhandledCases = []struct{ name, stdout, message string }{
	{"unhandled", "", "rejected"},
	{"chained", "", "chained"},
	{"handled-later", "handled\n", ""},
	{"timer-checkpoint", "timer\n", "timer rejection"},
	{"first-only", "", "first"},
}

func asyncUnhandledBuild(t *testing.T, variant string, options native.Options) string {
	t.Helper()
	data, err := os.ReadFile("testdata/unhandled_rejection/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	code := string(data)
	binary := filepath.Join(t.TempDir(), "unhandled-rejection")
	if variant == "control" {
		err = native.Build(code, binary, options)
	} else {
		directory := t.TempDir()
		entries, readErr := os.ReadDir("runtime")
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, entry := range entries {
			if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
				continue
			}
			contents, readErr := os.ReadFile(filepath.Join("runtime", entry.Name()))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if entry.Name() == "async.c" {
				var before, after string
				count := 1
				switch variant {
				case "skip-check":
					before, after, count = "\n    check_unhandled();", "\n    /* mutant: omit the rejection checkpoint */", 2
				case "no-retain":
					before, after = "*entry = (pending){adamic_retain(promise), NULL};", "*entry = (pending){promise, NULL};"
				case "no-handled":
					before, after = "    promise->handled = true;", "    /* mutant: leave an awaited promise unhandled */"
				default:
					t.Fatal("unknown variant", variant)
				}
				if strings.Count(string(contents), before) != count {
					t.Fatal("mutant seam changed", variant)
				}
				contents = []byte(strings.ReplaceAll(string(contents), before, after))
			}
			if err := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0600); err != nil {
				t.Fatal(err)
			}
		}
		library, buildErr := native.RuntimeLibraryForSource(directory, code, options)
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		err = native.BuildWithRuntime(code, binary, options, library)
	}
	if err != nil {
		t.Fatal(err)
	}
	return binary
}

func asyncUnhandledRun(t *testing.T, binary, mode string) leakcheck.Run {
	t.Helper()
	got, expired := timerRunWithin(5*time.Second, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mode)
	if expired {
		t.Fatal("rejection harness missed its deadline")
	}
	return got
}

// Stopping cases keep the runtime's exact uncaught exit. Only finishing cases
// balance their counted owners; a panic deliberately leaves by _exit.
func asyncUnhandledDifference(mode string, counted bool, got leakcheck.Run) string {
	if bytes.Contains(got.Stderr, []byte("Sanitizer")) || bytes.Contains(got.Stderr, []byte("runtime error:")) {
		return "sanitizer report"
	}
	if counted {
		if report := leakcheck.Unbalanced(got); report != "" {
			return report
		}
		got.Stderr = leakcheck.CountsLine.ReplaceAll(got.Stderr, nil)
	}
	for _, each := range asyncUnhandledCases {
		if each.name != mode {
			continue
		}
		exit, stderr := 0, ""
		if each.message != "" {
			exit, stderr = 70, "adamic: panic: Error: "+each.message+"\n"
		}
		if got.ExitCode != exit || string(got.Stdout) != each.stdout || string(got.Stderr) != stderr {
			return "exit or output differs"
		}
		return ""
	}
	return "unknown mode"
}

// Lowering now refuses the original void views. These runtime cases keep the
// queue-only owner, an awaited inner and dropped outer, and the timer checkpoint.
func TestAsyncUnhandledRejections(t *testing.T) {
	t.Parallel()
	for _, variant := range []struct {
		name    string
		options native.Options
	}{
		{"sanitized", native.Options{Sanitize: true}},
		{"release", native.Options{Release: true}},
	} {
		binary := asyncUnhandledBuild(t, "control", variant.options)
		for _, each := range asyncUnhandledCases {
			if each.message == "" {
				continue
			}
			t.Run(each.name+"/"+variant.name, func(t *testing.T) {
				got := asyncUnhandledRun(t, binary, each.name)
				if difference := asyncUnhandledDifference(each.name, false, got); difference != "" {
					t.Fatalf("%s: %+v", difference, got)
				}
			})
		}
	}
	t.Run("handled-later/sanitized", func(t *testing.T) {
		binary := asyncUnhandledBuild(t, "control", native.Options{Sanitize: true, Count: true})
		got := asyncUnhandledRun(t, binary, "handled-later")
		if difference := asyncUnhandledDifference("handled-later", true, got); difference != "" {
			t.Fatalf("%s: %+v", difference, got)
		}
		t.Log(strings.TrimSpace(string(got.Stderr)))
	})
}

// The checkpoint mutant reds TestAsyncUnhandledRejections/unhandled/sanitized
// and /chained/sanitized. The retain mutant reds /unhandled/sanitized with a
// sanitizer use-after-free. The handled-bit mutant reds /handled-later/sanitized.
func TestAsyncUnhandledRejectionMutants(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"skip-check", "no-retain", "no-handled"} {
		t.Run(variant, func(t *testing.T) {
			counted := variant == "no-handled"
			binary := asyncUnhandledBuild(t, variant, native.Options{Sanitize: true, Count: counted})
			modes := []string{"unhandled"}
			if variant == "skip-check" {
				modes = append(modes, "chained")
			}
			if counted {
				modes = []string{"handled-later"}
			}
			for _, mode := range modes {
				got := asyncUnhandledRun(t, binary, mode)
				if asyncUnhandledDifference(mode, counted, got) == "" {
					t.Fatalf("%s survived %s: %+v", variant, mode, got)
				}
				if variant == "no-retain" && !strings.Contains(string(got.Stderr), "heap-use-after-free") {
					t.Fatalf("retain mutant must use a freed promise: %+v", got)
				}
				if variant == "skip-check" && got.ExitCode != 0 {
					t.Fatalf("checkpoint mutant must return silently: %+v", got)
				}
				if variant == "no-handled" && (got.ExitCode != 70 || !strings.Contains(string(got.Stderr), "adamic: panic: Error: handled\n")) {
					t.Fatalf("handled mutant must falsely report: %+v", got)
				}
				t.Logf("TestAsyncUnhandledRejections/%s/sanitized reds with %s", mode, variant)
			}
		})
	}
}
