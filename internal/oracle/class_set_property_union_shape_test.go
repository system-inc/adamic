package oracle

import "testing"

func init() {
	for _, name := range []string{"regexp", "stat", "iterator"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/class_set_property_union_shape_" + name + ".a", true, false})
	}
}

// The runtime shape check must run before this stat field is read or boxed.
func TestClassSetPropertyShapeStatCounted(t *testing.T) {
	t.Parallel()
	counted(t, "internal/oracle/testdata/class_set_property_union_shape_stat.a", false, nil, false, false)
}
