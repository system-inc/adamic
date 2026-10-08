package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func arrayTupleJointFile(t *testing.T, declarations, name string) string {
	t.Helper()
	source, err := os.ReadFile("../../stage3/interface-downcasts/lane2/joints/tuple/" + name + ".a")
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

func TestCheckedViewArrayTupleJoints(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	for _, sample := range []struct{ name, output string }{
		{"json-unrelated", "string\n[2,3]\n"}, {"four-reads", "number\nstring\nstring\nstring\n"}, {"alias", "second\n"}, {"lazy-first", "string\n"}, {"lazy-second", "1\n"}, {"unread", "kept\n"}, {"wrong-first", "string\n"}, {"wrong-second", "number\n"}, {"wrong-arity", "string\n"}, {"length-alias", "3\n"}, {"source-write-good", "number\n"}, {"source-write-bad", "boolean\n"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			file := arrayTupleJointFile(t, declarations, sample.name)
			node := run{stdout: []byte(sample.output)}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(sample.name, "source-write-") {
				found := false
				for _, c := range program.ViewContracts {
					if c.FixedTuple && c.Name == "[fileId: IncrementalBuildInfoFileId, signature: [] | EmitSignature]" {
						found = true
						if len(c.Tuple) != 2 || len(c.Fields) != 2 || c.Fields[0].Name != "0" || c.Fields[1].Name != "1" {
							t.Fatal("original tuple descriptor reduced")
						}
					}
				}
				if !found {
					t.Fatal("original tuple descriptor missing")
				}
			}
			want := node
			message := ""
			tupleType := "[fileId: IncrementalBuildInfoFileId, signature: [] | EmitSignature]"
			switch sample.name {
			case "source-write-bad":
				message = "field read failed: <array write> matches no member of string | number; expected string | number, found boolean"
			case "wrong-first":
				message = "field read failed: node[0] is not a IncrementalBuildInfoFileId; expected IncrementalBuildInfoFileId, found string"
			case "wrong-second":
				message = "field read failed: node[1] matches no member of [] | EmitSignature; expected [] | EmitSignature, found number"
			case "wrong-arity":
				message = "field read failed: node[1] receiver is not a " + tupleType + "; expected " + tupleType + ", found array"
			case "length-alias":
				message = "field read failed: node.length is not a " + tupleType + "; expected " + tupleType + ", found array"
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
			switch sample.name {
			case "source-write-bad":
				boolID := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Boolean, Name: "boolean"})
				rewrite := func(n any) any {
					if a, ok := n.(ir.ArrayLiteral); ok && a.ElementContract > 0 {
						c := &program.ViewContracts[a.ElementContract-1]
						if c.Kind == ir.ViewUnion {
							c.Members = append(c.Members, boolID)
							changed++
						}
					}
					return n
				}
				program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
			case "wrong-first":
				stringID := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.String, Name: "string"})
				changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "node[0]" }, func(p ir.Property) ir.Property {
					p.Of = ir.String
					p.ViewContract = stringID
					p.ViewType = "string"
					return p
				})
			case "wrong-second":
				numberID := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
				changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "node[1]" }, func(p ir.Property) ir.Property {
					program.ViewContracts[p.ViewContract-1].Members = append(program.ViewContracts[p.ViewContract-1].Members, numberID)
					changed++
					return p
				})
			case "wrong-arity":
				for i := range program.ViewContracts {
					c := &program.ViewContracts[i]
					if c.FixedTuple && len(c.Tuple) == 2 {
						c.Tuple = append(c.Tuple, c.Tuple[1])
						changed++
					}
				}
			case "length-alias":
				rewrite := func(n any) any {
					if read, ok := n.(ir.Length); ok && read.Tuple {
						read.TupleArity = 3
						changed++
						return read
					}
					return n
				}
				program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
				for i := range program.Functions {
					program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
				}
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

func TestCheckedViewArrayTupleConsumerBoundary(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	file := arrayTupleJointFile(t, declarations, "four-reads")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	head := strings.Split(string(data), "function read")[0]
	source := head + "function read(node: Target): string { return `${JSON.stringify(node)}`; }\nconst base: (number | string)[] = [1, 'sig'];\nconsole.log(read(base as Target));\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("[1,\"sig\"]\n")}, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	_, err = lowered(t, file)
	if err == nil || !strings.Contains(err.Error(), "JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)") {
		t.Fatalf("unpinned consumer refusal: %v", err)
	}
}
func TestCheckedViewArrayTupleJointCounts(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	root, err := filepath.Abs(checkedViewFixturePath(repository))
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{}
	for _, name := range []string{"four-reads", "alias", "lazy-first", "lazy-second", "unread", "wrong-first", "wrong-second", "wrong-arity", "length-alias", "source-write-good", "source-write-bad", "json-unrelated"} {
		path := arrayTupleJointFile(t, declarations, name)
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, checkedViewFixturePath(relative), false, nil, false, false)
		label := "stage3/interface-downcasts/lane2/joints/tuple/" + name + ".a"
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
