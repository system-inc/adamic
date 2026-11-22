package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

const recordSeenAsIntersectionDirectory = "internal/oracle/testdata/record_seen_as_intersection"

// recordSeenAsIntersectionStops are #mft3zc6's records seen as an intersection of plain shapes, which is
// held as a fixed object, and such an intersection seen as a record, each stopping at the relation in
// recordStorageView's words, whole: where is the stop's line and column, and want is Node's output, so
// each program is known to check and run. The stop is a NotYet, never a refusal, since each is a valid
// program Node runs (Adamic's condition on #mft3zc6). On main partial_number.ts compiled and panicked
// natively as a compiler bug, and the two intersection sources compiled and were an AddressSanitizer SEGV
// natively (J-rmbk4bb), where Node printed. partial_boolean is partial_number's road with an optional
// boolean field, whose intersection is held as a fixed object since #tq6yzqy.
var recordSeenAsIntersectionStops = []struct{ name, where, stop, want string }{
	{"partial_number", "12:7", `stage 0 can't lower a Partial<Record<"count", number>> seen as Entry (fixed objects and records have different storage; copy explicitly) yet`, "1\nundefined\n2\n"},
	{"partial_boolean", "12:7", `stage 0 can't lower a Partial<Record<"valid", boolean>> seen as Entry (fixed objects and records have different storage; copy explicitly) yet`, "false\nundefined\ntrue\n"},
	{"intersection_partial_record", "6:7", `stage 0 can't lower a Entry seen as Partial<Record<"count" | "label", number>> (fixed objects and records have different storage; copy explicitly) yet`, "3\nundefined\n"},
	{"intersection_argument", "9:7", `stage 0 can't lower a Entry seen as Partial<Record<"count" | "label", number>> (fixed objects and records have different storage; copy explicitly) yet`, "3\n"},
	{"intersection_index_signature", "6:7", `stage 0 can't lower a Entry seen as { [key: string]: number | undefined; } (fixed objects and records have different storage; copy explicitly) yet`, "3\nundefined\n"},
}

// recordSeenAsIntersectionAgreeing are the controls: fixed objects seen as the same intersection, judged
// field by field, lower and agree. want is Node's output.
var recordSeenAsIntersectionAgreeing = []struct{ name, want string }{
	{"fixed_object", "1\nundefined\n2\n"},
}

// Each stopping program runs on Node as its row says, and stops as a NotYet where and with exactly the
// words its row says.
func TestRecordSeenAsIntersectionStops(t *testing.T) {
	t.Parallel()
	for _, stop := range recordSeenAsIntersectionStops {
		t.Run(stop.name+".ts", func(t *testing.T) {
			t.Parallel()
			path := recordSeenAsIntersectionPath(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.want {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, stop.want, truth.stderr)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			want := stop.name + ".ts:" + stop.where + ": " + stop.stop
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Error(), want) {
				t.Fatalf("got %v, want a NotYet ending %q", err, want)
			}
		})
	}
}

// Each control agrees with Node natively, sanitized and on the JavaScript backend, and lets go of
// everything it made (objectIntersectionAgrees).
func TestRecordSeenAsIntersectionFixedAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range recordSeenAsIntersectionAgreeing {
		t.Run(pin.name+".ts", func(t *testing.T) {
			t.Parallel()
			path := recordSeenAsIntersectionPath(t, pin.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != pin.want {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, pin.want, truth.stderr)
			}
			objectIntersectionAgrees(t, path)
		})
	}
}

// Every program in the directory is a row above, so one added without its row can't go unheld.
func TestRecordSeenAsIntersectionProgramsArePinned(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, recordSeenAsIntersectionDirectory))
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]bool{}
	for _, stop := range recordSeenAsIntersectionStops {
		pinned[stop.name+".ts"] = true
	}
	for _, pin := range recordSeenAsIntersectionAgreeing {
		pinned[pin.name+".ts"] = true
	}
	for _, entry := range entries {
		if !pinned[entry.Name()] {
			t.Errorf("%s in %s has no row", entry.Name(), recordSeenAsIntersectionDirectory)
		}
		delete(pinned, entry.Name())
	}
	for name := range pinned {
		t.Errorf("row %s has no program in %s", name, recordSeenAsIntersectionDirectory)
	}
}

// recordSeenAsIntersectionPath is a program's absolute path in the directory.
func recordSeenAsIntersectionPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, recordSeenAsIntersectionDirectory, name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	return path
}
