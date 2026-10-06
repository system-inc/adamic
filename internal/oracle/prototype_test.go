package oracle

// Keep this unit's fixture registration here so oracle_test.go can evolve independently.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{
		"internal/oracle/testdata/object_prototype.a", true, false,
	})
}
