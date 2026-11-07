package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestOptionalWideningRefused(test *testing.T) {
	test.Parallel()
	paths, err := filepath.Glob("testdata/optional_widening/*.a")
	if err != nil || len(paths) == 0 {
		test.Fatalf("fixtures: %v", err)
	}
	for _, path := range paths {
		test.Run(strings.TrimSuffix(filepath.Base(path), ".a"), func(test *testing.T) {
			test.Parallel()
			program, err := load.Load([]string{path})
			if filepath.Base(path) == "incompatible.a" {
				var checked *load.CheckError
				if !errors.As(err, &checked) || !strings.Contains(err.Error(), "not assignable") {
					test.Fatalf("want checker rejection, got %v", err)
				}
				return
			}
			if err != nil {
				test.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			var refused *Refused
			property := "y"
			if filepath.Base(path) == "field_access_paths.a" {
				property = "amount"
			}
			if filepath.Base(path) == "spread_other_missing.a" {
				property = "z"
			}
			cast := filepath.Base(path) == "cast.a"
			if !errors.As(err, &refused) ||
				!strings.Contains(refused.Fix, "adamic/no-optional-widening") ||
				(!cast && !strings.Contains(refused.What, "optional property "+property)) ||
				(cast && !strings.Contains(refused.What, "an unproven relation")) {
				test.Fatalf("want optional-property refusal, got %v", err)
			}
			// Casts reach main's proven-relation refusal first. Other positions must retain
			// the optional-widening diagnostic. Pin every complete message independently.
			want, readErr := os.ReadFile(strings.TrimSuffix(path, ".a") + ".refused")
			if readErr != nil {
				test.Fatal(readErr)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				test.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), absolute, path)
			if got != strings.TrimSpace(string(want)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(want)))
			}
			if filepath.Base(path) == "initializer.a" {
				want := "testdata/optional_widening/initializer.a:3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
				if !strings.HasSuffix(refused.Error(), want) {
					test.Errorf("got %q, want %q", refused.Error(), want)
				}
			}
		})
	}
}

// These are allowed relations even though adding a spread field is not lowered yet. Test the
// refusal pass directly so a NotYet in the backend cannot conceal a wrong optional-field refusal.
func TestOptionalWideningSpreadOverwrite(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := "const original = { x: 1, y: 'wrong', z: 'wrong' }; const view: { x: number } = original; const wider: { x: number; y?: number } = { ...view, y: 3 };"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: checker, program: program}
	if err := l.refuse(file); err != nil {
		test.Fatal(err)
	}
}
