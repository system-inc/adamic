package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Drop the observed backlink in a source mutant. The unchanged Node oracle must catch it.
func TestCyclesDecisionEdgeMutants(t *testing.T) {
	for _, mode := range []string{"weak", "graph"} {
		for _, shape := range []string{"parent", "symbols", "relations"} {
			t.Run(mode+"_"+shape, func(t *testing.T) {
				path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_"+mode+"_"+shape+".a"))
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				edge := map[string]string{"parent": "child.parent = root", "symbols": "declaration.symbol = symbol", "relations": "left.peer = right"}[shape]
				mutant := strings.Replace(string(source), edge, strings.Split(edge, " = ")[0]+" = undefined", 1)
				if mutant == string(source) {
					t.Fatal("mutation did not apply")
				}
				mutantPath := filepath.Join(t.TempDir(), "mutant.a")
				if err := os.WriteFile(mutantPath, []byte(mutant), 0644); err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, mutantPath)
				if err != nil {
					t.Fatal(err)
				}
				binary := filepath.Join(t.TempDir(), "mutant")
				if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
				result := execute(t, binary)
				if oracle.exitCode != 0 || result.exitCode != 0 || disagreement(oracle, result) == "" {
					t.Fatalf("edge mutant escaped: %s", result.stderr)
				}
				if report := leakChecked(t, native.C(program), binary); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}
