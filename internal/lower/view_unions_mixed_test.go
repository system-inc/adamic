package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func mixedUnionLowering(test *testing.T, source string) (*lowering, *checker.Type, func()) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "contract.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	node := file.Statements.Nodes[len(file.Statements.Nodes)-1]
	return &lowering{
		program: program,
		checker: checked,
		result:  &ir.Program{},
	}, checked.GetTypeAtLocation(node.Name()), release
}

func TestMixedUnionContractGraph(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`interface Identifier {readonly text:string} type Target=string|Identifier;`,
		`interface Node {readonly ready:boolean} type Target=number|Node;`,
		`interface Node {readonly ready:boolean} type Target=Node|readonly Node[];`,
		`interface Expression {readonly value:number} type Target=boolean|Expression;`,
		`interface Left {readonly left:string} interface Right {readonly right:number} type Target=Left|Right;`,
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			if strings.Contains(source, "readonly Node[]") {
				// Array membership needs V3 element-kind metadata and hole-aware reads.
				test.Skip("awaits compiler/views-v3: array element kind and holes (b065fa576)")
			}
			lowering, target, release := mixedUnionLowering(test, source)
			defer release()
			node := lowering.program.Files()[0].Statements.Nodes[0]
			id, internMixedUnionViewContractError := internMixedUnionViewContract(
				lowering,
				node,
				target,
				func(member *checker.Type) (ir.ViewContractID, error) { return lowering.viewContract(node, member) },
			)
			if internMixedUnionViewContractError != nil {
				test.Fatal(internMixedUnionViewContractError)
			}
			contract := lowering.result.ViewContracts[int(id)-1]
			if len(contract.Members) != len(target.Types()) {
				test.Fatal("union lost a member contract")
			}
			for index, member := range target.Types() {
				child := lowering.result.ViewContracts[int(contract.Members[index])-1]
				if child.Name != lowering.checker.TypeToString(member) || child.Kind == ir.ViewUnknown {
					test.Fatalf("member was substituted: %#v", child)
				}
				if child.Kind == ir.ViewObject && len(child.Fields) == 0 {
					test.Fatal("object member lost transitive fields")
				}
				if child.Kind == ir.ViewArray && child.Element == 0 {
					test.Fatal("array member lost element contract")
				}
			}
		})
	}
}

func TestMixedUnionContractFailureDoesNotCertifyRetry(test *testing.T) {
	test.Parallel()
	lowering, target, release := mixedUnionLowering(test, `type Target=string|number;`)
	defer release()
	node := lowering.program.Files()[0].Statements.Nodes[0]
	sentinel := errors.New("member unsupported")
	calls := 0
	build := func(member *checker.Type) (ir.ViewContractID, error) {
		calls++
		if calls == 2 {
			return 0, sentinel
		}
		return lowering.viewContract(node, member)
	}
	id, internMixedUnionViewContractError := internMixedUnionViewContract(lowering, node, target, build)
	if id != 0 || !errors.Is(internMixedUnionViewContractError, sentinel) {
		test.Fatalf("failure lost: id %d, %v", id, internMixedUnionViewContractError)
	}
	if len(lowering.result.ViewContracts) != 0 || len(lowering.result.ViewContractTypes) != 0 {
		test.Fatal("failed graph remains memoized as certified")
	}
	id, internMixedUnionViewContractError = internMixedUnionViewContract(
		lowering,
		node,
		target,
		func(member *checker.Type) (ir.ViewContractID, error) { return lowering.viewContract(node, member) },
	)
	if internMixedUnionViewContractError != nil ||
		id == 0 ||
		len(lowering.result.ViewContracts[int(id)-1].Members) != 2 {
		test.Fatalf("retry used incomplete graph: id %d, %v", id, internMixedUnionViewContractError)
	}
}

func TestMixedUnionContractUnknownMemberFails(test *testing.T) {
	test.Parallel()
	lowering, target, release := mixedUnionLowering(test, `type Target=string|number;`)
	defer release()
	node := lowering.program.Files()[0].Statements.Nodes[0]
	for _, kind := range []ir.ViewKind{ir.ViewUnknown, ir.ViewScalar} {
		id, internMixedUnionViewContractError := internMixedUnionViewContract(lowering, node, target, func(
			member *checker.Type,
		) (ir.ViewContractID, error) {
			if kind == ir.ViewScalar {
				return 0, nil
			}
			lowering.result.ViewContracts = append(lowering.result.ViewContracts, ir.ViewContract{Kind: kind})
			return ir.ViewContractID(len(lowering.result.ViewContracts)), nil
		})
		if internMixedUnionViewContractError == nil || id != 0 || len(lowering.result.ViewContracts) != 0 {
			test.Fatal("unavailable member certified")
		}
	}
}

func TestMixedUnionContractRecursiveMember(test *testing.T) {
	test.Parallel()
	lowering, target, release := mixedUnionLowering(
		test,
		`interface Node {readonly next:Target} type Target=Node|number;`,
	)
	defer release()
	node := lowering.program.Files()[0].Statements.Nodes[0]
	id, internMixedUnionViewContractError := internMixedUnionViewContract(
		lowering,
		node,
		target,
		func(member *checker.Type) (ir.ViewContractID, error) { return lowering.viewContract(node, member) },
	)
	if internMixedUnionViewContractError != nil {
		test.Fatal(internMixedUnionViewContractError)
	}
	found := false
	for _, member := range lowering.result.ViewContracts[int(id)-1].Members {
		child := lowering.result.ViewContracts[int(member)-1]
		if child.Kind == ir.ViewObject {
			if len(child.Fields) != 1 || child.Fields[0].Contract != id {
				test.Fatal("recursive member lost the shared union id")
			}
			found = true
		}
	}
	if !found || len(lowering.result.ViewContracts) != 3 {
		test.Fatal("recursive contract graph expanded instead of linking")
	}
}

func TestMixedUnionContractPhantomBrandUsesPrimitiveBase(test *testing.T) {
	test.Parallel()
	lowering, target, release := mixedUnionLowering(
		test,
		`type Brand=string&{readonly marker:void}; type Target=Brand|number;`,
	)
	defer release()
	node := lowering.program.Files()[0].Statements.Nodes[0]
	id, internMixedUnionViewContractError := internMixedUnionViewContract(
		lowering,
		node,
		target,
		func(member *checker.Type) (ir.ViewContractID, error) { return lowering.viewContract(node, member) },
	)
	if internMixedUnionViewContractError != nil || id == 0 {
		test.Fatalf("approved phantom base refused: %v", internMixedUnionViewContractError)
	}
	found := false
	for _, child := range lowering.result.ViewContracts[id-1].Members {
		contract := lowering.result.ViewContracts[child-1]
		if contract.Of == ir.String {
			found = true
			if contract.Kind != ir.ViewScalar || len(contract.Fields) != 0 {
				test.Fatalf("brand became runtime object: %#v", contract)
			}
		}
	}
	if !found {
		test.Fatal("string member lost")
	}
}

func TestMixedUnionContractPhantomVoidIsUndefined(test *testing.T) {
	test.Parallel()
	lowering, target, release := mixedUnionLowering(test, `type Target=void&{readonly marker:void};`)
	defer release()
	node := lowering.program.Files()[0].Statements.Nodes[0]
	id, viewContractError := lowering.viewContract(node, target)
	if viewContractError != nil || id == 0 {
		test.Fatalf("phantom void refused: %v", viewContractError)
	}
	contract := lowering.result.ViewContracts[id-1]
	if contract.Kind != ir.ViewUndefined ||
		!contract.Undefined ||
		contract.Unsupported != "" ||
		len(contract.Fields) != 0 {
		test.Fatalf("phantom void became an object obligation: %#v", contract)
	}
}
