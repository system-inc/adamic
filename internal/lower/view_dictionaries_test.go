package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestViewDictionaryDescriptors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"interface Target { readonly [key:string]: boolean | undefined; readonly strict?: boolean }",
		"interface Child {readonly children:Target} interface Target {readonly [key:string]:Child | undefined}",
	} {
		t.Run(source, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
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
			build := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) }
			id, err := viewDictionaryContractHook(l, node, target, build)
			if err != nil {
				t.Fatal(err)
			}
			descriptor := l.result.ViewContracts[id-1]
			if descriptor.Element == 0 || descriptor.Unsupported != "dictionary source dispatch" {
				t.Fatalf("missing obligation: %#v", descriptor)
			}
			if supportedSlotContract(l.result, id, map[ir.ViewContractID]bool{}) {
				t.Fatal("unwired dictionary became a certificate")
			}
			if len(descriptor.Fields) == 1 && descriptor.Fields[0].Name != "strict" {
				t.Fatal("named field disappeared")
			}
			// Adapter failure must not leave an id that later looks like a complete graph.
			l.result = &ir.Program{}
			fail := errors.New("element unavailable")
			if _, err := viewDictionaryContractHook(l, node, target, func(*checker.Type) (ir.ViewContractID, error) { return 0, fail }); err != fail {
				t.Fatal(err)
			}
			if len(l.result.ViewContracts) != 0 || len(l.result.ViewContractTypes) != 0 {
				t.Fatal("failed dictionary intern poisoned registry")
			}
		})
	}
}
