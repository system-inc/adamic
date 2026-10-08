package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestSharedArrayContractAdapter(t *testing.T) {
	for _, probe := range []struct {
		source   string
		rejected bool
	}{
		{"type Target = readonly boolean[];", false},
		{"type Target = readonly (number | string)[];", false},
		{"interface Child {readonly children: readonly Child[]} type Target=readonly Child[];", false},
		{"type Target = unknown[];", true},
	} {
		t.Run(probe.source, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			node := file.Statements.Nodes[len(file.Statements.Nodes)-1]
			target := checked.GetTypeAtLocation(node.Name())
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			id, err := l.viewContract(node, target)
			if probe.rejected {
				if err != nil {
					t.Fatal(err)
				}
				if supportedSlotContract(l.result, id, map[ir.ViewContractID]bool{}) {
					t.Fatal("unknown element was certified")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[int(id)-1]
			if contract.Kind != ir.ViewArray || contract.Element == 0 {
				t.Fatalf("missing logical element contract: %#v", contract)
			}
			if probe.source == "type Target = readonly (number | string)[];" {
				element := l.result.ViewContracts[int(contract.Element)-1]
				if element.Kind != ir.ViewUnion || len(element.Members) != 2 {
					t.Fatalf("mixed storage is not a logical element certificate: %#v", element)
				}
			}
			if len(l.result.ViewContracts) > 4 {
				t.Fatal("recursive array graph expanded instead of linking")
			}
		})
	}
}
