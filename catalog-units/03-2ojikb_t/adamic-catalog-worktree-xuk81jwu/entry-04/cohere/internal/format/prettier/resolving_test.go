package prettier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestResolvingFormatsEachFileWithItsOwnConfig is the defect the format phase had, run through the
// replacement: one formatter, two repositories, two configs, and each file gets its own.
//
// The JSX is long enough to break, so bracketSameLine decides where `>` lands. If Resolving applied one
// config to both, the two outputs would put it in the same place.
func TestResolvingFormatsEachFileWithItsOwnConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	apart := filepath.Join(root, "apart")
	together := filepath.Join(root, "together")
	// Each repository's options live in its own Nexus tier, the one file a format block may live in.
	writeFile(t, filepath.Join(apart, "NexusCohereSettings.json"), `{"format": {"printWidth": 60, "tabWidth": 4, "singleQuote": true}}`)
	writeFile(t, filepath.Join(together, "NexusCohereSettings.json"), `{"format": {"printWidth": 60, "tabWidth": 4, "singleQuote": true, "bracketSameLine": true}}`)
	writeFile(t, filepath.Join(apart, "CohereSettings.json"), `{"extends": "./NexusCohereSettings.json"}`)
	writeFile(t, filepath.Join(together, "CohereSettings.json"), `{"extends": "./NexusCohereSettings.json"}`)

	resolving, err := NewResolving(apart)
	if err != nil {
		t.Fatal(err)
	}
	source := "const element = <Component firstAttribute=\"a long value here\" secondAttribute=\"another long value\" third=\"x\">child</Component>;\n"

	formattedApart, err := resolving.Format(filepath.Join(apart, "source", "Probe.tsx"), source)
	if err != nil {
		t.Fatal(err)
	}
	formattedTogether, err := resolving.Format(filepath.Join(together, "source", "Probe.tsx"), source)
	if err != nil {
		t.Fatal(err)
	}

	// JSX attributes keep double quotes under singleQuote (that is jsxSingleQuote's job), so the probe
	// is `"x">`. An earlier draft looked for `'x'>`, which no output contains, so the first check passed
	// vacuously and only the second one's failure exposed it.
	if strings.Contains(formattedApart, `"x">`) {
		t.Errorf("the repository without bracketSameLine put `>` on the attribute line:\n%s", formattedApart)
	}
	if !strings.Contains(formattedTogether, `"x">`) {
		t.Errorf("the repository with bracketSameLine did not put `>` on the attribute line:\n%s", formattedTogether)
	}
	if len(resolving.engines) != 2 {
		t.Errorf("built %d engines for two distinct configs, want 2", len(resolving.engines))
	}
}

// TestResolvingRefusesAtConstructionNotMidRun pins why the first resolution is eager.
func TestResolvingRefusesAtConstructionNotMidRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "prettier.config.js"), "module.exports = {}")
	if _, err := NewResolving(root); err == nil {
		t.Fatal("a directory with a config cohere cannot read built a formatter, so the refusal would land mid-run")
	}
}
