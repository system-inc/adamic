package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const intersectionOptionalBooleanDirectory = "internal/oracle/testdata/intersection_optional_boolean"

// intersectionOptionalBooleanAgreeing are #tq6yzqy's admitted programs: an intersection field held as
// an optional boolean by the plain object rule, as every member that has it holds it (objectIntersection,
// intersectionMembersHoldMaybeBoolean). want is Node's output, held too, so an agreement on the wrong
// text can't pass. read is the scout's program; declared_by_one, agreeing_members and inherited are
// tsc's namedEvaluation.ts, program.ts and expressionToTypeNode.ts shapes; optional_true is
// { valid?: boolean } & { valid?: true }, whose true | undefined is held as an optional boolean too, so
// its members agree.
var intersectionOptionalBooleanAgreeing = []struct{ name, want string }{
	{"read", "false\nundefined\n"},
	{"declared_by_one", "equals true\nequals false\nequals undefined\n"},
	{"agreeing_members", "module true\nmodule false\nmodule undefined\n"},
	{"inherited", "1 true\n2 false\n3 undefined\n"},
	{"optional_true", "true\nundefined\ntrue\n"},
}

// intersectionOptionalBooleanStops are the programs that keep a stop, each in its exact words: where is
// the stop's line and column, and want is Node's output, so each program is known to check and run.
// disagreeing has a member holding valid as a boxed union, so the optional-boolean stop stays as it
// was. write_widening is the scout's write: a plain boolean seen through the intersection, whose
// optional field can be written undefined, stopped by the widening rule before any intersection.
var intersectionOptionalBooleanStops = []struct{ name, where, stop, want string }{
	{"disagreeing", "5:16", "stage 0 can't lower a value of type Entry, an intersection whose field valid is an optional boolean, which its record storage doesn't hold yet", "false\nundefined\n"},
	{"write_widening", "7:22", "Adamic 0.1 refuses a value of type Own seen as Entry, which can write boolean | undefined where boolean is read; ", "true\nundefined\nundefined\n"},
}

// Each admitted program agrees with Node natively, sanitized and on the JavaScript backend, and lets go
// of everything it made (objectIntersectionAgrees).
func TestIntersectionOptionalBooleanAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, pin := range intersectionOptionalBooleanAgreeing {
		t.Run(pin.name+".ts", func(t *testing.T) {
			t.Parallel()
			path := intersectionOptionalBooleanPath(t, pin.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != pin.want {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, pin.want, truth.stderr)
			}
			objectIntersectionAgrees(t, path)
		})
	}
}

// Each stopping program runs on Node as its row says, and stops where and with the words its row says.
func TestIntersectionOptionalBooleanStops(t *testing.T) {
	t.Parallel()
	for _, stop := range intersectionOptionalBooleanStops {
		t.Run(stop.name+".ts", func(t *testing.T) {
			t.Parallel()
			path := intersectionOptionalBooleanPath(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.want {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, stop.want, truth.stderr)
			}
			_, err := lowered(t, path)
			want := stop.name + ".ts:" + stop.where + ": " + stop.stop
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want a stop saying %q", err, want)
			}
		})
	}
}

// Every program in the directory is a row above, so one added without its row can't go unheld.
func TestIntersectionOptionalBooleanProgramsArePinned(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, intersectionOptionalBooleanDirectory))
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]bool{}
	for _, pin := range intersectionOptionalBooleanAgreeing {
		pinned[pin.name+".ts"] = true
	}
	for _, stop := range intersectionOptionalBooleanStops {
		pinned[stop.name+".ts"] = true
	}
	for _, entry := range entries {
		if !pinned[entry.Name()] {
			t.Errorf("%s in %s has no row", entry.Name(), intersectionOptionalBooleanDirectory)
		}
		delete(pinned, entry.Name())
	}
	for name := range pinned {
		t.Errorf("row %s has no program in %s", name, intersectionOptionalBooleanDirectory)
	}
}

// intersectionOptionalBooleanPath is a program's absolute path in the directory.
func intersectionOptionalBooleanPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, intersectionOptionalBooleanDirectory, name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	return path
}
