package oracle

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewRecursiveMutableArrays(t *testing.T) {
	for _, site := range []string{"control", "producer", "read", "write", "cycle"} {
		t.Run(site, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-nominal-recursive-array-mutable-"+site)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			got, binary := nativelyUncached(t, program)
			for _, result := range []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(truth, result); diff != "" {
					t.Fatal(diff)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node=%q; both native modes and JavaScript agree, leak check clean", truth.stdout)
		})
	}
}

func TestCheckedViewRecursiveMutableArrayMutants(t *testing.T) {
	for _, site := range []string{"producer", "read", "push", "index"} {
		t.Run(site, func(t *testing.T) {
			fixture := site
			if site == "push" || site == "index" {
				fixture = "write"
			}
			program, path := interfaceFixture(t, "nullish/maps/entry-nominal-recursive-array-mutable-"+fixture)
			truth := onNode(t, path)
			fake, leaf := -1, -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
				if local.Name == "leaf" {
					leaf = i
				}
			}
			if fake < 0 || leaf < 0 {
				t.Fatal("missing mutation locals")
			}
			forged := ir.ObjectLiteral{Fields: []ir.Field{
				{Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}},
				{Name: "children", Value: ir.ArrayLiteral{Element: ir.Object}},
			}}
			changed := false
			for i, statement := range program.Main {
				if site == "push" {
					if evaluation, ok := statement.(ir.Evaluate); ok {
						if push, ok := evaluation.Value.(ir.ArrayPush); ok {
							push.Value = forged
							evaluation.Value = push
							program.Main[i] = evaluation
							program.Main = program.Main[:i+1]
							changed = true
							break
						}
					}
				} else if site == "index" {
					if write, ok := statement.(ir.SetIndex); ok {
						write.Value = forged
						program.Main[i] = write
						program.Main = program.Main[:i+1]
						changed = true
						break
					}
				} else if declaration, ok := statement.(ir.Declare); ok {
					name := program.Locals[declaration.Local].Name
					if site == "producer" && name == "item" || site == "read" && name == "source" {
						mutation := ir.SetProperty{Object: ir.Read{Local: leaf, Of: ir.Object}, Name: "child", Value: ir.Read{Local: fake, Of: ir.Object}}
						rest := append([]ir.Statement(nil), program.Main[i+1:]...)
						program.Main = append(program.Main[:i+1], mutation)
						program.Main = append(program.Main, rest...)
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			got, _ := nativelyUncached(t, program)
			for backend, result := range []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				expected := "Map nominal producer failed:"
				if site == "read" {
					expected = "field read failed:"
				}
				if result.exitCode != 70 || !strings.Contains(string(result.stderr), expected) || !strings.Contains(string(result.stderr), "class identity") || ((site == "push" || site == "index") && !strings.Contains(string(result.stderr), "<array write>")) {
					t.Fatalf("%s mutation escaped at backend %d: %#v", site, backend, result)
				}
				t.Logf("backend=%d exit=%d stderr=%q", backend, result.exitCode, result.stderr)
			}
			t.Logf("unmodified source on Node=%q; real lookalike payload rejected at %s", truth.stdout, site)
		})
	}
}

func recursiveMutableArrayCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"control", "producer", "read", "write", "cycle"} {
		rows = append(rows, counted(t, "stage3/interface-downcasts/nullish/maps/entry-nominal-recursive-array-mutable-"+name+".a", false, nil, false, false))
	}
	return rows
}

func TestCheckedViewRecursiveMutableArrayCounts(t *testing.T) {
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range recursiveMutableArrayCounts(t) {
		if !strings.Contains(string(recorded), row+"\n") {
			t.Errorf("unrecorded fixture counts: %s", row)
		}
		t.Log(row)
	}
}
