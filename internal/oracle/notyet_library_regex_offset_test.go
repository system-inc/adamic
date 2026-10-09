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

// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestNotYetLibraryRegexOffsetMutants(t *testing.T) {
	for _, rule := range []string{"offset", "input"} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
		t.Run(rule, func(t *testing.T) { librarySmallRuntimeMutant(t, "notyet_library_regex_offset.a", rule) })
	}
}
