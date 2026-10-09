package oracle

// Register with the existing Node, both-backend, sanitizer, leak and counts harness.
// This keeps the unit out of the shared oracle_test.go registry.
func init() {
	for _, name := range []string{"factories", "nested", "hooks", "throw", "stage3_roots", "null_hook_refused"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/json_objects_" + name + ".a", name != "null_hook_refused", false})
	}
}
