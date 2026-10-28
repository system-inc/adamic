package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// typeof says object for a typed array and a class instance alike, so a local narrowed to Box by instanceof can't be
// checked once a call has written a typed array into it. localRead used to stop only when the union also held an
// object, array, Map or record held unlike the narrowed member, and let every other object kind through to be read as
// a Box: on main, release native printed 0 for this fixture where Node prints undefined. It now lets through, beside
// the narrowed members, only undefined, null (which the recheck then tests for) and the kinds typeof tells from an
// object (number, boolean, string, function, their maybe pairs), so every other member stops at the narrowing, a kind
// added to the IR later too. An Array.isArray recheck keeps its own list. The fixtures sit in reland_refused, beside
// narrowed_union_record_tag.ts, since they don't lower and the flow and fresh corpora lower every program in testdata.
// Mutant: the old list put back in internal/lower/locals.go; with it applied, this test goes red, and
// ADAMIC_NARROWED_OBJECT_KINDS_MUTANT=1 holds the mutant's admitted program to Node on sanitized and release native to
// show the miscompile (testdata/narrowed-union-object-kinds-mutant.json).
func TestNarrowedUnionTypedArrayTag(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused/narrowed_union_typed_array_tag.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("undefined\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if os.Getenv("ADAMIC_NARROWED_OBJECT_KINDS_MUTANT") == "1" {
		if err != nil {
			t.Fatalf("the mutant must admit the program: %v", err)
		}
		native, _ := nativelyUncached(t, program)
		for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program)} {
			if disagreement(truth, got) == "" {
				t.Fatalf("the mutant's %s program agrees with Node: %+v", name, got)
			}
			t.Logf("mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q", name, got.exitCode, got.stdout, got.stderr, truth.stdout)
		}
		return
	}
	refusedAtNarrowing(t, err, "narrowed_union_typed_array_tag.ts:4:5")
}

// A promise and another class are each typeof object too, so each stops at the narrowing. On main, release native
// printed 0 for the promise and 5.15e-310 for the other class, where Node prints undefined.
func TestNarrowedUnionObjectKindTags(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at string }{
		{"narrowed_union_promise_tag.ts", "4:5"},
		{"narrowed_union_class_tag.ts", "5:5"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte("undefined\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			refusedAtNarrowing(t, err, probe.file+":"+probe.at)
		})
	}
}

// null is typeof object as well, but it needn't stop the read: where the union holds it, the recheck tests for it.
// Box | number | null, nulled by a call inside the narrowed block, now panics as the narrowing's other members do
// (TestNarrowedUnionMemberCheck), where main's release native read 4.8e-310 and sanitized saw a global overflow at
// adamic_null; Node throws a TypeError. Box | null is held as one object reference, not a union, so it never reaches
// this recheck: defined (internal/lower/narrowed.go) checks the reference the checker narrowed null out of and throws
// Node's own TypeError, on every backend, on main and here.
func TestNarrowedNullableObjectRead(t *testing.T) {
	t.Parallel()
	typeError := run{exitCode: 70, stderr: []byte("adamic: panic: TypeError: Cannot read properties of null (reading 'v')\n")}
	for _, probe := range []struct {
		name, union string
		want        run
	}{
		{"object", "Box | null", typeError},
		{"union", "Box | number | null", run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "narrowed_nullable_"+probe.name+".ts")
			source := "class Box { v = 1; }\nlet x: " + probe.union + " = new Box();\nfunction clear(): void { x = null; }\nif (x instanceof Box) {\n    clear();\n    console.log(String(x.v));\n}\n"
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(typeError, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, _ := nativelyUncached(t, program)
			for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(probe.want, got); difference != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

func refusedAtNarrowing(t *testing.T, err error, where string) {
	t.Helper()
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), where) || !strings.Contains(err.Error(), "stage 0 can't lower a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables yet") {
		t.Fatalf("want the object-tag NotYet at the narrowing, got %v", err)
	}
}
