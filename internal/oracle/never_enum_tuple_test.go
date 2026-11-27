package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A numeric enum's never read of a tuple's element is judged by the element's own stored type, as an
// array's is (enumStoredType, #xfpsqtn): processEnding's switch over allowedEndings[0], a
// readonly [E, ...E[]] (moduleSpecifiers.ts:1386, first), the same through endings[1], which the rest holds
// (rest), and a fixed [Phase, Level] switched over entry[1], judged by Level and never by Phase | Level,
// each enum holding a value the other lacks (pair). The rest tuples come back from a function declared to
// return one, as the adapted tree's allowedEndings comes from getAllowedEndingsInPreferredOrder, each a
// literal of two, the one length that lowers into readonly [E, ...E[]] today. Each program in range agrees
// with Node on the JavaScript backend and natively under the leak check. Each _out program runs the same
// calls and then one with `n as E` outside the element's enum: Node's assertNever throws there, and Adamic
// traps at the read, before the call, naming the element's enum. Node's words differ, so stdout and the
// exit are held to Node and stderr to the trap's own words. The stops are pinned below, and the mutants in
// testdata/never_enum_tuple/mutants.md.
func TestNeverEnumTupleAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"first": "js minimal ts index\n",
		"rest":  "ts js minimal index\n",
		"pair":  "top low high top\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_tuple", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != stdout {
				t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, truth.exitCode, truth.stdout, truth.stderr)
			}
			backend := onJavaScriptBackend(t, program)
			if difference := disagreement(truth, backend); difference != "" {
				t.Fatalf("javascript: %s; stdout %q stderr %q exit %d", difference, backend.stdout, backend.stderr, backend.exitCode)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("native: %s; stdout %q stderr %q exit %d", difference, actual.stdout, actual.stderr, actual.exitCode)
				}
			}
		})
	}
}

// pair_out's 5 is Phase.Check and no member of Level, so a reading by Phase | Level, or by the tuple's
// first element, names Phase here, or doesn't judge the read at all.
func TestNeverEnumTupleTrapsBeforeTheCall(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct {
		name, stdout, value, identity string
	}{
		{"first_out", "js minimal ts index\n", "9", "ModuleSpecifierEnding"},
		{"rest_out", "ts js minimal index\n", "8", "ModuleSpecifierEnding"},
		{"pair_out", "top low high top\n", "5", "Level"},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_tuple", pin.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 70 || string(truth.stdout) != pin.stdout || !strings.Contains(string(truth.stderr), "Illegal value: "+pin.value+"\n") {
				t.Fatalf("Node must print %q and throw assertNever's Illegal value: %s: exit %d, stdout %q, stderr %q", pin.stdout, pin.value, truth.exitCode, truth.stdout, truth.stderr)
			}
			// The trap's own words, which assertNever's throw can't print: the read stopped before the call.
			want := run{stdout: truth.stdout, stderr: []byte("adamic: panic: unreachable value " + pin.value + " for numeric enum " + pin.identity + "\n"), exitCode: 70}
			native, _, _ := nativelyCheckingLeaks(t, program, false)
			for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
				}
			}
		})
	}
}

// The stops keep today's words: a key that is no numeric literal, endings[first], narrowed by the switch
// over endings[0] (variable_key), and an optional element read at its index (optional).
func TestNeverEnumTupleStops(t *testing.T) {
	t.Parallel()
	const condition = "a call whose argument for a never parameter isn't a rulings.md:76 never read"
	for name, location := range map[string]string{
		"variable_key": "19:23",
		"optional":     "14:21",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file := name + ".ts"
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_tuple", file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), file+":"+location+": stage 0 can't lower "+condition+" yet") {
				t.Fatalf("want a stop at %s:%s with %q, got %v", file, location, condition, err)
			}
		})
	}
}
