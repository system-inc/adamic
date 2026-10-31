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
// result. At run time the arguments choose the overload, not the result (Compiler's should on
// #8k12t4s), so the pairing is exact only while no registered member has two overloads whose
// parameters accept each other and whose results differ in the host object they hold with neither
// result fitting the other: then the view's arguments could pick the one its result left unpaired.
// This holds it, over every registered member with more than one call or construct signature.
//
// Two such overloads where the first's result fits the later one's, as realpathSync's and
// mkdtempSync's string and then string | Buffer do, aren't that road. The view's result then
// excludes only the later one's extra arm, and the first overload is the one the checker resolves
// those arguments to, so a direct call is typed the same way. They're logged, and their views are pinned where they stop
// (TestNodeHostViewStopsHeld and TestNodeHostViewRegisteredAgrees in internal/oracle).
func TestNodeRegisteredOverloadsSplitNoHostObject(t *testing.T) {
	t.Parallel()
	l, resolved := nodeRegisteredDeclarations(t, implementedNodeMembers)
	overloaded := 0
	var widened []string
	for _, name := range slices.Sorted(maps.Keys(resolved)) {
		if l.nodeOverloadCount(resolved[name]) > 1 {
			overloaded++
		}
		split, wider := l.nodeOverloadHostSplit(resolved[name])
		if split != "" {
			t.Errorf("%s has overloads whose parameters accept each other and whose results differ in holding a host object (%s): pair by parameters in nodeHostSignaturesReach", name, split)
		}
		if wider != "" {
			widened = append(widened, name+" ("+wider+")")
		}
	}
	t.Logf("%d registered names resolved, %d with more than one signature, read pair by pair; one result fitting the other: %s", len(resolved), overloaded, strings.Join(widened, ", "))
}

// The check above can find what it looks for, and only that: two synthetic overloads taking the same
// string, one giving performance's host object and then one a number, split; one giving a number
// and then one a number or the host object only widen.
func TestNodeRegisteredOverloadsFindASplit(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		`import { performance } from "node:perf_hooks";`,
		`declare function split(key: string): typeof performance;`,
		`declare function split(key: string): number;`,
		`declare function widen(key: string): number;`,
		`declare function widen(key: string): number | typeof performance;`,
		`console.log(String(performance.now() >= 0) + String(typeof split) + String(typeof widen));`,
	}, "\n")
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked}
	declared := map[string]*ast.Symbol{}
	for _, statement := range program.Files()[0].Statements.Nodes {
		if statement.Kind == ast.KindFunctionDeclaration && statement.Name() != nil {
			declared[statement.Name().Text()] = checked.GetSymbolAtLocation(statement.Name())
		}
	}
	if declared["split"] == nil || declared["widen"] == nil {
		t.Fatalf("found the declarations %v", slices.Collect(maps.Keys(declared)))
	}
	if split, _ := l.nodeOverloadHostSplit(declared["split"]); !strings.Contains(split, "Performance") {
		t.Fatalf("want split's two overloads found splitting on Performance, got %q", split)
	}
	if split, wider := l.nodeOverloadHostSplit(declared["widen"]); split != "" || !strings.Contains(wider, "Performance") {
		t.Fatalf("want widen's two overloads found only widening, got split %q, wider %q", split, wider)
	}
}

// nodeOverloadCount is how many call and construct signatures a member's type has.
func (l *lowering) nodeOverloadCount(member *ast.Symbol) int {
	proven := l.checker.GetTypeOfSymbol(member)
	return len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) + len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct))
}

// nodeOverloadHostSplit looks at each two overloads of a member, of one kind, whose parameters accept
// each other (the same count, each type assignable both ways) and whose results differ in the host
// object they may hold. It names the first such pair as wider when the earlier overload's result fits
// the later one's, and as split otherwise; each is "" when there's none.
func (l *lowering) nodeOverloadHostSplit(member *ast.Symbol) (split, wider string) {
	proven := l.checker.GetTypeOfSymbol(member)
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		signatures := l.checker.GetSignaturesOfType(proven, kind)
		for first := range signatures {
			for second := first + 1; second < len(signatures); second++ {
				if !l.nodeParametersAcceptEachOther(signatures[first], signatures[second]) {
					continue
				}
				one, other := l.checker.GetReturnTypeOfSignature(signatures[first]), l.checker.GetReturnTypeOfSignature(signatures[second])
				if l.nodeHeldHostObject(one, 0) == l.nodeHeldHostObject(other, 0) {
					continue
				}
				pair := l.checker.TypeToString(one) + " and " + l.checker.TypeToString(other)
				// The first overload is the one the checker resolves the arguments to, so only a first
				// whose result fits the later one's is the safe widening.
				if l.checker.IsTypeAssignableTo(one, other) {
					if wider == "" {
						wider = pair
					}
				} else if split == "" {
					split = pair
				}
			}
		}
	}
	return split, wider
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
