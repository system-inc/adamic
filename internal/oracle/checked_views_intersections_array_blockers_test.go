package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These original programs use the existing boxed array storage. Assert admission
// against Node rather than retaining the source lane's obsolete refusal claim.
func TestCheckedViewIntersectionOriginalArrayBlockers(t *testing.T) {
	t.Parallel()
	declarations, _ := intersectionOriginalInputs(t)
	for _, sample := range []struct{ name, target, refusal string }{
		{"97923", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[]", "an array of IncrementalMultiFileEmitBuildInfoFileInfo"},
		{"98493", "readonly IncrementalMultiFileEmitBuildInfoFileInfo[] | readonly IncrementalBundleEmitBuildInfoFileInfo[]", "an array of string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo"},
	} {
		for _, value := range []struct{ name, expression string }{
			{"string", "'v'"},
			{"intersection-object", "{version: 'v', signature: false, affectsGlobalScope: undefined, impliedFormat: undefined}"},
			{"file-info-object", "{version: 'v', signature: 's', affectsGlobalScope: undefined, impliedFormat: undefined}"},
		} {
			t.Run(sample.name+"/"+value.name, func(t *testing.T) {
				source, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane7/array-admission/"+sample.name+"-"+value.name+".a"))
				if err != nil {
					t.Fatal(err)
				}
				original := fmt.Sprintf("import type { IncrementalMultiFileEmitBuildInfoFileInfo, IncrementalBundleEmitBuildInfoFileInfo } from 'original-tsc-types';\nfunction visit(fileInfos: %s): void { fileInfos.forEach(() => console.log('item')); }\nvisit([%s]);\n", sample.target, value.expression)
				if strings.TrimPrefix(string(source), "// a-check: type error TS2307\n") != original {
					t.Fatal("preserved array program changed")
				}
				bound := strings.Replace(string(source), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))), 1)
				path := filepath.Join(t.TempDir(), "array-"+sample.name+".a")
				if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte("item\n")}, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatalf("expected admitted boxed array storage, got %v", err)
				}
				want := run{stdout: []byte("item\n")}
				actual, binary := nativelyUncached(t, program)
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("admitted original array disagrees with Node: %s; got %#v", difference, got)
					}
				}
				t.Logf("admitted %s/%s agrees with Node; old lane expected %q", sample.name, value.name, sample.refusal)
			})
		}
	}
}
