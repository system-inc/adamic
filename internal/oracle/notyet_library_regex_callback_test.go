package oracle

import (
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet_library_regex_callback.a", true, false})
}

func TestNotYetLibraryRegexCallbackMutants(test *testing.T) {
	test.Parallel()
	for _, rule := range []string{"match", "literal", "unicode", "global", "offset", "reset", "collection-order"} {
		test.Run(rule, func(test *testing.T) {
			test.Parallel()
			librarySmallRuntimeMutant(test, "notyet_library_regex_callback.a", rule)
		})
	}
}
