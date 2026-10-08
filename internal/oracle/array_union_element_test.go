package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/array_union_reference_elements.a", true, false})
}

func TestArrayUnionIncompatibleStorageRefused(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/array_union_scalar_reference.a"))
	if err != nil {
		t.Fatal(err)
	}
	if got := onNode(t, path); disagreement(run{stdout: []byte("string\nnumber\n")}, got) != "" {
		t.Fatalf("Node: %+v", got)
	}
	_, err = lowered(t, path)
	var unsupported *lower.NotYet
	if !errors.As(err, &unsupported) || unsupported.What != "an array union with incompatible element storage; narrow the array before reading its elements" {
		t.Fatalf("want incompatible slot refusal, got %v", err)
	}
}
