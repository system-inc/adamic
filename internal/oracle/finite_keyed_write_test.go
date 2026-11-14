package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// o[k] = v on a fixed object, k a finite union of the object's own field names (#13p3y0c): each program in
// testdata/finite_keyed_write agrees with Node natively, sanitized and on the JavaScript backend, and leaks
// nothing (generatorAgrees). The receiver, then the key, then the value run once each, in that order,
// before the store, which writes the field the key names as o.name = v writes it, a class's fields where
// its layout puts them, strings the program builds let go of when replaced, and a value held two ways
// fitted by each arm to its own field. The mutants are in testdata/finite_keyed_write/mutants.md.
func TestFiniteKeyedWritesAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/finite_keyed_write"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 9 {
		t.Fatalf("%d programs in %s, want 9: a program added or removed needs this count changed with it", programs, directory)
	}
}

// finiteKeyedWriteStops is where each program in testdata/finite_keyed_write_stops stops, and in what
// words: refused is a Refused, otherwise a NotYet, at line of the program. A keyed program with a twin is
// Adamic's condition on #13p3y0c, that each arm is judged exactly as the plain o.f = v it stands for: the
// twin is that named write, which stops today, and the keyed one must stop in the twin's own words, read
// off the twin's lowering, not off this table. The four refused rows are tsc's shape and its probes on
// #13p3y0c, refused at their {} by no-optional-widening until #7n03qzg lands, when they move on.
var finiteKeyedWriteStops = []struct {
	name    string
	refused bool
	line    int
	what    string
	twin    string
}{
	{"absent_optional_keyed.a", false, 12, "writing a possibly absent optional own field", "absent_optional_named.a"},
	{"absent_optional_named.a", false, 10, "writing a possibly absent optional own field", ""},
	{"checked_view_keyed.a", false, 11, "writing a checked object field without its source-slot type certificate", "checked_view_named.a"},
	{"checked_view_named.a", false, 10, "writing a checked object field without its source-slot type certificate", ""},
	{"coalesce.a", true, 5, "optional property Low in Partial<Levels> absent from structural source {}, which can hide fields", ""},
	{"index_signature_union.a", false, 8, "a computed field write to an object with an index signature", ""},
	{"literal_key.a", true, 6, "optional property Low in Partial<Levels> absent from structural source {}, which can hide fields", ""},
	{"or_temporary.a", true, 5, "optional property Low in Partial<Levels> absent from structural source {}, which can hide fields", ""},
	{"readonly_constructor.a", false, 8, "a computed write to readonly field Low", ""},
	{"receiver_alternatives.a", false, 6, "a computed field stored differently across receiver alternatives", ""},
	{"tsc_custom_levels.a", true, 14, "optional property Low in Partial<Levels> absent from structural source {}, which can hide fields", ""},
	{"value_held_differently.a", false, 7, "a computed field write whose value isn't held as every named field is", ""},
}

// Each program Node runs to the end stops at lowering where the table says, in its words, and a keyed
// program in its twin's. The directory holds exactly the table's programs, so a new stop gets a row.
func TestFiniteKeyedWriteStops(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/finite_keyed_write_stops"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]bool{}
	for _, row := range finiteKeyedWriteStops {
		rows[row.name] = true
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		if !rows[entry.Name()] {
			t.Errorf("%s/%s has no row in finiteKeyedWriteStops", directory, entry.Name())
		}
	}
	if programs != len(finiteKeyedWriteStops) {
		t.Fatalf("%d programs in %s, want the table's %d", programs, directory, len(finiteKeyedWriteStops))
	}
	// stop lowers one program, which must stop, and gives the stop's kind, where and what.
	stop := func(t *testing.T, name string) (bool, string, string) {
		t.Helper()
		path, err := filepath.Abs(filepath.Join(repository, directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if truth := onNode(t, path); truth.exitCode != 0 || len(truth.stderr) != 0 {
			t.Fatalf("Node: exit %d, stderr %q", truth.exitCode, truth.stderr)
		}
		_, err = lowered(t, path)
		var notYet *lower.NotYet
		var refused *lower.Refused
		switch {
		case errors.As(err, &notYet):
			return false, notYet.Where, notYet.What
		case errors.As(err, &refused):
			return true, refused.Where, refused.What
		}
		t.Fatalf("want a stop, got %v", err)
		return false, "", ""
	}
	for _, row := range finiteKeyedWriteStops {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			refused, where, what := stop(t, row.name)
			kind := map[bool]string{false: "NotYet", true: "Refused"}
			if refused != row.refused || what != row.what || !strings.Contains(where, row.name+":"+strconv.Itoa(row.line)+":") {
				t.Fatalf("stopped by %s at %s: %q; want %s at %s:%d: %q", kind[refused], where, what, kind[row.refused], row.name, row.line, row.what)
			}
			if row.twin == "" {
				return
			}
			twinRefused, twinWhere, twinWhat := stop(t, row.twin)
			if twinRefused != refused || twinWhat != what {
				t.Fatalf("the keyed write stopped by %s, %q, and its named twin by %s at %s, %q: each arm must be judged as the named write is", kind[refused], what, kind[twinRefused], twinWhere, twinWhat)
			}
			t.Logf("%s and its twin %s: %q", where, twinWhere, what)
		})
	}
}
