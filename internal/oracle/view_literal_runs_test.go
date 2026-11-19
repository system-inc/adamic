package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// A checked view's literal set of three or more consecutive integers is one range test natively
// (#zpzm9ta), which must answer as the === chain it stands for. Kind is the runs 0..7 and 10..13, so
// each fixture's read of imported.argument.literal is two run tests and no equality. Mutants, in
// testdata/view_literal_runs/mutants.md: the trunc test dropped lets stop_fraction.ts run on, and >=
// written > stops honest.ts at 0.
const viewLiteralRunsDirectory = "internal/oracle/testdata/view_literal_runs/"

func viewLiteralRunsFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, viewLiteralRunsDirectory+name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// viewLiteralRunsEmitted fails unless every check of imported.argument.literal the program emits is
// Kind's two runs, so the fixtures test the run path and not the chain. It counts the upper bounds,
// which neither mutant touches, so each mutant is caught where the program runs.
func viewLiteralRunsEmitted(t *testing.T, program *ir.Program) {
	t.Helper()
	checks := 0
	for line := range strings.SplitSeq(native.C(program), "\n") {
		if !strings.Contains(line, "adamic_view_literal_failure(") || !strings.Contains(line, `"imported.argument.literal"`) {
			continue
		}
		checks++
		if strings.Count(line, " <= ") != 2 || strings.Contains(line, "adamic_same_number(") {
			t.Fatalf("the literal check isn't Kind's two runs: %s", strings.TrimSpace(line))
		}
	}
	if checks == 0 {
		t.Fatal("no literal check of imported.argument.literal emitted")
	}
}

// The honest reads pass at the first run's bounds and inside it, at the second's bounds, and at -0,
// which 0 === -0 admits, each agreeing with Node natively, sanitized and on the JavaScript backend.
func TestViewLiteralRunsHonestReadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	const node = "other\nnode:fs 0\nnode:fs 7\nnode:fs 3\nnode:fs 0\nnode:fs 10\nnode:fs 13\n"
	if truth := onNode(t, viewLiteralRunsFixture(t, "honest.ts")); truth.exitCode != 0 || string(truth.stdout) != node {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, viewLiteralRunsFixture(t, "honest.ts"))
	if err != nil {
		t.Fatal(err)
	}
	viewLiteralRunsEmitted(t, program)
	loomViewsDAgrees(t, viewLiteralRunsDirectory+"honest.ts")
}

// Each lie stops at the read with the view's literal failure and exit 70, as the === chain stopped
// it, after the untagged call has printed what Node prints: one below the first run, one above the
// second, a non-integer inside the first run's bounds, NaN, and a value in the hole between the runs.
// Node, without the checks, prints on.
func TestViewLiteralRunsStopAtTheRead(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, found string }{
		{"stop_below.ts", "-1"},
		{"stop_above.ts", "14"},
		{"stop_fraction.ts", "0.5"},
		{"stop_nan.ts", "NaN"},
		{"stop_hole.ts", "8"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := viewLiteralRunsFixture(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "other\nnode:fs "+stop.found+"\n" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			viewLiteralRunsEmitted(t, program)
			want := run{exitCode: 70, stdout: []byte("other\n"), stderr: []byte("adamic: panic: field read failed: imported.argument.literal expected Kind, found number " + stop.found + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
				}
			}
		})
	}
}
