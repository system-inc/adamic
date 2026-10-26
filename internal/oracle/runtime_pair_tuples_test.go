package oracle

import "testing"

// A [key, value] pair the runtime makes is a tuple to a view, as a literal's is (item 150). An object view with
// index-named fields reads a Map's entry as Node does, and a tuple view reads a Set's [value, value] entry, natively,
// sanitized and on the JavaScript backend. Each casts through a view that extends its source with no tag of its own,
// a downcast whose checks no proof erases, which .a refuses (item 8), so each is .ts.
func TestRuntimePairTuples(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"map_entries_object_view", "set_entries_tuple_view"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, "internal/oracle/testdata/runtime_pair_tuples/"+name+".ts")
		})
	}
}

// Four native sites build a pair with no field-type bytes, so their review programs await #qyfyh4p natively. The
// JavaScript backend marks the same pairs, and is held to Node on them here until the review programs run.
func TestRuntimePairTuplesJavaScript(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"map_spread", "plain_object_entries", "checked_object_entries", "regexp_indices"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgreesBackend(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p150_tuple_from_"+name+".ts", "javascript")
		})
	}
}
