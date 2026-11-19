package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// rulings.md:72's first slice (#0b9xqhh): in .ts, a relation whose only failure under the hiding rule is an
// optional field of the target absent from a structural source crosses as an untagged checked view of the
// slot's type (internal/lower/relation_views.go), where rule 3300 refused it before. Each program is held to
// Node, and its row says what it shows:
//
//   - agrees: it lowers, and release native, sanitized native and the JavaScript backend print what Node
//     prints: the four relation sites (argument, initializer, return, assignment), each reading an absent and
//     a present field, Language's road 3 read through Target, and road 4's Target | undefined widening;
//   - panics: it lowers, prints what Node prints up to a checked read the lie fails, and stops there with exit
//     70 and the read's message on every backend, where Node prints on: Language's roads 2 (with the alias's
//     label required) and 5, and Lower's nested union row;
//   - stops: it doesn't lower, and the stop holds text: Language's road 1 (the getter, stopped before any read),
//     road 2 with the alias's label optional (its write NotYet, waiting on :73),
//     road 3 through a second type and the generic body lowered twice (both NotYet, never raw), road 4's lie
//     (refused at the alias that writes it), the shorthand this slice can't view (refused by the post-pass in
//     rule 3300's words) and the .a twin (refused by rule 3300 as before).
//
// The fixtures sit in a directory of their own, since the flow and fresh corpora lower every program at
// testdata's top level. Mutants (internal/lower/relation_views.mutants.md): checkRelationViewsTaken removed
// lowers shorthand.ts and reads its label raw; the entry deleted when it's taken lowers lowered_twice.ts with
// the second instantiation unviewed; the .ts condition dropped lowers adamic_twin.a. Each turns its row red.
// The rows are three tests, one a kind, so no one test builds every program on three backends.
const relationViewsDirectory = "internal/oracle/testdata/relation_views"

type relationViewRow struct{ kind, stdout, detail string }

var relationViewRows = map[string]relationViewRow{
	"argument.ts":              {kind: "agrees"},
	"initializer.ts":           {kind: "agrees"},
	"return.ts":                {kind: "agrees"},
	"assignment.ts":            {kind: "agrees"},
	"road3_containers.ts":      {kind: "agrees"},
	"road4_undefined_slot.ts":  {kind: "agrees"},
	"road2_required_alias.ts":  {kind: "panics", stdout: "first\n", detail: "adamic: panic: field read failed: target.label"},
	"road5_nested_two_deep.ts": {kind: "panics", stdout: "absent\ntwo\n", detail: "adamic: panic: field read failed: target.inner"},
	"nested_union.ts":          {kind: "panics", stdout: "absent\n2\n", detail: "adamic: panic: field read failed:"},
	"road1_getter.ts":          {kind: "stops", detail: "stage 0 can't lower a getter in a checked field contract yet"},
	// Waits on :73 (#x0nd1mh): the wider alias's optional label can't be written until its certificate covers it.
	"road2_wider_alias.ts":   {kind: "stops", detail: relationViewsDirectory + "/road2_wider_alias.ts:15:23: stage 0 can't lower writing a possibly absent optional own field yet"},
	"road3_other_type.ts":    {kind: "stops", detail: "stage 0 can't lower a read of field label through type Seen, which may hold a checked view's value yet"},
	"lowered_twice.ts":       {kind: "stops", detail: "stage 0 can't lower a read of field label through type Seen, which may hold a checked view's value yet"},
	"road4_undefined_lie.ts": {kind: "stops", detail: "Adamic 0.1 refuses a value of type Holder seen as LooseHolder, which can write"},
	"shorthand.ts":           {kind: "stops", detail: relationViewsDirectory + "/shorthand.ts:14:20: Adamic 0.1 refuses optional property label in Target absent from structural source Source, which can hide fields; declare label on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	"adamic_twin.a":          {kind: "stops", detail: relationViewsDirectory + "/adamic_twin.a:12:22: Adamic 0.1 refuses optional property label in Target absent from structural source Source, which can hide fields; declare label on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
}

func TestRelationViewsAgreeWithNode(t *testing.T) {
	t.Parallel()
	relationViewsOfKind(t, "agrees", func(t *testing.T, path string, _ relationViewRow) {
		loomViewsDAgrees(t, path)
	})
}

func TestRelationViewsPanicAtTheRead(t *testing.T) {
	t.Parallel()
	relationViewsOfKind(t, "panics", func(t *testing.T, path string, row relationViewRow) {
		absolute, err := filepath.Abs(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, absolute)
		if err != nil {
			t.Fatalf("Lower: %v", err)
		}
		panicsAtTheRead(t, absolute, program, row.stdout, row.detail)
	})
}

// panicsAtTheRead holds a lowered program whose lie fails a checked read (rulings.md:72): Node runs it to
// the end with exit 0 and prints past stdout, and release native, sanitized native and the JavaScript backend
// each print stdout, what Node printed before the read, then stop at the read with exit 70 and a panic line
// holding words. TestLogicalAssignmentViewsArePinned, TestCastSlotHiding and TestConstraintRelationSites hold
// their moved pins with it too.
func panicsAtTheRead(t *testing.T, absolute string, program *ir.Program, stdout, words string) {
	t.Helper()
	truth := onNode(t, absolute)
	if truth.exitCode != 0 || !strings.HasPrefix(string(truth.stdout), stdout) || len(truth.stdout) == len(stdout) {
		t.Fatalf("source Node: exit %d, stdout %q: want it to print past %q", truth.exitCode, truth.stdout, stdout)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != stdout || !strings.Contains(string(got.stderr), words) {
			t.Errorf("%s: exit %d, stdout %q, stderr %q: want exit 70, stdout %q, stderr holding %q", backend, got.exitCode, got.stdout, got.stderr, stdout, words)
		}
	}
}

// panicsStop is a .stop file's form for a pin whose outcome is panicsAtTheRead's: its first line is
// "panics", and its second the words the panic line holds. The program stops at its first read of the lie,
// before it prints anything. It says whether expected is that form, and gives the words.
func panicsStop(expected string) (string, bool) {
	first, words, _ := strings.Cut(strings.TrimSpace(expected), "\n")
	return words, first == "panics"
}

func TestRelationViewsStop(t *testing.T) {
	t.Parallel()
	relationViewsOfKind(t, "stops", func(t *testing.T, path string, row relationViewRow) {
		absolute, err := filepath.Abs(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowered(t, absolute)
		if err == nil {
			t.Fatalf("lowered, want a stop holding %q", row.detail)
		}
		if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); !strings.Contains(got, row.detail) {
			t.Fatalf("stop: got %q, want one holding %q", got, row.detail)
		}
	})
}

// relationViewsOfKind runs check on each program of the directory whose row is of kind, and fails on a
// program with no row or a row with no program, so one lost from the directory can't shrink the walk.
func relationViewsOfKind(t *testing.T, kind string, check func(t *testing.T, path string, row relationViewRow)) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repository, relationViewsDirectory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		row, listed := relationViewRows[entry.Name()]
		if !listed {
			t.Errorf("%s has no row: say what it shows", entry.Name())
			continue
		}
		if row.kind != kind {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			check(t, relationViewsDirectory+"/"+entry.Name(), row)
		})
	}
	if programs != len(relationViewRows) {
		t.Fatalf("found %d programs, want %d", programs, len(relationViewRows))
	}
}
