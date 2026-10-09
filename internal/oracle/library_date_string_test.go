package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDateStringRuling10Refusals(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"library_date_string", "library_date_string_utc", "library_date_string_refusal", "library_date_dynamic_parse"} {
		_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/library_date_refused/"+name+".a"))
		if err == nil || !strings.Contains(err.Error(), "ruling 10") {
			t.Fatalf("%s: want ruling 10 refusal, got %v", name, err)
		}
	}
}
func init() {
	for _, name := range []string{"format", "own", "parse", "utc"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_date_area_" + name + ".a", true, false})
	}
}
