package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are TypeScript sites: use a .ts temporary rather than admitting
// prototype mutation into Adamic's sound .a subset. Gap closure is intentional:
// a compiler landing must replace the corresponding refusal assertion.
func TestScout22Slice2CompilerAndPrototypeGaps(t *testing.T) {
	paths, err := filepath.Glob("../../stage3/scout/22-slice2/gaps/*.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 11 {
		t.Fatalf("want 11 remaining site witnesses, got %d", len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			temporary := filepath.Join(t.TempDir(), "main.ts")
			if err = os.WriteFile(temporary, source, 0o644); err != nil {
				t.Fatal(err)
			}
			program, loadErr := load.Load([]string{temporary})
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			_, err = Lower(context.Background(), program)
			if name := filepath.Base(path); name == "debug-554-function-provenance.a" || name == "debug-594-function-provenance.a" {
				// The merged Object intrinsic observations already prove typeof this global method.
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var refused *Refused
			var notYet *NotYet
			if !errors.As(err, &refused) && !errors.As(err, &notYet) {
				t.Fatalf("want explicit compile refusal, got %v", err)
			}
			if filepath.Base(path) == "prototype-mixed-cycle.a" {
				for _, reason := range []string{"field 'a'", "[[Prototype]] link", "compiler graph-region proof"} {
					if !strings.Contains(err.Error(), reason) {
						t.Fatalf("missing %q: %v", reason, err)
					}
				}
			}
			if filepath.Base(path) == "prototype-region-lifetimes.a" && !strings.Contains(err.Error(), "prototype link across region lifetimes; compiler graph-region proof is the way out") {
				t.Fatal(err)
			}
			t.Log(err)
		})
	}
}
