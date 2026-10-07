package oracle

// Register these probes separately to keep the shared oracle harness unchanged.
func init() {
	for _, name := range []string{"throw", "region", "consume", "region_alias"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/class_instance_key_" + name + ".a", true, false})
	}
	// Container unions in JSON need runtime metadata outside this unit. Until then,
	// the JSON probe must be refused rather than reach the runtime with a scalar schema.
	for _, name := range []string{"json", "typeof"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{genericInstanceKeyFixture(name), name != "json", false})
	}
}

func genericInstanceKeyFixture(name string) string {
	if name == "json" {
		return "internal/oracle/testdata/reland_refused/generic_instance_key_json.a"
	}
	return "internal/oracle/testdata/generic_instance_key_" + name + ".a"
}
