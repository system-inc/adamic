package oracle

// The returned polling literal has Node's enumeration order by construction.
// TestNativeAgreesWithNode holds both backends to the keys and values in order.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/numeric_keys_enumeration_order.ts", true, false})
}
