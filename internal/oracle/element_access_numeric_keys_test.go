package oracle

// element_access_numeric_keys.ts is a const table keyed by two enum members and a literal, read only through its
// keys, which lowers by Adamic's ruling (A) on #d1chbm7 (internal/lower/numeric_keys.go): both backends are held
// to Node's `a b seven` by TestNativeAgreesWithNode.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/element_access_numeric_keys.ts", true, false})
}
