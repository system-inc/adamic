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
// widening without its earlier storage stop). So is each write of a value a declared property it can
// land under can't hold: a literal whose own type declares outFile, a computed key folded to
// "outFile", undefined and a number | undefined written under a required number, and a union value
// one member of which misfits. So is a record bound as a type argument to a type parameter its view
// reaches through, Language's roads on 41717bcc: Options as T extends Dict, by a generic function's
// parameter (n05), its return (n26), an explicit put<Options> (n27), a generic class's field (n28), a
// generic arrow (g01), a class's extends clause (g02) and a generic class taken as a value by a
// construct type that instantiates it (g04) or returned by a call into a slot that does (g05), and a Dict
// as T extends Options (g03). A view
// through a conditional names the branch's type, not the conditional's (n24), and a conditional with an
// object literal arm is judged arm by arm, the literal made as its slot and the dictionary arm refused
// (n29). Each program's whole
// diagnostic is pinned in its .refused file, and its first line names its refusal. Mutants, each
// alone: admitting a declared property seen through a wider index signature (declaredKeyWidened's
// first direction) turns r03, r07, r26 and v01 red; admitting a dictionary read through a declared
// property (its second) turns r05, r15 and r16 red; admitting the write (declaredKeyWrite's refusal
// dropped) turns r01, r02 and v06 red; leaving the type arguments unjudged (declaredKeyBindings
// dropped) turns n05 red; skipping a conditional's arms that aren't literals (declaredKeyArmsOf's
// conditional case returning no arms) turns n29 red. Judging such a conditional whole instead refuses
// records_literal_branches.a, whose arms are fresh literals made as Options. A delete through a key that
// isn't one literal, of a required declared property it can remove, is refused too (#p9jj9jg): by a
// string key, the road records_named_operations.a ran (d01), and by a "target" | "required" one (d02).
// Mutant: recordDeclaredDeleteRefusal returning nil turns d01 and d02 red.
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
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			if strings.HasSuffix(string(source), "console.log(typeof options.outFile);\n") && string(truth.stdout) != "number\n" {
				t.Fatalf("source Node: stdout %q, want the lie, number", truth.stdout)
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
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				t.Fatalf("want first line %q", header)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 40 {
		t.Fatalf("found %d refused programs, want 40", programs)
	}
}

// The rest of the roads keep the stops they had on this change's parent, pinned in each program's
// .stop file. In .a, the control lowers, and so do a computed key folded to a name no declared property
// takes and the scanner's ["" + "constructor"] beside its declared keywords (01_scanner_keywords.a's
// shape). A write whose value each declared property it can land under takes by tsc and by the record's
// storage is judged by the registry as its direct twin is (namedWriteStop), in .a as in .ts: an Other
// into names: Point, by a string key or a 'names' | 'other' one, is refused as adamic/nominal-class, as
// places.names = new Other() is, and a SubPoint into names: Point and an open number into kind: Kind
// lower, as their direct twins do. Undefined under outFile?, which the optional property may hold, is
// left to namedWriteStop's not yet, as ruled (v17). Object.assign(options, dict) (r06) is left not yet
// at its record operation, since tsc infers its target and source apart and no relation sees Dict as
// Options, and r21 at its Readonly<Options> view's storage gap, a line before its widening: both fail
// closed. In .ts, every road keeps the stop it had before the .a refusal: the storage views, the named
// write and the optional widening stop where they did, and the type-argument roads too, whose .a
// twins are refused. Each program that lowers agrees with Node on
// release and sanitized native and on the JavaScript backend. Mutant: the key folding dropped
// (writtenNames judging a folded key as any name) refuses the scanner's shape here. In .ts, d01's delete
// of a required declared property through a string key is not yet (#p9jj9jg), and c01's deletes, by a
// "target" | "extra" key and by the optional target's name, lower in both, since neither can remove a
// required one. Mutant: recordDeclaredDeleteRefusal returning nil turns d01 red.
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
			truth := onNode(t, absolute)
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	if programs != 60 {
		t.Fatalf("found %d programs, want 60", programs)
	}
}
