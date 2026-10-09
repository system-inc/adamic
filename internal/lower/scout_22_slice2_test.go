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
	t.Parallel()
	paths, err := filepath.Glob("../../stage3/scout/22-slice2/gaps/*.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 12 {
		t.Fatalf("want 12 site witnesses, got %d", len(paths))
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
			if filepath.Base(path) == "compiler-cached-intrinsic.a" {
				if err != nil {
					t.Fatal(err)
				}
				t.Log("main proves the cached intrinsic signature from provenance")
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

func TestScout22PrototypeRuntimeBoundary(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`const prototype = {}; const child: {} = Object.create(prototype);`, `const value = {}; Object.getPrototypeOf(value);`} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), "awaits runtime/step22-prototype-links") || !strings.Contains(err.Error(), "use a declared object or class with composition") {
			t.Fatalf("want runtime boundary with fix, got %v", err)
		}
	}
}
func TestScout22NullConversionResultStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const value = { toString: (): string | null => null }; console.log(String(value));`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "distinct null and undefined tags") {
		t.Fatalf("want null-result refusal, got %v", err)
	}
}
