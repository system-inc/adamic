package oracle

// Register these beside the existing narrowed-read fixtures without editing the shared list.
func init() {
	for _, name := range []string{"e4eec87_f1_field_narrowed", "e4eec87_f1_class_narrowed", "e4eec87_f1_alias_narrowed", "e4eec87_f1_field_present"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/" + name + ".a", true, name != "e4eec87_f1_field_present",
		})
	}
}
