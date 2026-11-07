package oracle

import (
	"path/filepath"
	"testing"
)

// An array literal read through a checked view whose declared type is an object stops at the view's
// read, on both backends, in the same words (#nazvcep). Natively an array literal is an array, which a
// view's object check refuses. The JavaScript backend used to mark every array literal as it marks a
// tuple (adamicTuple), and its object check (adamicViewObject) admits a marked array, so it read the
// array as the object the view said: field_object.ts then stopped one read later, at child.label, and
// field_empty_interface.ts, whose object type holds only an optional field, ran on to the end. Node,
// without the checks, prints on. Mutant: the ArrayLiteral case's adamicTuple wrap put back in
// internal/javascript/javascript.go, and both go red on the JavaScript backend.
func TestViewArrayLiteralIsNotAnObject(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/view_array_object_stops/"
	for _, stop := range []struct{ name, stdout, stderr string }{
		{
			"field_object.ts",
			"before\n",
			"field read failed: (node as Leaf).child is not a Thing; expected Thing, found array",
		},
		{
			"field_empty_interface.ts",
			"before\n",
			"field read failed: (holder as Boxed).inner is not a { readonly a?: number; }; expected { readonly a?: number; }, found array",
		},
	} {
		test.Run(stop.name, func(test *testing.T) {
			test.Parallel()
			program, loweredError := lowered(test, filepath.Join(repository, directory+stop.name))
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := run{
				exitCode: 70,
				stdout:   []byte(stop.stdout),
				stderr:   []byte("adamic: panic: " + stop.stderr + "\n"),
			}
			sanitized, _ := nativelyUncached(test, program)
			for backend, got := range map[string]run{
				"native":     releasedUncached(test, program),
				"sanitized":  sanitized,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if diff := disagreement(want, got); diff != "" {
					test.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}

// A real tuple still reads through a view that says a tuple, on both backends, as Node does: a
// tuple-typed literal lowers as an ObjectLiteral marked Tuple, which keeps its mark.
func TestViewTupleStillReadsAsATuple(test *testing.T) {
	test.Parallel()
	loomViewsDAgrees(test, "internal/oracle/testdata/view_array_object_stops/tuple_agrees.ts")
}
