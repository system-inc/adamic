package oracle

import (
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_library_regex_offset.a", true, false})
}

func TestNotYetLibraryRegexOffsetMutants(t *testing.T) {
	t.Parallel()
	for _, rule := range []string{"offset", "input"} {
		t.Run(rule, func(t *testing.T) {
			t.Parallel()
			librarySmallRuntimeMutant(t, "notyet_library_regex_offset.a", rule)
		})
	}
}
