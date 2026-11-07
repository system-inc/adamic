package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// Intl.Collator's head 1 (#zjndrx5, Adamic's ruling at 06:39Z, Runtime's two musts): tsc's collator
// lowers, and its compare stops honestly at run time. The fixtures are in testdata/intl_collator,
// out of the top of testdata, which the flow and fresh corpora lower whole. Each is held to Node,
// natively under the sanitizers and released, and on the JavaScript backend.
func intlCollatorFixture(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/intl_collator", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// constructs.a, the one collator fixture that runs to its end, is in the oracle's table too, so its
// counts are recorded (testdata/counts) and every whole-table check runs it. The stop fixtures stop
// as run-time NotYets, which stay out of the count registry.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/intl_collator/constructs.a", true, false})
}

// A collator made from an undefined locale goes on as Node's does, and natively it and its compare
// closure are freed: the leak check finds nothing.
func TestIntlCollatorConstructsAsNodeDoes(t *testing.T) {
	t.Parallel()
	path := intlCollatorFixture(t, "constructs.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("tsc's collator must lower: %v", err)
	}
	truth := onNode(t, path)
	if want := "collator object\ntwo distinct\n"; truth.exitCode != 0 || string(truth.stdout) != want {
		t.Fatalf("source Node: exit %d, stdout %q, want %q", truth.exitCode, truth.stdout, want)
	}
	got, binary := natively(t, program)
	for name, run := range map[string]run{"sanitized": got, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, run); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, run.exitCode, run.stdout, run.stderr)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

// A program's own collator, a class implementing Intl.Collator (user_class.a), an object literal typed
// Intl.Collator (user_literal.a), and either handed to a parameter typed Intl.Collator
// (user_parameter.a), calls its own compare, as on Node, natively under the sanitizers and released,
// and on the JavaScript backend (Language's c01 to c03 on #zjndrx5). The call finds its callee on the
// object at run time, never by the declared type: the host collator's compare is a private slot only
// its shape reaches (library_intl.c, object.c's adamic_object_callee). Mutant: lowering the call as a
// read of #compare off anything typed Intl.Collator, so each finds no such field.
func TestIntlCollatorUserCollatorsCallTheirOwnCompare(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"user_class.a", "2\n"},
		{"user_literal.a", "2\n"},
		{"user_parameter.a", "2\n2\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := intlCollatorFixture(t, probe.file)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("source Node: exit %d, stdout %q, want %q", truth.exitCode, truth.stdout, probe.stdout)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a program's own collator must lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"sanitized": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// intlCollatorHostStop is the host form's whole message for a key, as ir.DynamicKeyedStops says it.
func intlCollatorHostStop(key string) string {
	for _, stop := range ir.DynamicKeyedStops {
		if stop.Name == "RuntimeValue" {
			return stop.Message + " (reading '" + key + "')"
		}
	}
	return "no RuntimeValue stop in ir.DynamicKeyedStops"
}

// intlCollatorStops are the stop fixtures: what each prints before it stops, the whole message it
// stops in, and what Node prints running the source, which goes on past the stop.
var intlCollatorStops = []struct{ file, stdout, message, node string }{
	// Pin 1, tsc's shape. Mutant: library_intl.c's collator_compare answering ordinal order, so the
	// run goes on past the call natively and prints a line Node's doesn't match.
	{"tsc_shape.a", "comparer made\n", ir.IntlCollatorCompareStop, "comparer made\ncompared after\n"},
	// Pin 6, a defined locale, each its own program, since each stops at its construction: 'en' written
	// in place, a malformed tag whose RangeError the program catches on Node, and 'en' held in a
	// variable. Mutant: adamic_intl_collator_new's locale test dropped, so each constructs natively and
	// the run goes on.
	{"locale_en.a", "before\n", ir.IntlCollatorLocaleStop, "before\ncollator object\n"},
	{"locale_malformed.a", "undefined object\n", ir.IntlCollatorLocaleStop, "undefined object\nmalformed RangeError\n"},
	{"locale_variable.a", "locale en\n", ir.IntlCollatorLocaleStop, "locale en\ncollator object\n"},
	// Pin 7, Object.keys, which stops in the keyless host form and never lists compare. Mutant:
	// library_intl.c's shape left plain (NULL methods), so native answers from its slots.
	{"own_keys.a", "collator\n", ir.DynamicKeysStop, "collator\nkeys []\n"},
	// Pin 7, an in test of compare, true on Node, which stops in the host form keyed on compare. Its
	// mutant is own_keys.a's: the shape left plain, so native answers from its slot.
	{"in_compare.a", "collator\n", intlCollatorHostStop("compare"), "collator\nin yes\n"},
	// A program's collator and the host's in one array, called through one call site (Runtime's seam on
	// head 1-4): the program's prints 2, and the host's reaches its #compare slot by its shape and
	// stops. Mutant: adamic_object_callee's arm for the collator's shape dropped, so the host's call
	// finds no compare and stops in other words.
	{"user_and_host.a", "2\n", ir.IntlCollatorCompareStop, "2\n-1\n"},
}

// Each stop, on both backends, in the same whole message byte for byte, after exactly Node's output
// up to it, while Node runs the source to the end.
func TestIntlCollatorStopsAlikeOnBothBackends(t *testing.T) {
	t.Parallel()
	for _, probe := range intlCollatorStops {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := intlCollatorFixture(t, probe.file)
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("the program must lower and stop at run time: %v", err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte("adamic: panic: " + probe.message + "\n"), exitCode: 70}
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": backend} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if difference := disagreement(native, backend); difference != "" {
				t.Errorf("the backends differ: %s", difference)
			}
		})
	}
}

// intlCollatorWriteStop is what a write to a collator's member stops in, named by its member.
func intlCollatorWriteStop(member string) string {
	return "a write to an Intl.Collator's " + member + ", which Node's collator holds only on Intl.Collator.prototype (compare there is a getter, and writing it throws a TypeError)"
}

// intlCollatorOwnPropertiesStop is what an Object call reading a collator's own properties stops in.
func intlCollatorOwnPropertiesStop(call string) string {
	return "Object." + call + " on an Intl.Collator, which has no own property on Node"
}

// intlCollatorCompileStops are the roads Runtime named reviewing head 1's C (#zjndrx5), each a program
// that type-checks, since TypeScript types compare as an assignable method, and that native would
// answer from its shape's slots where Node's collator has no own property: compare is a getter on
// Intl.Collator.prototype. Each stops at compile time by name, a NotYet, or, for the spread, the
// library spread rule's refusal (librarySpreadRefusal), never lowering. node and exit are what Node
// does with the source. Mutants: library_intl.go's write stop dropped, so write.a and write_caught.a
// meet the method value refusal's words; that and the method value refusal skipping a write's target,
// so the write lowers; Object.assign, entries, values and getOwnPropertyNames let by
// intlCollatorOwnProperties, so each meets other words or lowers.
var intlCollatorCompileStops = []struct {
	file, node string
	exit       int
	refused    bool
	what       string
}{
	{"write.a", "collator\n", 70, false, intlCollatorWriteStop("compare")},
	{"write_caught.a", "collator\nTypeError\n", 0, false, intlCollatorWriteStop("compare")},
	{"spread.a", "collator\nkeys 0\n", 0, true, "an object spread through a view that may hold library class Collator (its internal fields are not own enumerable properties)"},
	{"assign.a", "collator\nkeys 0\n", 0, false, intlCollatorOwnPropertiesStop("assign")},
	{"entries.a", "collator\nentries 0\n", 0, false, intlCollatorOwnPropertiesStop("entries")},
	{"values.a", "collator\nvalues 0\n", 0, false, intlCollatorOwnPropertiesStop("values")},
	{"own_names.a", "collator\nnames []\n", 0, false, intlCollatorOwnPropertiesStop("getOwnPropertyNames")},
}

// Each road stops before anything of it lowers, in its own words, where Node runs on.
func TestIntlCollatorWritesAndCopiesStopByName(t *testing.T) {
	t.Parallel()
	for _, probe := range intlCollatorCompileStops {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path := intlCollatorFixture(t, probe.file)
			if truth := onNode(t, path); truth.exitCode != probe.exit || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q, want exit %d, stdout %q", truth.exitCode, truth.stdout, truth.stderr, probe.exit, probe.node)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			var refused *lower.Refused
			switch {
			case probe.refused && errors.As(err, &refused):
				if refused.What != probe.what {
					t.Errorf("refused %q, want %q", refused.What, probe.what)
				}
			case !probe.refused && errors.As(err, &notYet):
				if notYet.What != probe.what {
					t.Errorf("NotYet %q, want %q", notYet.What, probe.what)
				}
			default:
				t.Errorf("got %v, want the stop %q (refused %v)", err, probe.what, probe.refused)
			}
		})
	}
}
