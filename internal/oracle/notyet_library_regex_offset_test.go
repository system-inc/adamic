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

func TestNotYetLibraryRegexOffsetMutants(test *testing.T) {
	test.Parallel()
	for _, rule := range []string{"offset", "input"} {
		test.Run(rule, func(test *testing.T) {
			test.Parallel()
			librarySmallRuntimeMutant(test, "notyet_library_regex_offset.a", rule)
		})
	}
}
