package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Pending work must stop at the first uncaught throw, including a rejection Node promotes to one.
func TestOracleUncaughtExitStopsPendingTimers(t *testing.T) {
	t.Parallel()
	// Reads: internal/oracle/testdata/uncaught_exit/*.mjs
	for _, name := range []string{"two_timers", "top_level", "promise_job"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/uncaught_exit", name+".mjs"))
			if err != nil {
				t.Fatal(err)
			}
			got := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if !oracleUncaughtExitMatches(got) {
				t.Fatalf("uncaught throw must stop pending timers: %+v", got)
			}
			plain := execute(t, "node", path)
			if plain.exitCode != 1 || len(plain.stdout) != 0 || !strings.Contains(string(plain.stderr), "Error: uncaught exit\n") {
				t.Fatalf("plain Node must stop pending timers: %+v", plain)
			}
			t.Logf("plain Node: exit %d, stderr %q", plain.exitCode, plain.stderr)
		})
	}
}

func oracleUncaughtExitMatches(got run) bool {
	return got.exitCode == 70 && len(got.stdout) == 0 && string(got.stderr) == "adamic: panic: Error: uncaught exit\n"
}

// Restore the old handler in a private bundle: the same two-timer assertion must reject its output.
func TestOracleUncaughtExitCodeMutantKeepsTimerAlive(t *testing.T) {
	t.Parallel()
	// Reads: oracle/*.mjs
	// Reads: internal/oracle/testdata/uncaught_exit/two_timers.mjs
	directory := t.TempDir()
	for _, name := range []string{"node.mjs", "adamic.mjs", "json_encode.mjs", "json_decode.mjs"} {
		data, err := os.ReadFile(filepath.Join(repository, "oracle", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "adamic.mjs" {
			anchor := "process.stderr.write(`adamic: panic: ${message}\\n`);\n\tprocess.exit(70);\n});"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("uncaught exit mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), anchor, strings.Replace(anchor, "process.exit(70)", "process.exitCode = 70", 1), 1))
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/uncaught_exit/two_timers.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	got := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(directory, "node.mjs"), path)
	if oracleUncaughtExitMatches(got) {
		t.Fatal("exitCode mutant survived the two-timer assertion")
	}
	if got.exitCode != 70 || string(got.stdout) != "pending timer\n" || string(got.stderr) != "adamic: panic: Error: uncaught exit\n" {
		t.Fatalf("mutant must fail only because the pending timer ran: %+v", got)
	}
	t.Log("caught by TestOracleUncaughtExitStopsPendingTimers/two_timers: pending timer printed")
}
