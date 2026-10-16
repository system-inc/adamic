package oracle

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"generic_body_read.a", "generic_body_indexed.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// This source was NotYet for dynamic this and bind. The universal return rule
// now refuses it earlier; its independent Node observation remains in census_small.
func TestGenericBodyRelationsMaybeBind(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "generic body return") || !strings.Contains(err.Error(), "OmitThisParameter") || !strings.Contains(err.Error(), "adamic/generic-body-relations") {
		t.Fatalf("want dependent return refusal, got %v", err)
	}
}
