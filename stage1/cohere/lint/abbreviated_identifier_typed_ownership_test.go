package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	typed "github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/oracles"
)

// RuleContext has no type checker (#17wp0m7). This pin comes out in the change where
// #45rq89s gives it one. Never broaden this exception to other findings.
const contextualKeyGapFinding = ":1:69\n" +
	"  nexus/consistency-no-abbreviated-identifier  Identifier \"durationMs\" should not abbreviate milliseconds as \"Ms\". Use \"durationInMilliseconds\", which is how the rest of the tree spells a millisecond value: \"durationInMilliseconds\" outnumbers \"durationMs\" more than two to one for the identical value, so the rename follows what the codebase already decided rather than introducing a third spelling.\n" +
	"\nrange 68 78 millisecondSuffix \t\t\t68 78\n"

func contextualKeyGapAgreement(path string, got, want []byte) error {
	extra := []byte(path + contextualKeyGapFinding)
	if count := bytes.Count(got, extra); count != 1 {
		return fmt.Errorf("#17wp0m7 pinned 68..78 finding occurs %d times, want exactly one", count)
	}
	if bytes.Contains(want, extra) {
		return fmt.Errorf("#17wp0m7 typed cohere also reports the pinned extra finding")
	}
	withoutExtra := bytes.Replace(got, extra, nil, 1)
	if diff := difference(withoutExtra, want); diff != "" {
		return fmt.Errorf("#17wp0m7 differs beyond the pinned 68..78 finding: %s", diff)
	}
	return nil
}

func TestAbbreviatedIdentifierTypedOwnershipGap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	oracle := typed.TypedOwnership(t, root)
	binary, module := ownedWitnessProducts(t, packageDirectory)
	data, err := os.ReadFile("rules/nexus-consistency-no-abbreviated-identifier/testdata/contextual-key.ts.txt")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "subject.tsx")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	config := `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[]},"files":["subject.tsx"]}`
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	want := ownedWitnessExecute(t, ctx, oracle, path).output
	manifestPath := manifest(t, []string{path + "\tnexus/consistency-no-abbreviated-identifier"})
	runner := filepath.Join(root, "oracle/node.mjs")
	nodeModule := filepath.Join(packageDirectory, "main.ts")
	for _, side := range []struct {
		name, command string
		arguments     []string
	}{
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, nodeModule, "--manifest", manifestPath}},
		{"emitted JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", manifestPath}},
		{"sanitized native", binary, []string{"--manifest", manifestPath}},
	} {
		t.Run(side.name, func(t *testing.T) {
			t.Parallel()
			got := ownedWitnessExecute(t, ctx, side.command, side.arguments...).output
			if err := contextualKeyGapAgreement(path, got, want); err != nil {
				t.Fatal(err)
			}
			t.Log("whole output matches typed cohere after removing exactly the pinned 68..78 finding")
		})
	}
}
