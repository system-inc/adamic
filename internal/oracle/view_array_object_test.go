package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An array literal read through a checked view whose declared type is an object stops at the view's
// read, on both backends, in the same words (#nazvcep), where the object's contract doesn't admit an
// array (ir.ViewContract's ArraysAdmitted, #s03ndgw): Thing's label is required. Natively an array
// literal is an array, which a view's object check refuses. The JavaScript backend used to mark every
// array literal as it marks a tuple (adamicTuple), and its object check (adamicViewObject) admits a
// marked array, so it read the array as the object the view said: field_object.ts then stopped one read
// later, at child.label. Node, without the checks, prints on. Mutant: the ArrayLiteral case's
// adamicTuple wrap put back in internal/javascript/javascript.go, and this goes red on the JavaScript
// backend. An object whose every field is optional and no array's is admitted instead
// (TestViewArrayObjectAgreesWithNode).
func TestViewArrayLiteralIsNotAnObject(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_array_object_stops/"
	for _, stop := range []struct{ name, stdout, stderr string }{
		{"field_object.ts", "before\n", "field read failed: (node as Leaf).child is not a Thing; expected Thing, found array"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, filepath.Join(repository, directory+stop.name))
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte("adamic: panic: " + stop.stderr + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}

// A real tuple still reads through a view that says a tuple, on both backends, as Node does: a
// tuple-typed literal lowers as an ObjectLiteral marked Tuple, which keeps its mark.
func TestViewTupleStillReadsAsATuple(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/oracle/testdata/view_array_object_stops/tuple_agrees.ts")
}

const viewArrayObjectDirectory = "internal/oracle/testdata/view_array_object/"

// viewArrayObjectAgreeing and viewArrayObjectNested are the directory's programs the stops table doesn't
// name; TestViewArrayObjectProgramsAreAllRun holds the directory to every name.
var viewArrayObjectAgreeing = []string{"field_empty_interface.ts", "two_optional_fields.ts", "union_member.ts"}

const viewArrayObjectNested = "nested_field.ts"

// viewArrayObjectStops are the programs that keep a stop, though Node runs each, held as piece 1's
// view_static_class stops are (viewStaticClassStop): at lowering in lowering's words, or at run time on
// every backend.
var viewArrayObjectStops = []viewStaticClassStop{
	// A name an array has keeps the plain read's stop: Array's push, Object's toString, and length.
	{name: "push_field.ts", stdout: "before\n", stderr: viewArrayObjectFound("{ readonly push?: number; }")},
	{name: "to_string_field.ts", stdout: "before\n", stderr: viewArrayObjectFound("{ readonly toString?: number; }")},
	{name: "length_field.ts", stdout: "before\n", stderr: viewArrayObjectFound("{ readonly length?: number; }")},
	// On the union road, an index and a required field leave the selection no member for an array.
	{name: "index_field.ts", stdout: "before\n", stderr: viewArrayObjectNoMember("number | Indexed")},
	{name: "required_field.ts", stdout: "before\n", stderr: viewArrayObjectNoMember("number | Demanding")},
	// A write through the view to its optional field stops at lowering: no array the view admits has its slot.
	{name: "write.ts", lowering: "a write to optional field a through a checked view"},
	// Written through a type no view checks, it stops at run time, after its value, before it writes, as a
	// class constructor's does (ir.Program's ArraysAdmittedWrite).
	{name: "widened_write.ts", stdout: "before\n", stderr: "adamic: panic: field write failed: a write to a through a checked view that holds an array is not yet lowered\n"},
	{name: "object_keys.ts", stdout: "before\n", stderr: "adamic: panic: Object.keys through a checked view that holds an array is not yet lowered\n"},
}

func viewArrayObjectFound(declared string) string {
	return "adamic: panic: field read failed: (holder as Boxed).inner is not a " + declared + "; expected " + declared + ", found array\n"
}

func viewArrayObjectNoMember(declared string) string {
	return "adamic: panic: field read failed: (holder as Boxed).inner matches no member of " + declared + "; expected " + declared + ", found array\n"
}

// An array held where a checked view expects an object reads as Node reads it (#s03ndgw) when the
// object's contract admits one (ir.ViewContract's ArraysAdmitted: every field optional, and none Array's,
// Object's, length or an index, asked of the checker), and the array holds none of its fields of its own.
// It's never converted: typeof says object, and each field reads absent, undefined, never from a slot.
// Each program agrees with Node natively, sanitized and on the JavaScript backend, and leaks nothing
// (generatorAgrees). field_empty_interface and two_optional_fields hold it at a field typed as the object
// itself, which the read's own check admits (native's adamic_object_view_arrays, adamicViewField's
// arrays); union_member holds it as a member of number | Loose, which the union's selection admits
// (plain_matches, matches). Before, each stopped at the read, found array. Mutants (scratch-s03p2/proof.sh):
// each backend's admission refusing every array, on each road, and these go red.
func TestViewArrayObjectAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range viewArrayObjectAgreeing {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, viewArrayObjectDirectory+name)
		})
	}
}

// Where the array isn't what the view says, or the view would do more than read its absent fields, the
// program keeps a stop, though Node runs it: a field an array has (push, toString, length, an index), a
// required field, a write through the view (at lowering) and through a type no view checks (at run time),
// and Object.keys through it, which would read the array's indices. Each is held to where it stops, on every backend, in today's words. Mutants
// (scratch-s03p2/proof.sh): each backend's admission admitting every array, lowering's name check
// dropped, and each backend's write guard dropped, and these go red.
func TestViewArrayObjectStops(t *testing.T) {
	t.Parallel()
	for _, stop := range viewArrayObjectStops {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, viewArrayObjectDirectory+stop.name))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); node.exitCode != 0 {
				t.Fatalf("Node: %#v", node)
			}
			program, err := lowered(t, path)
			if stop.lowering != "" {
				if err == nil || !strings.Contains(err.Error(), stop.lowering) {
					t.Fatalf("lowering: %v, want a stop saying %q", err, stop.lowering)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte(stop.stderr)}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stdout %q, stderr %q", backend, diff, got.stdout, got.stderr)
				}
			}
		})
	}
}

// A field whose own contract is an object, Nesting's b, reads absent from an admitted array as Node reads
// it, or stops: never the array read raw as that object. Which of the two it is, each backend logs.
func TestViewArrayObjectNestedReadsAbsentOrStops(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, viewArrayObjectDirectory+viewArrayObjectNested))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "before\nabsent\n" {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Logf("stops at lowering: %v", err)
		return
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Logf("%s: agrees with Node", backend)
			continue
		}
		if got.exitCode != 70 || string(got.stdout) != "before\n" || !strings.HasPrefix(string(got.stderr), "adamic: panic: ") {
			t.Errorf("%s: neither Node's absent field nor a stop: exit %d, stdout %q, stderr %q", backend, got.exitCode, got.stdout, got.stderr)
			continue
		}
		t.Logf("%s: stops: %q", backend, got.stderr)
	}
}

// Every program in the directory is one of the tests' above, and each is named once: a program added or
// removed needs its table changed with it.
func TestViewArrayObjectProgramsAreAllRun(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, viewArrayObjectDirectory))
	if err != nil {
		t.Fatal(err)
	}
	named := map[string]int{viewArrayObjectNested: 1}
	for _, name := range viewArrayObjectAgreeing {
		named[name]++
	}
	for _, stop := range viewArrayObjectStops {
		named[stop.name]++
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		if named[entry.Name()] != 1 {
			t.Errorf("%s is named %d times by the tests, want once", entry.Name(), named[entry.Name()])
		}
	}
	if programs != 12 || len(named) != 12 {
		t.Fatalf("%d programs in %s and %d named, want 12: a program added or removed needs this count changed with it", programs, viewArrayObjectDirectory, len(named))
	}
}
