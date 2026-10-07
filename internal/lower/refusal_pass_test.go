package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefusalPassRulings(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/refusal_pass/*.a")
	if err != nil || len(paths) == 0 {
		t.Fatalf("fixtures: %v, %v", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".want")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var refused *Refused
			var notYet *NotYet
			var got string
			// Pin the error category as well as the complete location and explanation.
			if strings.Contains(string(want), ": stage 0 can't lower ") {
				if !errors.As(err, &notYet) {
					t.Fatalf("got %v, want NotYet", err)
				}
				got = filepath.Base(notYet.Where) + ": stage 0 can't lower " + notYet.What + " yet"
			} else {
				if !errors.As(err, &refused) {
					t.Fatalf("got %v, want Refused", err)
				}
				got = filepath.Base(refused.Where) + ": Adamic 0.1 refuses " + refused.What + "; " + refused.Fix
			}
			if got != strings.TrimSpace(string(want)) {
				t.Fatalf("got %q, want %q", got, strings.TrimSpace(string(want)))
			}
		})
	}
}

func TestRefusalPassSoundNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"interface Box { n: number; } interface Box { extra: number; } const box: Box = { n: 1, extra: 2 }; console.log(`${box.extra}`);",
		"class Box { n = 1; } interface Box { n: number; } console.log(`${new Box().n}`);",
		"interface Extra { extra: number; } class Box { n = 1; extra = 2; } interface Box extends Extra {} console.log(`${new Box().extra}`);",
		"function take(value: (n: number) => number): number { return value(1); } console.log(`${take((n) => n + 1)}`);",
		"interface Function { readonly n: number; } const value: Function = { n: 1 }; console.log(`${value.n}`);",
		"type Record<K, T> = { readonly value: T }; const value: Record<string, number> = { value: 1 };",
		"const value: Record<'n', number> = { n: 1 }; console.log(`${value.n}`);",
		"class Function { n = 1; } console.log(`${new Function().n}`);",
		"const value = { extra: 1 }; value.extra = 2;",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("neighbor: %v", err)
		}
	}
}
