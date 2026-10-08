package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareViewCallableRead(t *testing.T) {
	for _, declaration := range []string{
		"type Target = (value: number) => number;",
		"type Target = ((value: number) => number) | undefined;",
		"type Target = () => string;",
		"interface Result { unread: unknown; }; type Target = (value: number) => Result;",
		"type Target = <T>(value: T) => T;",
		"type Target = (...values: number[]) => number;",
		"type Target = { (value: number): number; (value: string): string; } | ((value: number) => number);",
	} {
		t.Run(declaration, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "read.a")
			if err := os.WriteFile(path, []byte(declaration), 0600); err != nil {
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
			// A normal cast has already interned the registry before a union read.
			if declaration == "type Target = { (value: number): number; (value: string): string; } | ((value: number) => number);" {
				l.result.ViewContractTypes = map[int]ir.ViewContractID{}
			}
			id, err := l.prepareViewCallableRead(node, target)
			if declaration == "type Target = <T>(value: T) => T;" || declaration == "type Target = (...values: number[]) => number;" || declaration == "type Target = { (value: number): number; (value: string): string; } | ((value: number) => number);" {
				if err == nil {
					t.Fatal("unsupported read admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[id-1]
			if contract.Kind != ir.ViewCallable || contract.Result == 0 {
				t.Fatalf("missing reached signature %#v", contract)
			}
			if len(contract.Parameters) != 0 && l.result.ViewContracts[contract.Parameters[0]-1].Of != ir.Number {
				t.Fatal("wrong parameter representation")
			}
			for _, child := range l.result.ViewContracts {
				if len(child.Fields) != 0 {
					t.Fatal("signature preparation visited descendant members")
				}
			}
		})
	}
}
