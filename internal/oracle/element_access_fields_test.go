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

// fields.a writes a fixed object's fields by a finite key, value[key] at line 12, which lowers since
// #13p3y0c (finiteElementWrite), so the witness now stops at its next construct: line 26's object literal
// whose field names are numbers, which methodName doesn't name yet. Without setIndex's fixed-object arm
// it stops at 12 again, "assigning an element of a value" (finite_keyed_write/mutants.md, (a)).
func TestComputedFieldsWriteRefusalPin(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/notyet_element_access/fields.a"
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, absolute)
	const reason = "a computed field name"
	const location = "fields.a:26:"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), reason) || !strings.Contains(notYet.Where, location) {
		t.Fatalf("want located refusal %q at %s: %v", reason, location, err)
	}
	t.Logf("%v", err)
}
