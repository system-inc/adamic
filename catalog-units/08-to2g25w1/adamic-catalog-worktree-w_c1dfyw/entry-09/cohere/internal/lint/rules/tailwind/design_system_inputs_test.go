package tailwind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// The run cache's input recorder sees every stylesheet the design system read, and every entry-point
// candidate it probed and missed.
//
// The loader used to ask `os` directly, beneath the program's filesystem where the recorder sits, so
// none of it was an input to the run cache. An ignored or untracked stylesheet the theme imports,
// edited in place, replayed the old verdict (#ym4v8bc). It asks the program's filesystem now.
//
// The fixture's entry is the second candidate, `app/globals.css`, so the first,
// `app/_theme/styles/theme.css`, is probed and missed and must be recorded absent: creating it would
// change which stylesheet is the entry. The entry imports `extra.css`, which only the engine reads,
// and which must be recorded present.
func TestDesignSystemReadsAreRunCacheInputs(t *testing.T) {
	t.Parallel()
	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so there is no design system to load")
	}

	directory := t.TempDir()
	files := map[string]string{
		"tsconfig.json":   `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "jsx": "preserve", "types": []}, "include": ["**/*.tsx"]}`,
		"Component.tsx":   `export const element = <div className="flex" />;`,
		"app/globals.css": "@import \"tailwindcss\";\n@import \"./extra.css\";\n",
		"app/extra.css":   "@utility probe-utility {\n    color: red;\n}\n",
	}
	for name, contents := range files {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(directory, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(packageRoot, filepath.Join(directory, "node_modules", "tailwindcss")); err != nil {
		t.Fatal(err)
	}

	recorder := program.NewInputRecorder()
	graph, err := program.Build(program.Options{
		ConfigFileName:   filepath.Join(directory, "tsconfig.json"),
		CurrentDirectory: directory,
		SingleThreaded:   true,
		Inputs:           recorder,
	})
	if err != nil {
		t.Fatalf("building the program: %v", err)
	}

	result := DesignSystemForProgram(designSystemProgram(graph.Program))
	if result.Err != nil {
		t.Fatalf("the design system did not load, so this proves nothing: %v", result.Err)
	}

	present, absent, _ := recorder.Inputs()
	hasSuffix := func(paths []string, suffix string) bool {
		for _, path := range paths {
			if strings.HasSuffix(path, suffix) {
				return true
			}
		}
		return false
	}

	for _, stylesheet := range []string{"/app/globals.css", "/app/extra.css"} {
		if !hasSuffix(present, stylesheet) {
			t.Errorf("%s was read by the design system and is not a recorded input, so editing it would "+
				"replay a stale run", stylesheet)
		}
	}
	if !hasSuffix(absent, "/app/_theme/styles/theme.css") {
		t.Errorf("the first entry-point candidate was probed and missed and is not recorded absent, so " +
			"creating it would not invalidate a run whose entry it would replace")
	}
	if hasSuffix(present, "/app/_theme/styles/theme.css") {
		t.Errorf("a candidate that does not exist is recorded present")
	}

	// The design system's own read set, which the findings cache keys the Tailwind rules on
	// (#pyhm2t2), holds the same facts: the stylesheets present, the missed candidate absent. Every
	// path in it is one the run cache's recorder also saw, so the two observations agree.
	readPresent, readAbsent := []string{}, []string{}
	for _, read := range result.Reads {
		if read.Present {
			readPresent = append(readPresent, read.Path)
		} else {
			readAbsent = append(readAbsent, read.Path)
		}
	}
	for _, stylesheet := range []string{"/app/globals.css", "/app/extra.css"} {
		if !hasSuffix(readPresent, stylesheet) {
			t.Errorf("%s is missing from the design system's read set, so a cache keyed on it would replay "+
				"the Tailwind rules after it changed", stylesheet)
		}
	}
	if !hasSuffix(readAbsent, "/app/_theme/styles/theme.css") {
		t.Errorf("the missed entry-point candidate is missing from the design system's read set as absent")
	}
	recorded := map[string]bool{}
	for _, path := range append(present, absent...) {
		recorded[path] = true
	}
	for _, read := range result.Reads {
		if !recorded[read.Path] {
			t.Errorf("the design system read %s and the run cache's recorder did not see it", read.Path)
		}
	}
}
