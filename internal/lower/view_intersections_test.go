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

func TestViewIntersectionContracts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source string
		members      int
		refused      bool
	}{
		{"interfaces", "interface A { readonly text: string } interface B { readonly count: number } type Target = A & B;", 2, false},
		{"nested", "interface A { readonly child: { readonly text: string } } type Target = A & { readonly child: { readonly count: number } };", 2, false},
		{"brand", "interface A { readonly text: string } type Target = A & { readonly __brand: void };", 1, false},
		{"runtime member", "interface A { readonly text: string } type Target = A & { readonly __brand: number };", 2, false},
		{"optional brand", "interface A { readonly text: string } type Target = A & { readonly __brand?: undefined };", 1, false},
		{"primitive brand lane", "type Target = string & { readonly __brand: void };", 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			node := file.Statements.Nodes[len(file.Statements.Nodes)-1]
			target := checked.GetTypeAtLocation(node.Name())
			calls := 0
			build := func(part *checker.Type) (ir.ViewContractID, error) {
				calls++
				l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewObject, Name: checked.TypeToString(part)})
				return ir.ViewContractID(len(l.result.ViewContracts)), nil
			}
			members, err := l.viewIntersectionContracts(node, target, build)
			if test.refused {
				if err == nil || calls != 0 {
					t.Fatalf("primitive family must remain separate: %v, calls %d", err, calls)
				}
				return
			}
			if err != nil || len(members) != test.members || calls != test.members {
				t.Fatalf("members %v, calls %d, error %v; want %d", members, calls, err, test.members)
			}
			sentinel := errors.New("member unavailable")
			members, err = l.viewIntersectionContracts(node, target, func(*checker.Type) (ir.ViewContractID, error) { return 0, sentinel })
			if !errors.Is(err, sentinel) || members != nil {
				t.Fatalf("lost constituent refusal: %v, %v", members, err)
			}
			members, err = l.viewIntersectionContracts(node, target, func(*checker.Type) (ir.ViewContractID, error) { return 0, nil })
			if err == nil || members != nil {
				t.Fatal("unknown member certified intersection")
			}
		})
	}
}
