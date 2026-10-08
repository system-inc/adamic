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
	}{"internal/oracle/testdata/evolving_array_storage.a", true, false})
}
func TestEvolvingArrayChangingStorageOnNode(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/evolving_array_changing_storage.a"))
	if err != nil {
		t.Fatal(err)
	}
	if got := onNode(t, path); disagreement(run{stdout: []byte("1\n2\n")}, got) != "" {
		t.Fatalf("Node: %+v", got)
	}
	_, err = lowered(t, path)
	var unsupported *lower.NotYet
	if !errors.As(err, &unsupported) || unsupported.What != "an array of any" {
		t.Fatalf("want changing storage unsupported, got %v", err)
	}
}
