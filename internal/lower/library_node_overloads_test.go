package lower

import (
	"context"
	"maps"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// nodeHostSignaturesReach pairs a view's signature with each overload whose result fits the view's
// result. At run time the arguments choose the overload, not the result (Compiler's should on
// #8k12t4s), so the pairing is exact only while no registered member has two overloads that one
// argument list fits and whose results differ in the host object they hold with neither result
// fitting the other: then the view's arguments could pick the one its result left unpaired.
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
			t.Errorf("%s has overloads that one argument list fits and whose results differ in holding a host object (%s): pair by parameters in nodeHostSignaturesReach", name, split)
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
		`declare function arity(): typeof performance;`,
		`declare function arity(key?: string): number;`,
		`declare function overlap(key: string): typeof performance;`,
		`declare function overlap(key: string | number): number;`,
		`declare function apart(key: string): typeof performance;`,
		`declare function apart(key: number): number;`,
		`console.log(String(performance.now() >= 0) + typeof split + typeof widen + typeof arity + typeof overlap + typeof apart);`,
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
	for _, name := range []string{"split", "widen", "arity", "overlap", "apart"} {
		if declared[name] == nil {
			t.Fatalf("found the declarations %v, not %s", slices.Collect(maps.Keys(declared)), name)
		}
	}
	// arity's overloads both take no arguments, and overlap's both take a string, though neither
	// pair has the same parameters; apart's take nothing in common.
	for _, name := range []string{"split", "arity", "overlap"} {
		if split, _ := l.nodeOverloadHostSplit(declared[name]); !strings.Contains(split, "Performance") {
			t.Fatalf("want %s's two overloads found splitting on Performance, got %q", name, split)
		}
	}
	if split, wider := l.nodeOverloadHostSplit(declared["apart"]); split != "" || wider != "" {
		t.Fatalf("want apart's two overloads passed over, got split %q, wider %q", split, wider)
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

// nodeOverloadHostSplit looks at each two overloads of a member, of one kind, that one argument list
// fits (nodeParametersAcceptEachOther) and whose results differ in the host
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

// nodeParametersAcceptEachOther says whether some one argument list fits both signatures: an
// argument count both admit, and at each position below it, types with a value in common. It errs
// toward yes, so the check above fails closed. The fewest arguments both admit is the list with the
// fewest positions to agree on, so only those are compared; a rest parameter admits any count past
// its place, and GetTypeAtPosition reads its element there.
func (l *lowering) nodeParametersAcceptEachOther(one, other *checker.Signature) bool {
	most := func(signature *checker.Signature) int {
		if l.checker.HasEffectiveRestParameter(signature) {
			return math.MaxInt
		}
		return len(signature.Parameters())
	}
	fewest := max(one.MinArgumentCount(), other.MinArgumentCount())
	if fewest > min(most(one), most(other)) {
		return false
	}
	for position := range fewest {
		if !l.nodeTypesOverlap(l.checker.GetTypeAtPosition(one, position), l.checker.GetTypeAtPosition(other, position)) {
			return false
		}
	}
	return true
}

// nodeTypesOverlap says whether two parameter types may share a value: one assignable to the other,
// or, for a union, any member overlapping the other type. A union's members are read one level
// down, which is where Node's declarations put their alternatives.
func (l *lowering) nodeTypesOverlap(one, other *checker.Type) bool {
	if l.checker.IsTypeAssignableTo(one, other) || l.checker.IsTypeAssignableTo(other, one) {
		return true
	}
	for _, pair := range [][2]*checker.Type{{one, other}, {other, one}} {
		if pair[0].Flags()&checker.TypeFlagsUnion == 0 {
			continue
		}
		for _, member := range pair[0].Types() {
			if l.checker.IsTypeAssignableTo(member, pair[1]) || l.checker.IsTypeAssignableTo(pair[1], member) {
				return true
			}
		}
	}
	return false
}
