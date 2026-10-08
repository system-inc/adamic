package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func mixedArrayJointFile(t *testing.T, declarations, name string) string {
	t.Helper()
	input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane2/joints/mixed/" + name + ".a"))
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.ReplaceAll(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))))
	path := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckedViewOriginalMixedArrayJoints(t *testing.T) {
	declarations, _ := intersectionOriginalInputs(t)
	data, err := os.ReadFile(filepath.Join(declarations, "mixed-array-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fields map[string][]string
		Pairs  []struct {
			Reads int `json:"reads"`
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Pairs) != 5 {
		t.Fatal("mixed array census drift")
	}
	for _, p := range manifest.Pairs {
		if p.Reads != 1 {
			t.Fatal("mixed array read count drift")
		}
	}
	for _, family := range []struct{ name, contract string }{
		{"bundle", "IncrementalBundleEmitBuildInfoFileInfo"},
		{"multi", "IncrementalMultiFileEmitBuildInfoFileInfo"},
		{"union", "string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo"},
	} {
		for _, sample := range []struct{ name, source, found string }{
			{"string", "string\ndone\n", ""},
			{"object", "object\ndone\n", ""},
			{"alias", "second\ndone\n", ""},
			{"lazy-element", "found\n", ""},
			{"number", "number\ndone\n", "number"},
			{"wrong-version", "object\ndone\n", "object"},
			{"missing-version", "object\ndone\n", "object"},
		} {
			t.Run(family.name+"-"+sample.name, func(t *testing.T) {
				file := mixedArrayJointFile(t, declarations, family.name+"-"+sample.name)
				node := run{stdout: []byte(sample.source)}
				if diff := disagreement(node, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range program.ViewContracts {
					expected, ok := manifest.Fields[c.Name]
					if !ok {
						continue
					}
					fields := []string{}
					for _, f := range c.Fields {
						fields = append(fields, f.Name)
					}
					slices.Sort(fields)
					if !slices.Equal(fields, expected) {
						t.Fatal("original field set reduced: " + c.Name)
					}
				}
				want := node
				if sample.found != "" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.fileInfos[element] matches no member of " + family.contract + "; expected " + family.contract + ", found " + sample.found + "\n")}
				}

				actual, binary := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Errorf("%s; got %#v", diff, got)
					}
				}
				if sample.found == "" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
					return
				}
				// Mutate only the demanded logical descriptor. Storage and consumer code
				// remain identical, and every mutant must finish with the source Node output.
				changed := 0
				if sample.name == "number" {
					number := ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
					for i := range program.ViewContracts {
						c := &program.ViewContracts[i]
						if ir.MixedArrayContract(program, ir.ViewContractID(i+1)) {
							c.Members = append(c.Members, number)
							changed++
						}
					}
				} else {
					for i := range program.ViewContracts {
						c := &program.ViewContracts[i]
						if c.Kind != ir.ViewObject {
							continue
						}
						fields := []ir.ViewFieldContract{}
						for _, f := range c.Fields {
							if f.Name == "version" {
								changed++
								continue
							}
							fields = append(fields, f)
						}
						c.Fields = fields
					}
				}
				if changed == 0 {
					t.Fatal("mutant changed no descriptor")
				}
				mutated, mutantBinary := nativelyUncached(t, program)
				for _, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if disagreement(want, got) == "" {
						t.Fatal("mutant escaped the stopping check")
					}
					if diff := disagreement(node, got); diff != "" {
						t.Errorf("mutant must execute Node output: %s; got %#v", diff, got)
					}
				}
				if report := leaksUncached(t, program, mutantBinary); report != "" {
					t.Fatal(report)
				}
				t.Logf("descriptor mutant caught by pinned %s rejection in all three modes", sample.found)
			})
		}
	}
}

// Not parallel: only these measured labels change in the shared count table.
func TestCheckedViewMixedArrayJointCounts(t *testing.T) {
	declarations, _ := intersectionOriginalInputs(t)
	root, err := filepath.Abs(checkedViewFixturePath(repository))
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{}
	for _, family := range []string{"bundle", "multi", "union"} {
		for _, variant := range []string{"string", "object", "alias", "lazy-element", "number", "wrong-version", "missing-version"} {
			name := family + "-" + variant
			path := mixedArrayJointFile(t, declarations, name)
			relative, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			row := counted(t, checkedViewFixturePath(relative), false, nil, false, false)
			label := "stage3/interface-downcasts/lane2/joints/mixed/" + name + ".a"
			rows = append(rows, strings.Replace(row, relative, label, 1))
		}
	}
	path := checkedViewFixturePath(filepath.Join(repository, "internal/oracle/counts.md"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded counts: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
	lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
	for _, row := range rows {
		key := strings.Split(row, " | ")[0] + " | "
		found := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i] = row
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, row)
		}
	}
	result := strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		result += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		t.Fatal(err)
	}
}
