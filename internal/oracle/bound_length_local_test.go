package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func TestBoundLengthLocalAgrees(t *testing.T) {
	t.Parallel()
	optionalReceiverAgrees(t, "bound_length_local.a")
}
func TestBoundLengthLocalKeepsRefusals(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, what string }{
		{"class_method", "a bound function's length (bound length metadata)"},
		{"own_method", "a bound function's length (bound length metadata)"},
		{"erased_rebind", "a bound function's length (bound length metadata)"},
		{"name", "a bound function's name (bound name metadata)"},
		{"rest", "bind of a rest-parameter function"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, e := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bound_length_refused", row.name+".a"))
			if e != nil {
				t.Fatal(e)
			}
			_, e = lowered(t, path)
			var n *lower.NotYet
			if !errors.As(e, &n) || n.What != row.what {
				t.Fatalf("want %q, got %v", row.what, e)
			}
		})
	}
}

// These historical stop fixtures now agree; keep their paths so the admission
// delta also measures their transition from the base's refusal.
func TestBoundLengthFormerRefusalsAgree(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"length.a", "typeof_length.a"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			optionalReceiverAgrees(t, "bind_refused/"+fixture)
		})
	}
}
