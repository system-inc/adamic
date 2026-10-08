package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fileNamesJointFile(t *testing.T, declarations, name string) string {
	t.Helper()
	source, err := os.ReadFile("../../stage3/interface-downcasts/lane2/joints/file-names/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.ReplaceAll(string(source), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))))
	file := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestCheckedViewFileNamesJoint(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	data, err := os.ReadFile(filepath.Join(declarations, "file-names-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fields map[string][]string
		Pairs  []struct {
			ReadCount int `json:"read_count"`
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Pairs) != 1 || manifest.Pairs[0].ReadCount != 4 {
		t.Fatal("original read census drift")
	}
	for _, sample := range []struct{ name, output string }{{"join", "a;b\n"}, {"four-reads", "a;b\na;b\na;b\na;b\n"}, {"alias", "a;second\n"}, {"lazy", "found\n"}, {"unread", "kept\n"}, {"wrong-element", "a;42\n"}, {"wrong-array", "string\n"}, {"missing-array", "undefined\n"}} {
		t.Run(sample.name, func(t *testing.T) {
			file := fileNamesJointFile(t, declarations, sample.name)
			node := run{stdout: []byte(sample.output)}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			for name, expected := range manifest.Fields {
				found := false
				for _, c := range program.ViewContracts {
					if c.Name == name {
						fields := []string{}
						for _, f := range c.Fields {
							fields = append(fields, f.Name)
						}
						slices.Sort(fields)
						found = found || slices.Equal(fields, expected)
					}
				}
				if !found {
					t.Fatal("original field set reduced: " + name)
				}
			}
			want := node
			message := ""
			switch sample.name {
			case "wrong-element":
				message = "element read failed: node.fileNames[element] expected string, found number"
			case "wrong-array":
				message = "field read failed: node.fileNames is not a readonly string[]; expected readonly string[], found string"
			case "missing-array":
				message = "field read failed: node.fileNames is not initialized; expected readonly string[], found missing"
			}
			if message != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + message + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			if message == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			changed := 0
			stringID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.String, Name: "string"})
			switch sample.name {
			case "wrong-element":
				numberID := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
				unionID := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Union, Name: "string | number", Members: []ir.ViewContractID{stringID, numberID}})
				rewrite := func(n any) any {
					if read, ok := n.(ir.ArrayJoin); ok {
						read.Element = ir.Union
						read.Stringify = false
						read.ViewRead.Element = ir.Union
						read.ViewRead.ViewContract = unionID
						changed++
						return read
					}
					return n
				}
				for i := range program.Functions {
					program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
				}
			case "wrong-array", "missing-array":
				changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "node.fileNames" }, func(p ir.Property) ir.Property {
					p.Of = ir.String
					p.ViewContract = stringID
					p.ViewType = "string"
					if sample.name == "missing-array" {
						p.Absent = true
						p.Optional = true
					}
					return p
				})
			}
			if changed == 0 {
				t.Fatal("mutant changed no read")
			}
			mutated, mutantBinary := nativelyUncached(t, program)
			for _, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatal("mutant escaped stopping oracle")
				}
				if diff := disagreement(node, got); diff != "" {
					t.Errorf("mutant must finish with Node output: %s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			if report := leaksUncached(t, program, mutantBinary); report != "" {
				t.Fatal(report)
			}
			t.Log("mutant caught by pinned stopping oracle in all three modes")

		})
	}
}

func TestCheckedViewFileNamesJointCounts(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	root, err := filepath.Abs(checkedViewFixturePath(repository))
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{}
	for _, name := range []string{"join", "four-reads", "alias", "lazy", "unread", "wrong-element", "wrong-array", "missing-array"} {
		path := fileNamesJointFile(t, declarations, name)
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, checkedViewFixturePath(relative), false, nil, false, false)
		label := "stage3/interface-downcasts/lane2/joints/file-names/" + name + ".a"
		rows = append(rows, strings.Replace(row, relative, label, 1))
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

func TestCheckedViewFileNamesWholeUnionBoundary(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	file := fileNamesJointFile(t, declarations, "join")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	head := strings.Split(string(data), "function read")[0]
	source := head + "interface Wrapper extends Base { readonly item: IncrementalBuildInfo; }\nfunction read(node: Wrapper): string { return node.item.fileNames.join(';'); }\nconst raw = {version: '1', item: {version: '1', fileNames: ['a','b']}};\nconst base: Base = raw;\nconsole.log(read(base as Wrapper));\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("a;b\n")}, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.item matches no member of IncrementalBuildInfo; expected IncrementalBuildInfo, found object\n")}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("whole member proof: %s; stderr %q", diff, got.stderr)
		}
	}
	changed := 0
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		if c.Name != "IncrementalBundleEmitBuildInfo" && c.Name != "IncrementalMultiFileEmitBuildInfo" {
			continue
		}
		fields := []ir.ViewFieldContract{}
		for _, f := range c.Fields {
			if f.Name == "fileNames" || f.Name == "version" {
				fields = append(fields, f)
			} else {
				changed++
			}
		}
		c.Fields = fields
	}
	if changed == 0 {
		t.Fatal("whole member mutant changed no field")
	}
	node := run{stdout: []byte("a;b\n")}
	mutated, binary := nativelyUncached(t, program)
	for _, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Fatal("whole member mutant escaped rejection")
		}
		if diff := disagreement(node, got); diff != "" {
			t.Fatalf("whole member mutant: %s; stdout %q stderr %q", diff, got.stdout, got.stderr)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("whole member omission mutant caught by pinned union rejection")
}
