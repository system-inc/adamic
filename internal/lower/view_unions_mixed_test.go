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

func mixedUnionLowering(t *testing.T, source string) (*lowering, *checker.Type, func()) {
	t.Helper()
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
	node := file.Statements.Nodes[len(file.Statements.Nodes)-1]
	return &lowering{program: program, checker: checked, result: &ir.Program{}}, checked.GetTypeAtLocation(node.Name()), release
}

func TestMixedUnionContractGraph(t *testing.T) {
	for _, source := range []string{
		`interface Identifier {readonly text:string} type Target=string|Identifier;`,
		`interface Node {readonly ready:boolean} type Target=number|Node;`,
		`interface Node {readonly ready:boolean} type Target=Node|readonly Node[];`,
		`interface Expression {readonly value:number} type Target=boolean|Expression;`,
		`interface Left {readonly left:string} interface Right {readonly right:number} type Target=Left|Right;`,
	} {
		t.Run(source, func(t *testing.T) {
			l, target, release := mixedUnionLowering(t, source)
			defer release()
			node := l.program.Files()[0].Statements.Nodes[0]
			id, err := internMixedUnionViewContract(l, node, target, func(member *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, member) })
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[int(id)-1]
			if len(contract.Members) != len(target.Types()) {
				t.Fatal("union lost a member contract")
			}
			for i, member := range target.Types() {
				child := l.result.ViewContracts[int(contract.Members[i])-1]
				if child.Name != l.checker.TypeToString(member) || child.Kind == ir.ViewUnknown {
					t.Fatalf("member was substituted: %#v", child)
				}
				if child.Kind == ir.ViewObject && len(child.Fields) == 0 {
					t.Fatal("object member lost transitive fields")
				}
				if child.Kind == ir.ViewArray && child.Element == 0 {
					t.Fatal("array member lost element contract")
				}
			}
		})
	}
}

func TestMixedUnionContractFailureDoesNotCertifyRetry(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `type Target=string|number;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	sentinel := errors.New("member unsupported")
	calls := 0
	build := func(member *checker.Type) (ir.ViewContractID, error) {
		calls++
		if calls == 2 {
			return 0, sentinel
		}
		return l.viewContract(node, member)
	}
	id, err := internMixedUnionViewContract(l, node, target, build)
	if id != 0 || !errors.Is(err, sentinel) {
		t.Fatalf("failure lost: id %d, %v", id, err)
	}
	if len(l.result.ViewContracts) != 0 || len(l.result.ViewContractTypes) != 0 {
		t.Fatal("failed graph remains memoized as certified")
	}
	id, err = internMixedUnionViewContract(l, node, target, func(member *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, member) })
	if err != nil || id == 0 || len(l.result.ViewContracts[int(id)-1].Members) != 2 {
		t.Fatalf("retry used incomplete graph: id %d, %v", id, err)
	}
}

func TestMixedUnionContractUnknownMemberFails(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `type Target=string|number;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	for _, kind := range []ir.ViewKind{ir.ViewUnknown, ir.ViewScalar} {
		id, err := internMixedUnionViewContract(l, node, target, func(member *checker.Type) (ir.ViewContractID, error) {
			if kind == ir.ViewScalar {
				return 0, nil
			}
			l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: kind})
			return ir.ViewContractID(len(l.result.ViewContracts)), nil
		})
		if err == nil || id != 0 || len(l.result.ViewContracts) != 0 {
			t.Fatal("unavailable member certified")
		}
	}
}

func TestMixedUnionContractRecursiveMember(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `interface Node {readonly next:Target} type Target=Node|number;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	id, err := internMixedUnionViewContract(l, node, target, func(member *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, member) })
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, member := range l.result.ViewContracts[int(id)-1].Members {
		child := l.result.ViewContracts[int(member)-1]
		if child.Kind == ir.ViewObject {
			if len(child.Fields) != 1 || child.Fields[0].Contract != id {
				t.Fatal("recursive member lost the shared union id")
			}
			found = true
		}
	}
	if !found || len(l.result.ViewContracts) != 3 {
		t.Fatal("recursive contract graph expanded instead of linking")
	}
}

func TestMixedUnionContractPhantomBrandRemainsUnsupported(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `type Brand=string&{readonly marker:void}; type Target=Brand|number;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	id, err := internMixedUnionViewContract(l, node, target, func(member *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, member) })
	if err == nil || id != 0 || len(l.result.ViewContracts) != 0 {
		t.Fatal("runtime kind was treated as a phantom brand certificate")
	}
}
