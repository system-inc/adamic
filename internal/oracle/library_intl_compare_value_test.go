package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A new collator's compare read as a value (#9tw2b9z, lower/library_intl.go's intlCollatorCompareValue)
// is the closure in its #compare slot, admitted only off the new expression itself and into a slot that
// sees it exactly as itself, the const tsc's comparer is among them (Runtime's ruling (b)). Each program
// in testdata/intl_collator_compare_value, in .a and in .ts, has its stop in its .stop file. "lowers"
// is a program that lowers, agrees with Node natively, sanitized and on the JavaScript backend, and
// leaks nothing: p05, a slot of compare's type. "panics" and a message, then the lines printed before
// the stop, is a program that lowers, whose every backend prints exactly those lines and stops in that
// whole run-time NotYet, where Node prints them and runs on past: p01, tsc's shape, stopping at
// compare's call in head 1's form (#zjndrx5). "record" is a stop not measured yet, which fails here
// with what was measured, so it's written down before anything trusts it. Anything else is the whole
// compile-time diagnostic, its path from the repository's root: p02, a held collator's compare; p03, a
// slot wider than compare's; p04, a class's compare, each the method value refusal. The mutants are in
// testdata/intl_collator_compare_value/mutants.md.
func TestIntlCollatorCompareValueStopsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/intl_collator_compare_value"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := directory + "/" + entry.Name()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			stop, before, _ := strings.Cut(string(expected), "\n")
			switch {
			case stop == "lowers":
				// generatorAgrees holds any program to Node on every backend and to leaking nothing.
				generatorAgrees(t, path)
			case strings.HasPrefix(stop, "panics "):
				truth := onNode(t, absolute)
				if truth.exitCode != 0 || !strings.HasPrefix(string(truth.stdout), before) || len(truth.stdout) == len(before) {
					t.Fatalf("source Node: exit %d, stdout %q, want %q and more after it", truth.exitCode, truth.stdout, before)
				}
				program, err := lowered(t, absolute)
				if err != nil {
					t.Fatalf("the program must lower and stop at run time: %v", err)
				}
				want := run{stdout: []byte(before), stderr: []byte("adamic: panic: " + strings.TrimPrefix(stop, "panics ") + "\n"), exitCode: 70}
				native, _ := natively(t, program)
				for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
					}
				}
			default:
				_, err := lowered(t, absolute)
				got := "lowers"
				if err != nil {
					got = strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), "")
				}
				if stop == "record" {
					t.Fatalf("stop not recorded yet; measured %q", got)
				}
				if got != strings.TrimSpace(string(expected)) {
					t.Fatalf("stop: got %q, want %q", got, strings.TrimSpace(string(expected)))
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 10 {
		t.Fatalf("found %d programs in %s, want 10", programs, directory)
	}
}
