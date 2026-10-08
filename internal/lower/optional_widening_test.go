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

func TestOptionalWideningRefused(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/optional_widening/*.a")
	if err != nil || len(paths) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	for _, path := range paths {
		t.Run(strings.TrimSuffix(filepath.Base(path), ".a"), func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{path})
			if filepath.Base(path) == "incompatible.a" {
				var checked *load.CheckError
				if !errors.As(err, &checked) || !strings.Contains(err.Error(), "not assignable") {
					t.Fatalf("want checker rejection, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			if filepath.Base(path) == "field_access_paths.a" {
				// The retained presence-side proof knows this unannotated const's exact keys.
				if err != nil {
					t.Fatalf("exact-key alias must lower: %v", err)
				}
				return
			}
			var refused *Refused
			property := "y"
			if filepath.Base(path) == "field_access_paths.a" {
				property = "amount"
			}
			if filepath.Base(path) == "spread_other_missing.a" {
				property = "z"
			}
			cast := filepath.Base(path) == "cast.a"
			if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "adamic/no-optional-widening") || (!cast && !strings.Contains(refused.What, "optional property "+property)) || (cast && !strings.Contains(refused.What, "an unproven relation")) {
				t.Fatalf("want optional-property refusal, got %v", err)
			}
			// Casts reach main's proven-relation refusal first. Other positions must retain
			// the optional-widening diagnostic. Pin every complete message independently.
			want, readErr := os.ReadFile(strings.TrimSuffix(path, ".a") + ".refused")
			if readErr != nil {
				t.Fatal(readErr)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), absolute, path)
			if got != strings.TrimSpace(string(want)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(want)))
			}
			if filepath.Base(path) == "initializer.a" {
				want := "testdata/optional_widening/initializer.a:3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
				if !strings.HasSuffix(refused.Error(), want) {
					t.Errorf("got %q, want %q", refused.Error(), want)
				}
			}
		})
	}
}

func TestOptionalWideningAllowed(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"class", "fresh", "declared", "field_access_paths"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := "../oracle/testdata/optional_widening_" + name + ".a"
			if name == "field_access_paths" {
				path = "../oracle/testdata/field_access_paths.a"
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lowerSource(t, string(source)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// These are allowed relations even though adding a spread field is not lowered yet. Test the
// refusal pass directly so a NotYet in the backend cannot conceal a wrong optional-field refusal.
func TestOptionalWideningSpreadOverwrite(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "const original = { x: 1, y: 'wrong', z: 'wrong' }; const view: { x: number } = original; const wider: { x: number; y?: number } = { ...view, y: 3 };"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: checker, program: program}
	if err := l.refuse(file); err != nil {
		t.Fatal(err)
	}
}

func TestOptionalWideningIsRefused(t *testing.T) {
	t.Parallel()
	const hidden = `const wide: { name: string; extra: number } = { name: 'a', extra: 1 };
const view: { name: string } = wide;
`
	for _, probe := range []struct{ name, source string }{
		{"assignment", hidden + `const narrow: { name: string; extra?: string } = view; console.log(narrow.extra ?? 'none');`},
		{"explicit type argument", hidden + `function read<T extends { readonly name: string; readonly extra?: string }>(value: T): string { return value.extra ?? 'none'; } console.log(read<{ readonly name: string; readonly extra?: string }>(view));`},
		{"argument", hidden + `function read(value: { name: string; extra?: string }): string { return value.extra ?? 'none'; } console.log(read(view));`},
		{"array element", hidden + `const values: { name: string }[] = [view]; const narrow: { name: string; extra?: string }[] = values; console.log(narrow.length.toFixed(0));`},
		{"source union member", `function read(value: { readonly name: string } | { readonly name: string; readonly extra: string }): string { const narrow: { readonly name: string; readonly extra?: string } = value; return narrow.extra ?? 'none'; } console.log(read({ name: 'a' }));`},
		{"target union member", hidden + `const narrow: { name: string; extra?: string } | { name: string; extra?: boolean } = view; console.log(narrow.name);`},
		{"conditional aliases", `function widen(flag: boolean): { name: string; extra?: string } { const first = { name: 'a' }; const second = { name: 'b' }; return flag ? first : second; } console.log(widen(true).name);`},
		{"return", `function widen(value: { name: string }): { name: string; extra?: string } { return value; } console.log(widen({ name: 'a' }).name);`},
		{"field", hidden + `class Box { readonly value: { name: string; extra?: string } = view; } console.log(new Box().value.name);`},
		{"readonly", hidden + `const narrow: { readonly name: string; readonly extra?: string } = view; console.log(narrow.name);`},
		{"nested alias", hidden + `const box = { value: view }; const narrow: { readonly value: { name: string; extra?: string } } = box; console.log(narrow.value.name);`},
		{"shorthand", hidden + `const narrow: { readonly view: { name: string; extra?: string } } = { view }; console.log(narrow.view.name);`},
		{"spread field", hidden + `const box = { view }; const narrow: { readonly view: { name: string; extra?: string } } = { ...box }; console.log(narrow.view.name);`},
		{"destructuring", hidden + `let narrow: { name: string; extra?: string } = { name: 'b' }; const pair: [{ name: string }] = [view]; [narrow] = pair; console.log(narrow.name);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want optional widening refused, got %v", err)
			}
			for _, want := range []string{"adamic/no-optional-widening", "extra", "declare", "source type", "fresh object", "known fields"} {
				if !strings.Contains(refused.Error(), want) {
					t.Errorf("refusal %q does not contain %q", refused.Error(), want)
				}
			}
		})
	}
}

func TestOptionalWideningSoundNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const source: { name: string; extra?: string } = { name: 'a' }; const view: { name: string; extra?: string } = source; console.log(view.extra ?? 'none');`,
		`const view: { name: string; extra?: string } = { name: 'a' }; console.log(view.extra ?? 'none');`,
		`const source = { name: 'a' }; const view: { readonly name: string; readonly extra?: string } = source; console.log(view.extra ?? 'none');`,
		`const source: { readonly name: string; readonly extra?: string } = { name: 'a', extra: 'yes' }; const view: { readonly name: string; readonly extra?: string } = source; console.log(view.extra ?? 'none');`,
		`function fresh(flag: boolean): { name: string; extra?: string } { return flag ? { name: 'a' } : { name: 'b' }; } console.log(fresh(true).name);`,
		`function keep(value: { name: string }): { name: string } | { name: string; extra?: string } { return value; } console.log(keep({ name: 'a' }).name);`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("sound neighbor: %v", err)
		}
	}
}
