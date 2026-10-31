package lower

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// nodeHostSignaturesReach pairs a view's signature with each overload whose result fits the view's
// result. At run time the arguments choose the overload, not the result, so that pairing is exact
// only while no registered member has two overloads whose parameters accept each other and whose
// results differ in holding a host object: the view's arguments could pick the one the result left
// unpaired (Compiler's should on #8k12t4s). This holds it, over every registered member with more
// than one call or construct signature. One that has such a pair fails, naming it.
func TestNodeRegisteredOverloadsSplitNoHostObject(t *testing.T) {
	t.Parallel()
	l, resolved := nodeRegisteredDeclarations(t, implementedNodeMembers)
	overloaded := 0
	for _, name := range slices.Sorted(maps.Keys(resolved)) {
		if l.nodeOverloadCount(resolved[name]) > 1 {
			overloaded++
		}
		if split := l.nodeOverloadHostSplit(resolved[name]); split != "" {
			t.Errorf("%s has overloads whose parameters accept each other and whose results differ in holding a host object (%s): pair by parameters in nodeHostSignaturesReach", name, split)
		}
	}
	t.Logf("%d registered names resolved, %d with more than one signature, read pair by pair", len(resolved), overloaded)
}

// The check above can find what it looks for: a synthetic pair of overloads taking the same string,
// one giving performance's host object and one a number.
func TestNodeRegisteredOverloadsFindASplit(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		`import { performance } from "node:perf_hooks";`,
		`declare function make(key: string): typeof performance;`,
		`declare function make(key: string): number;`,
		`console.log(String(performance.now() >= 0) + String(typeof make));`,
	}, "\n")
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked}
	var declared *ast.Symbol
	for _, statement := range program.Files()[0].Statements.Nodes {
		if statement.Kind == ast.KindFunctionDeclaration && statement.Name() != nil && statement.Name().Text() == "make" {
			declared = checked.GetSymbolAtLocation(statement.Name())
		}
	}
	if declared == nil {
		t.Fatal("found no make declaration")
	}
	if split := l.nodeOverloadHostSplit(declared); !strings.Contains(split, "Performance") {
		t.Fatalf("want make's two overloads found splitting on Performance, got %q", split)
	}
}

// nodeOverloadCount is how many call and construct signatures a member's type has.
func (l *lowering) nodeOverloadCount(member *ast.Symbol) int {
	proven := l.checker.GetTypeOfSymbol(member)
	return len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) + len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct))
}

// nodeOverloadHostSplit names the first two overloads of a member, of one kind, whose parameters
// accept each other (the same count, each type assignable both ways) and whose results differ in
// the host object they may hold, or "".
func (l *lowering) nodeOverloadHostSplit(member *ast.Symbol) string {
	proven := l.checker.GetTypeOfSymbol(member)
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		signatures := l.checker.GetSignaturesOfType(proven, kind)
		for first := range signatures {
			for second := first + 1; second < len(signatures); second++ {
				if !l.nodeParametersAcceptEachOther(signatures[first], signatures[second]) {
					continue
				}
				one := l.nodeHeldHostObject(l.checker.GetReturnTypeOfSignature(signatures[first]), 0)
				other := l.nodeHeldHostObject(l.checker.GetReturnTypeOfSignature(signatures[second]), 0)
				if one != other {
					return l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(signatures[first])) + " and " + l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(signatures[second]))
				}
			}
		}
	}
	return ""
}

// nodeParametersAcceptEachOther says whether two signatures take the same arguments: as many
// parameters, each one's type assignable to the other's.
func (l *lowering) nodeParametersAcceptEachOther(one, other *checker.Signature) bool {
	first, second := one.Parameters(), other.Parameters()
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		a, b := l.checker.GetTypeOfSymbol(first[index]), l.checker.GetTypeOfSymbol(second[index])
		if !l.checker.IsTypeAssignableTo(a, b) || !l.checker.IsTypeAssignableTo(b, a) {
			return false
		}
	}
	return true
}
