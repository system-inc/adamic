package lower

import (
	"context"
	"errors"
	"fmt"
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
			var refused *Refused
			property := "y"
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
	for _, name := range []string{"class", "fresh", "declared", "subclass", "generic_subclass"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/optional_widening_" + name + ".a")
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

// Whole-program ancestry includes imported modules and additional compilation roots.
func TestOptionalWideningWholeProgram(t *testing.T) {
	t.Parallel()
	for _, imported := range []bool{false, true} {
		t.Run(fmt.Sprint(imported), func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			base := filepath.Join(directory, "base.a")
			derived := filepath.Join(directory, "derived.a")
			main := filepath.Join(directory, "main.a")
			files := map[string]string{
				base:    "export class Base { x = 1; }",
				derived: "import { Base } from './base.a'; export class Derived extends Base { y = 'wrong'; }",
				main:    "import { Base } from './base.a'; const view = new Base(); const wider: { x: number; y?: number } = view;",
			}
			if imported {
				files[main] = "import './derived.a'; " + files[main]
			}
			for path, source := range files {
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
			roots := []string{main}
			if !imported {
				roots = append(roots, derived)
			}
			program, err := load.Load(roots)
			if err != nil {
				t.Fatal(err)
			}
			if imported {
				_, err = Lower(context.Background(), program)
			} else {
				// Lower accepts one entry. Exercise the rule directly on a checker program
				// with multiple roots, as load and developer inventories permit.
				file := program.Files()[0]
				checker, release := program.Checker(context.Background(), file)
				defer release()
				l := &lowering{checker: checker, program: program}
				err = l.refuse(file)
			}
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "subclass Derived declares y") || !strings.Contains(refused.What, "derived.a:1:") {
				t.Fatalf("want imported subclass refusal, got %v", err)
			}
		})
	}
}

// The coverage reproducer has an inhabited outer value and an impossible union arm.
// Exercise the relation pass directly: declare supplies no runtime initializer.
func TestOptionalWideningReducedSource(t *testing.T) {
	program, err := load.Load([]string{"testdata/reduced_optional_union.a"})
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
