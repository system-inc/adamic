package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewIntersectionOriginalAbsorption(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, name := range []string{"good", "wrong"} {
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane7/original/jsdoc-parent-good.a")
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(input), "JSDoc, SyntaxKind", "JSDoc, JSDocTypeLiteral, SyntaxKind", 1)
			source = strings.Replace(source, "jsDoc: JSDoc)", "jsDoc: JSDoc | JSDocTypeLiteral)", 1)
			source = strings.Replace(source, "kind: 243", "kind: 308", 1)
			source = strings.Replace(source, "parent.end > parent.pos", "parent.end > 0", 1)
			if name == "wrong" {
				source = strings.Replace(source, "pos: 20", "pos: 'wrong'", 1)
			}
			source = strings.Replace(source, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			path := filepath.Join(t.TempDir(), "jsdoc-union-parent-"+name+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			requireIntersectionOriginalComplete(t, program, manifest, []string{"Node", "JSDoc", "JSDocTypeLiteral"}, 7644)
			found := false
			for i, c := range program.ViewContracts {
				if !strings.Contains(c.Name, "Node |") || !c.IntersectionBounded {
					continue
				}
				found = true
				if c.Kind != ir.ViewObject || len(c.Fields) != len(manifest.Fields["Node"]) {
					t.Fatal("ancestor obligations lost")
				}
				if os.Getenv("ADAMIC_INTERSECTION_ABSORPTION_MUTANT") != "" {
					for j, f := range c.Fields {
						if f.Name == "pos" {
							program.ViewContracts[i].Fields = append(append([]ir.ViewFieldContract{}, c.Fields[:j]...), c.Fields[j+1:]...)
							break
						}
					}
				}
			}
			if !found {
				t.Fatal("union was not absorbed with bounded ancestor obligations")
			}
			actual, binary := nativelyUncached(t, program)
			if name == "good" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "good" {
					if difference := disagreement(run{stdout: []byte("true\n")}, got); difference != "" {
						t.Error(difference)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "jsDoc.parent.pos") {
					t.Errorf("missing named exit 70: %#v", got)
				}
			}
		})
	}
}
