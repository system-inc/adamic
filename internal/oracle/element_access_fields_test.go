package oracle

import "testing"

// fields.a reads a fixed object's fields by a finite key and writes them by one: value[key] at line 12,
// and receiver()[selected] at 49, whose receiver, key and value each run once, in that order, before the
// store. The writes lower since #13p3y0c (finiteElementWrite), so the whole program is held to Node
// natively, sanitized and on the JavaScript backend, and leaks nothing. Without setIndex's fixed-object
// arm it stops at 12 again, "assigning an element of a value" (finite_keyed_write/mutants.md, (a)).
func TestComputedFieldsAgreeWithNode(t *testing.T) {
	t.Parallel()
	generatorAgrees(t, "internal/oracle/testdata/notyet_element_access/fields.a")
}
