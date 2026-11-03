package oracle

import "testing"

// A class's accessors are found among each class's own entries, never through Object.prototype
// (#76y1mth). The JavaScript backend's table of a class's accessors is an object literal, so
// adamicFindAccessor, indexing it bare, found Object's own valueOf on a class without that accessor:
// base_getter.ts then stopped "TypeError: get is not a function" where Node and native print 1, and
// a class with a data field of that name would read as one with the getter. adamicClass's literal
// branch indexed the table the same way, so literal_data_member.ts's data member became an accessor
// without a getter and read undefined. Each agrees with Node natively, sanitized and on JavaScript.
// Mutants: the bare index put back in adamicFindAccessor, and base_getter.ts and dispatch_class.ts go
// red on JavaScript; put back in adamicClass, and literal_data_member.ts does.
func TestAccessorsNamedAfterObjectPrototypeAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"base_getter.ts", "dispatch_class.ts", "literal_data_member.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, "internal/oracle/testdata/accessor_prototype_names/"+name)
		})
	}
}
