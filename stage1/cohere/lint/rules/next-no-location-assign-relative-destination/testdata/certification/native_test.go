//go:build lintoracle

// Mounted as a temporary test overlay by run.py. It reuses the shared harness
// without editing it or weakening its native canary policy.
package lint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextjsCertification(t *testing.T) {
	oracle := goOracle(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	selected := os.Getenv("ADAMIC_NEXTJS_RULE")
	for _, descriptor := range prepareRegistry(t, ".") {
		if !strings.HasPrefix(descriptor.Slug, "next-no-") || descriptor.Slug == "next-no-document-import-in-page" || descriptor.Slug == "next-no-sync-scripts" || descriptor.Slug == "next-no-typos" {
			continue
		}
		if selected != "" && selected != descriptor.Slug {
			continue
		}
		t.Run(descriptor.Slug, func(t *testing.T) {
			var rows []string
			for _, row := range upstream(t) {
				fields := strings.Split(row, "\t")
				if len(fields) > 1 && fields[1] == descriptor.Name {
					rows = append(rows, row)
				}
			}
			t.Logf("upstream cases: %d", len(rows))
			rows = append(rows, ownedWitnessRows(t, ".", descriptor)...)
			path := manifest(t, recoveryRows(t, oracle, rows))
			compare(t, oracle, buildPort(t, directory, true), directory, path)
			var change struct{ Name, File, From, To string }
			data, err := os.ReadFile(filepath.Join("rules", descriptor.Slug, "mutant.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &change); err != nil {
				t.Fatal(err)
			}
			if change.File == "" {
				change.File = descriptor.Module
			}
			mutantDirectory := mutant(t, change.From, change.To, filepath.Join("rules", descriptor.Slug, change.File))
			want := execute(t, "", oracle, "--manifest", path).output
			source := node(t, mutantDirectory, path, false).output
			emitted := emittedNode(t, mutantDirectory, path, false).output
			native := execute(t, "", buildMutantPort(t, mutantDirectory), "--manifest", path).output
			for _, side := range []struct {
				name   string
				output []byte
			}{{"Node", source}, {"emitted JavaScript", emitted}, {"sanitized native", native}} {
				if bytes.Equal(side.output, want) {
					t.Fatalf("%s survived on %s", change.Name, side.name)
				}
				if !bytes.Equal(side.output, source) {
					t.Fatalf("mutated %s differs from Node: %s", side.name, difference(side.output, source))
				}
				t.Logf("%s caught on %s", change.Name, side.name)
			}
		})
	}
}
