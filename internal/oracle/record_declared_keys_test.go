package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// In .a, a declared property's type can't lie through an index signature (#npqr3sx): Language's roads
// on #bq3fvdt, each of which tsc accepts and Node runs to print number for outFile, declared string,
// are refused, and so are a shorthand property and a Readonly<Options> parameter widened (r21's
// widening without its earlier storage stop). Each program's whole
// diagnostic is pinned in its .refused file, and its first line names its refusal. Mutants: admitting
// a declared property seen through a wider index signature (declaredKeyWidened's first direction)
// lowers r03 and r07; admitting a dictionary read through a declared property (its second) lowers
// r16; admitting the write (declaredKeyWrite returning nil) lowers r01.
func TestRecordDeclaredKeysAreRefusedInA(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/record_declared_keys_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".a" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, absolute); truth.exitCode != 0 || string(truth.stdout) != "number\n" {
				t.Fatalf("source Node: exit %d stdout %q, want the lie, number", truth.exitCode, truth.stdout)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a declared-key refusal, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				t.Fatalf("want first line %q", header)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 22 {
		t.Fatalf("found %d refused programs, want 22", programs)
	}
}

// The rest of the roads keep the stops they had, pinned in each program's .stop file: in .a, the
// control lowers and agrees with Node, Object.assign(options, dict) (r06) stops at its record
// operation, since tsc infers its target and source apart and no relation sees Dict as Options, and
// r21 stops at its Readonly<Options> view's storage gap, a line before its widening. In .ts, every
// road keeps the stop it had before the .a refusal: the storage views, the named write and the
// optional widening stop where they did, and those that lower agree with Node, whose declared reads
// here are typeof observations rather than the checked read.
func TestRecordDeclaredKeyStopsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/record_declared_keys_stops"
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
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			if truth, compiled := onNode(t, absolute), released(t, program); disagreement(truth, compiled) != "" {
				t.Fatalf("%s: Node %q, native %q", disagreement(truth, compiled), truth.stdout, compiled.stdout)
			}
		})
	}
	if programs != 28 {
		t.Fatalf("found %d programs, want 28", programs)
	}
}
