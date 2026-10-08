package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures,
		struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/hidden_boundary_name_binding.a", true, false},
		struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/hidden_boundary_name_binding_refused.a", false, false},
	)
}

// The historical read followed rollback of a branded-name declaration. Keep
// that representation boundary explicit instead of treating any refusal as enough.
func TestHiddenBoundaryNameBindingRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_name_binding_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var boundary *lower.NotYet
	if !errors.As(err, &boundary) || boundary.What != "a value of type Name | undefined" {
		t.Fatalf("want the branded-name representation boundary, got %v", err)
	}
}
