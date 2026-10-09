package lower

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// The constructor frame lacks canonical identity. Keep selected closures refused
// until its frame finalization and the adopt-before-cache runtime can land together.
func TestProgramCanonicalConstructorBoundary(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"canonical.a", "canonical_counted.a"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("..", "oracle", "testdata", "program_region", fixture))
			if err != nil {
				t.Fatal(err)
			}
			source, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = LowerWithOptions(context.Background(), source, Options{ProgramRegion: true})
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "Program canonical closure before canonical-cache adoption" {
				t.Fatalf("selected constructor closure must remain refused at the canonical boundary, got %v", err)
			}
			t.Log(err)
		})
	}
}
