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
// through a conditional names the branch's type, not the conditional's (n24), and so does one with an
// object literal arm beside a dictionary arm (n29): a fresh literal declares no index signature, so it
// widens nothing, and the dictionary arm is what's refused. A literal arm can still carry the lie in:
// a spread of a dictionary inside it (n30) or a computed key (n31) gives the literal an index signature,
// and the conditional is refused at its declaration as that dictionary seen as Options. Each program's whole
// diagnostic is pinned in its .refused file, and its first line names its refusal. Mutants, each
// alone: admitting a declared property seen through a wider index signature (declaredKeyWidened's
// first direction) turns r03, r07, r26 and v01 red; admitting a dictionary read through a declared
// property (its second) turns r05, r15 and r16 red; admitting the write (declaredKeyWrite's refusal
// dropped) turns r01, r02 and v06 red; leaving the type arguments unjudged (declaredKeyBinding
// dropped) turns n05 red. Admitting a dictionary read through a declared property also turns n29, n30 and
// n31 red.
// A delete through a key that isn't one literal, of a required declared property it can remove, is
// refused too (#p9jj9jg): by a string key, the road records_named_operations.a ran (d01), and by a
// "target" | "required" one (d02). Mutant: recordDeclaredDeleteRefusal returning nil turns d01 and d02
// red. The same delete through a receiver typed T extends Options (d03) never reaches that rule:
// deleteRefusal refuses a delete whose receiver isn't a record target first, so the generic road is
// pinned refused by its words.
func TestRecordDeclaredKeysAreRefusedInA(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/record_declared_keys_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".a" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			if strings.HasSuffix(string(source), "console.log(typeof options.outFile);\n") &&
				string(truth.stdout) != "number\n" {
				test.Fatalf("source Node: stdout %q, want the lie, number", truth.stdout)
			}
			_, err = lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				test.Fatalf("want a declared-key refusal, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				test.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
			if header := "// a-check: refused " + refused.What + "\n"; !strings.HasPrefix(string(source), header) {
				test.Fatalf("want first line %q", header)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 44 {
		test.Fatalf("found %d refused programs, want 44", programs)
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
// closed. In .ts, a storage view that drops outFile lowers since #6c4swvt G3a (r03, r04, r07 to r09,
// r17 to r20, r27): the write through the view lands, typeof observes the number as Node does, and a
// named read of outFile stops on it (records_options_base). Every other road keeps the stop it had
// before the .a refusal: a view that adds outFile (r16), the named write and the optional widening stop
// where they did, and the type-argument roads too, whose .a
// twins are refused. A binding is a relation site (#xywvn06), but it makes no slot and no view, so the
// storage gap that stops each road's direct twin doesn't stop it; g03, a Dict bound to T extends Options,
// is refused there by the hiding rule instead, as its direct twins r05 and r23 are. Each program that
// lowers agrees with Node on release and sanitized native and on the JavaScript backend. Mutant: the key folding
// dropped (writtenNames judging a folded key as any name) refuses the scanner's shape here.
// In .ts, d01's delete of a required declared property through a string key is not yet (#p9jj9jg), and
// c01's deletes, by a "target" | "extra" key and by the optional target's name, lower in both, since
// neither can remove a required one. Mutant: recordDeclaredDeleteRefusal returning nil turns d01 red.
func TestRecordDeclaredKeyStopsArePinned(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/record_declared_keys_stops"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					test.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				test.Fatalf("Lower: %v", err)
			}
			truth := onNode(test, absolute)
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"release", released(test, program)},
				{"sanitized", sanitized},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if difference := disagreement(truth, backend.got); difference != "" {
					test.Fatalf(
						"%s: %s: Node %q, %s %q",
						backend.name,
						difference,
						truth.stdout,
						backend.name,
						backend.got.stdout,
					)
				}
			}
		})
	}
	if programs != 61 {
		test.Fatalf("found %d programs, want 61", programs)
	}
}
