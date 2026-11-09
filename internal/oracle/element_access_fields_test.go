package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet_element_access/fields.a", false, false})
}

// This witness must stop at its registered construct, with the same reason.
func TestComputedFieldsWriteRefusalPin(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/notyet_element_access/fields.a"
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, absolute)
	const reason = "assigning an element of a value"
	const location = "fields.a:12:"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), reason) || !strings.Contains(notYet.Where, location) {
		t.Fatalf("want located refusal %q at %s: %v", reason, location, err)
	}
	t.Logf("%v", err)
}
