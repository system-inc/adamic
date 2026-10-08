package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Complete original aliases keep their source arrays while consumers check each
// selected scalar or structural element before invoking the callback.
func TestCheckedViewIntersectionOriginalMixedArrays(t *testing.T) {
	declarations, _ := intersectionOriginalInputs(t)
	for _, sample := range []struct{ name, target string }{
		{"97923", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[]"},
		{"98493", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[] | readonly IncrementalBundleEmitBuildInfoFileInfo[]"},
	} {
		for _, value := range []struct{ name, expression string }{
			{"string", "'v'"},
			{"intersection-object", "{version: 'v', signature: false, affectsGlobalScope: undefined, impliedFormat: undefined}"},
			{"file-info-object", "{version: 'v', signature: 's', affectsGlobalScope: undefined, impliedFormat: undefined}"},
		} {
			t.Run(sample.name+"/"+value.name, func(t *testing.T) {
				source := fmt.Sprintf("import type { IncrementalMultiFileEmitBuildInfoFileInfo, IncrementalBundleEmitBuildInfoFileInfo } from %q;\nfunction visit(fileInfos: %s): void { fileInfos.forEach(() => console.log('item')); }\nvisit([%s]);\n", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts")), sample.target, value.expression)
				path := filepath.Join(t.TempDir(), "array-"+sample.name+".a")
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte("item\n")}, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				want := run{stdout: []byte("item\n")}
				actual, binary := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Errorf("%s; stderr %q", diff, got.stderr)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}
