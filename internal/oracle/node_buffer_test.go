package oracle

// Register this unit's fixtures without changing the shared oracle driver.
func init() {
	for _, name := range []string{"encodings", "utf8", "utf16", "bom", "fs_decode", "writes", "crypto", "random"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/node_buffer_" + name + ".a", true, false})
	}
	inputFixtures = append(inputFixtures, struct {
		path       string
		arguments  []string
		unreadable bool
		writes     bool
	}{"internal/oracle/testdata/node_buffer_input.a", nil, false, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/node_buffer_finalized.a", true, false})
}
