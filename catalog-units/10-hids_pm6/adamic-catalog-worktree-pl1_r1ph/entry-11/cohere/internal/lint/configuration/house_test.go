package configuration

import (
	"path/filepath"
	"strings"
	"testing"
)

// Zero config is the house stack, each per-file set where the program shows it fits (#bfxz13m,
// #j6p9t1g). These use stand-in sets so each fact is about the loader, not about today's rulings.

func withHouseSetsForTest(t *testing.T) {
	t.Helper()
	withSetFiles(t, map[string]string{
		"typescript": `{"rules": {"no-var": "error", "eqeqeq": ["error", "always"]}, "overrides": [{"files": ["**/*.generated.*"], "rules": {"no-var": "off"}}]}`,
		"react":      `{"extends": "cohere:typescript", "rules": {"react/no-danger": "error"}}`,
		"next":       `{"extends": "cohere:typescript", "rules": {"@next/next/no-img-element": "error"}, "ignorePatterns": ["**/.next/**"]}`,
		"tailwind":   `{"extends": "cohere:typescript", "rules": {"better-tailwindcss/no-unknown-classes": "error"}}`,
	})
}

// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestZeroConfigAppliesEachPerFileSetOnlyWhereItFits(t *testing.T) {
	withHouseSetsForTest(t)
	root := t.TempDir()
	component := filepath.Join(root, "source", "Button.tsx")
	route := filepath.Join(root, "app", "page.tsx")
	server := filepath.Join(root, "server", "index.ts")
	loaded, err := LoadHouse(filepath.Join(root, "CohereSettings.json"), nil, HouseDetection{
		ReactFiles:      map[string]bool{component: true, route: true},
		NextFiles:       map[string]bool{route: true},
		TailwindSkipped: "no Tailwind stylesheet",
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		file    string
		enabled []string
		off     []string
	}{
		{server, []string{"no-var", "eqeqeq"}, []string{"react/no-danger", "@next/next/no-img-element", "better-tailwindcss/no-unknown-classes"}},
		{component, []string{"no-var", "react/no-danger"}, []string{"@next/next/no-img-element", "better-tailwindcss/no-unknown-classes"}},
		{route, []string{"no-var", "react/no-danger", "@next/next/no-img-element"}, []string{"better-tailwindcss/no-unknown-classes"}},
	}
	for _, testCase := range cases {
		resolved := loaded.Resolve(testCase.file)
		for _, name := range testCase.enabled {
			if !resolved.Enabled(name) {
				t.Errorf("%s: %s should run", testCase.file, name)
			}
		}
		// Off with a reason, not unconfigured: a set that does not fit says why its rules did not run.
		for _, name := range testCase.off {
			if status, _ := resolved.StatusOf(name); status != StatusScopedOff {
				t.Errorf("%s: %s is %v, want off", testCase.file, name, status)
			}
		}
	}

	// The whole stack is what the run asks for, and each gated rule's reason names its set.
	for name, want := range map[string]string{
		"react/no-danger":                       "cohere:react applies only to files that import react",
		"better-tailwindcss/no-unknown-classes": "cohere:tailwind is not applied: no Tailwind stylesheet",
	} {
		if loaded.Rules[name].Severity != SeverityError {
			t.Errorf("the whole stack does not ask for %s", name)
		}
		if reason, found := loaded.OffReasonFor(name); !found || !strings.Contains(reason.Reason, want) {
			t.Errorf("%s's reason is %+v, want one containing %q", name, reason, want)
		}
	}

	// A set's ignore pattern is the house's, for every file, whichever set wrote it.
	if !loaded.Resolve(filepath.Join(root, ".next", "server", "chunk.ts")).Ignored {
		t.Error("cohere:next's ignore pattern must hold for a file no per-file set fits")
	}

	want := "sets: cohere:typescript on every file; cohere:react on 2 files that import react or contain JSX; " +
		"cohere:next on 1 file that imports next or is one of Next's own files; cohere:tailwind not applied, no Tailwind stylesheet"
	if line := loaded.SetsLine(); line != want {
		t.Errorf("first line\n  got:  %s\n  want: %s", line, want)
	}
}

// The control for the per-file evidence: with no file importing react, no file gets it, whatever else is
// true of the project, and the first line says so.
// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestZeroConfigWithNoReactFileAppliesNoReactRule(t *testing.T) {
	withHouseSetsForTest(t)
	root := t.TempDir()
	loaded, err := LoadHouse(filepath.Join(root, "CohereSettings.json"), nil, HouseDetection{TailwindSkipped: "none"})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Resolve(filepath.Join(root, "View.tsx")).Enabled("react/no-danger") {
		t.Error("react rules ran with no file importing react")
	}
	if !strings.Contains(loaded.SetsLine(), "cohere:react not applied, no file imports react or contains JSX") {
		t.Errorf("the first line must name react as not applied: %s", loaded.SetsLine())
	}
}

// Tailwind is project-wide: with the stylesheet found, every file gets it.
// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestZeroConfigAppliesTailwindEverywhereWhenTheStylesheetIsTailwinds(t *testing.T) {
	withHouseSetsForTest(t)
	root := t.TempDir()
	loaded, err := LoadHouse(filepath.Join(root, "CohereSettings.json"), nil, HouseDetection{TailwindEntryPoint: filepath.Join(root, "app", "globals.css")})
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Resolve(filepath.Join(root, "server", "index.ts")).Enabled("better-tailwindcss/no-unknown-classes") {
		t.Error("tailwind rules must run on every file once the stylesheet is Tailwind's")
	}
	if !strings.Contains(loaded.SetsLine(), "cohere:tailwind on every file, since app/globals.css is a Tailwind stylesheet") {
		t.Errorf("first line: %s", loaded.SetsLine())
	}
}

// An outsider's own file goes on top of the detected sets: its off needs no reason, its options replace
// the set's, and a rule it writes that a set also writes is its choice rather than a collision.
// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestZeroConfigPutsTheProjectsOwnFileOnTop(t *testing.T) {
	withHouseSetsForTest(t)
	directory := writeConfigs(t, map[string]string{
		"CohereSettings.json": `{"rules": {"no-var": "off", "eqeqeq": ["error", "smart"], "react/no-danger": "warn"}}`,
	})
	path := filepath.Join(directory, "CohereSettings.json")
	if zero, err := IsZeroConfig(path); err != nil || !zero {
		t.Fatalf("a file that names no set is zero config: %v %v", zero, err)
	}
	component := filepath.Join(directory, "View.tsx")
	loaded, err := LoadHouse(path, nil, HouseDetection{ReactFiles: map[string]bool{component: true}})
	if err != nil {
		t.Fatal(err)
	}
	resolved := loaded.Resolve(component)
	if resolved.Enabled("no-var") {
		t.Error("the project's off did not apply")
	}
	if status, setting := resolved.StatusOf("eqeqeq"); status != StatusEnabled || string(setting.Options[0]) != `"smart"` {
		t.Errorf("eqeqeq resolved %v %s, want the project's option", status, setting.Options)
	}
	if _, setting := resolved.StatusOf("react/no-danger"); setting.Severity != SeverityWarn {
		t.Errorf("the project's severity for a react rule did not apply: %v", setting.Severity)
	}
	if _, reasoned := loaded.OffReasonFor("no-var"); reasoned {
		t.Error("an outsider's unreasoned off must stay unreasoned, so coverage counts it")
	}
}

// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestAChainThatNamesASetIsNotZeroConfig(t *testing.T) {
	withHouseSetsForTest(t)
	directory := writeConfigs(t, map[string]string{
		"base.json":           `{"extends": "cohere:react"}`,
		"CohereSettings.json": `{"extends": "./base.json"}`,
	})
	if zero, err := IsZeroConfig(filepath.Join(directory, "CohereSettings.json")); err != nil || zero {
		t.Fatalf("a chain naming a set chose its sets: zero=%v err=%v", zero, err)
	}
	if zero, err := IsZeroConfig(filepath.Join(directory, "Missing.json")); err != nil || !zero {
		t.Fatalf("no settings file is zero config: zero=%v err=%v", zero, err)
	}
}

// A set's override is a shape, not a path in this project, so a project with no generated file is not a
// broken config; the project's own override is still held to reaching a file.
// Not parallel: it swaps the package's setFiles for stand-in sets through withHouseSetsForTest
// (withSetFiles), which every test that reads a set would see.
func TestSelectorValidationSkipsTheSetsOwnOverrides(t *testing.T) {
	withHouseSetsForTest(t)
	directory := writeConfigs(t, map[string]string{
		"CohereSettings.json": `{"overrides": [{"files": ["missing/**"], "rules": {"no-var": "warn"}}]}`,
	})
	loaded, err := LoadHouse(filepath.Join(directory, "CohereSettings.json"), nil, HouseDetection{})
	if err != nil {
		t.Fatal(err)
	}
	err = loaded.ValidateSelectors([]string{filepath.Join(directory, "index.ts")})
	if err == nil || strings.Contains(err.Error(), "generated") || !strings.Contains(err.Error(), "missing/**") {
		t.Errorf("want only the project's own vacuous selector refused, got %v", err)
	}
}
