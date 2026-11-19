package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// slotWritesDirectory holds the programs whose writes of an optional field are admitted where every value
// reaching them holds its slot at the write's type (#w8nanzw, Compiler's accept of Nov 19), and lie.ts, which
// stops at compile time.
const slotWritesDirectory = "internal/oracle/testdata/slot_writes"

// slotWritesThroughAView and slotWritesOnOtherRoads are the directory's programs that agree with Node, cut in
// two so each test builds about as much as view_optional_read's do.
var (
	slotWritesThroughAView = []string{"base_class_view.ts", "class_view.ts", "literal_undefined_view.ts", "literal_value_view.ts"}
	slotWritesOnOtherRoads = []string{"kind_union_optional.ts", "own_field.a"}
)

// Each admitted write through a checked view agrees with Node natively, sanitized and released, and on the
// JavaScript backend: an instance of a class declaring name or of one whose base declares it, a literal built at
// a type holding name as the view does that wrote it undefined or a value. A literal spreading one stops first at
// main's own whole-object copy of a value that may hold a checked view's (checkViewCopies), so it isn't here;
// internal/lower's TestSlotWritesThroughACheckedView names that stop, and its plain twin agrees there. The directory holds exactly the two
// lists' programs and lie.ts, so a program added gets a row.
func TestSlotWritesThroughAViewAgreeWithNode(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, slotWritesDirectory))
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, entry := range entries {
		found = append(found, entry.Name())
	}
	listed := append(append([]string{"lie.ts"}, slotWritesThroughAView...), slotWritesOnOtherRoads...)
	slices.Sort(found)
	slices.Sort(listed)
	if !slices.Equal(found, listed) {
		t.Fatalf("%s holds %v, want %v: a program added or removed needs its list changed with it", slotWritesDirectory, found, listed)
	}
	slotWritesAgree(t, slotWritesThroughAView)
}

// Each admitted write on the other two roads agrees with Node as the views' do: through a kind-checked cast's
// union, a literal that wrote flowNode undefined as tsc's factory does, and a plain own field write in .a.
func TestSlotWritesOnOtherRoadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	slotWritesAgree(t, slotWritesOnOtherRoads)
}

// slotWritesAgree runs each of names in slotWritesDirectory as its own parallel subtest: it must lower and agree
// with Node natively, sanitized and released, and on the JavaScript backend.
func slotWritesAgree(t *testing.T, names []string) {
	t.Helper()
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, slotWritesDirectory, name))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stdout) == 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"sanitized": sanitized, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
				}
			}
		})
	}
}

// A lie the slot alone would admit: lie was built with name, so it holds the slot the write through Element's
// checked view needs, but its own type holds a number there where Element says a string, so stage 0 refuses the
// program at the write, at compile time, in the helper's words, never a stop at run time. Node runs on.
func TestSlotWriteLieIsRefusedAtTheWrite(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, slotWritesDirectory, "lie.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "writing\nafter\n" {
		t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want NotYet at the write, got %v", err)
	}
	want := "a write to optional field name through a checked view, where a value reaching it is built as { kind: number; pos: number; name: number; }, whose name is number, not string | undefined"
	if notYet.What != want || !strings.HasSuffix(notYet.Where, "lie.ts:8:78") {
		t.Fatalf("want NotYet at lie.ts:8:78, %q, got %v", want, err)
	}
}
