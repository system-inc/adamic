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

func TestNotYetLibraryRegexCallbackMutants(t *testing.T) {
	t.Parallel()
	for _, rule := range []string{"match", "literal", "unicode", "global", "offset", "reset", "collection-order"} {
		t.Run(rule, func(t *testing.T) {
			t.Parallel()
			librarySmallRuntimeMutant(t, "notyet_library_regex_callback.a", rule)
		})
	}
}
