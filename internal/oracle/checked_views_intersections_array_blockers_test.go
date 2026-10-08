package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These original aliases need boxed mixed-element array storage before a view
// certificate can be reached. Keep their Node behavior and named refusal pinned.
func TestCheckedViewIntersectionOriginalArrayBlockers(t *testing.T) {
	declarations, _ := intersectionOriginalInputs(t)
	for _, sample := range []struct{ name, target, refusal string }{
		{"97923", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[]", "an array of IncrementalMultiFileEmitBuildInfoFileInfo"},
		{"98493", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[] | readonly IncrementalBundleEmitBuildInfoFileInfo[]", "an array of string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			source := fmt.Sprintf("import type { IncrementalMultiFileEmitBuildInfoFileInfo, IncrementalBundleEmitBuildInfoFileInfo } from %q;\nfunction visit(fileInfos: %s): void { fileInfos.forEach(() => console.log('item')); }\nvisit(['v']);\n", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts")), sample.target)
			path := filepath.Join(t.TempDir(), "array-"+sample.name+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("item\n")}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			_, err := lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), sample.refusal) {
				t.Fatalf("expected original mixed-element storage refusal, got %v", err)
			}
			t.Log(err)
		})
	}
}
