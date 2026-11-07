package lower

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// overflowingUnionsSources is how many members U1 has, and U2 one more, and overflowingUnionsDepth how
// many interfaces each member's chain holds.
const (
	overflowingUnionsSources = 1000
	overflowingUnionsDepth   = 4
)

// overflowingUnions is a program that lowers, with two unions nothing in it relates. Source member Si
// is the head of a chain of distinct interfaces Ai_1 to Ai_4, whose last link holds v, si and ti; target
// member Tj heads Bj_1 to Bj_4, whose last holds v and tj. So Si is assignable to Ti alone, U1 is
// assignable to U2, and every other pair recurses to the bottom of its chain and fails there.
//
// The checker tries U2's members in order for each Si until Ti answers, and U2 has one member more than
// U1, so its pairing of members by index never applies. Ti has a rank of its own in that order, so the
// failed pairs number at least 0 + 1 + ... + 999 = 499,500, each five distinct failed comparisons
// (Si with Tj, then each link), 2,497,500 in all: past a top-level query's budget, (16,000,000 less the
// relation's cache) / 8, which is at most 2,000,000 (relater.go:378).
func overflowingUnions() string {
	var builder strings.Builder
	chain := func(head, link string, member int, bottom string) {
		fmt.Fprintf(&builder, "interface %s%d {\n\treadonly x: %s%d_1;\n}\n", head, member, link, member)
		for level := 1; level < overflowingUnionsDepth; level++ {
			fmt.Fprintf(&builder, "interface %s%d_%d {\n\treadonly x: %s%d_%d;\n}\n", link, member, level, link, member, level+1)
		}
		fmt.Fprintf(&builder, "interface %s%d_%d {\n\treadonly v: number;\n%s}\n", link, member, overflowingUnionsDepth, bottom)
	}
	for member := range overflowingUnionsSources {
		chain("S", "A", member, fmt.Sprintf("\treadonly s%d: number;\n\treadonly t%d: number;\n", member, member))
	}
	for member := range overflowingUnionsSources + 1 {
		chain("T", "B", member, fmt.Sprintf("\treadonly t%d: number;\n", member))
	}
	union := func(name, head string, members int) {
		fmt.Fprintf(&builder, "type %s =", name)
		for member := range members {
			if member > 0 {
				builder.WriteString(" |")
			}
			fmt.Fprintf(&builder, " %s%d", head, member)
		}
		builder.WriteString(";\n")
	}
	union("U1", "S", overflowingUnionsSources)
	union("U2", "T", overflowingUnionsSources+1)
	builder.WriteString("console.log(\"ok\");\n")
	return builder.String()
}

// A query that spends the checker's budget answers a wrong false (U1 is assignable to U2), and the
// checker holds the overflow both ways RelationOverflow reads it: a TS2859 in its collection and a mark
// in its relation cache. Only the checker is asked here: lowering a program of ten thousand interfaces
// twice outran the test's bound on Cloud (job 01M6VVS1, 8 minutes), and the guard's lowering half is
// TestAnOverflowedRelationMakesLoweringNotYet's.
func TestACheckerHoldsTheRelationItGaveUpOn(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	checked, err := load.Load([]string{writeAgreementSource(t, "main.ts", overflowingUnions())})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := checked.Files()[0]
	typeChecker, release := checked.Checker(ctx, entry)
	defer release()
	if overflow := checked.RelationOverflow(typeChecker); overflow != "" || load.RelationOverflowMarked(typeChecker) {
		t.Fatalf("before any query of its own, the checker holds %q (or an overflow mark)", overflow)
	}
	unions := map[string]*checker.Type{}
	for _, statement := range entry.Statements.Nodes {
		if statement.Kind == ast.KindTypeAliasDeclaration {
			unions[statement.Name().Text()] = typeChecker.GetDeclaredTypeOfSymbol(typeChecker.GetSymbolAtLocation(statement.Name()))
		}
	}
	if unions["U1"] == nil || unions["U2"] == nil {
		t.Fatalf("found the unions %v, want U1 and U2", unions)
	}
	if typeChecker.IsTypeAssignableTo(unions["U1"], unions["U2"]) {
		t.Fatal("U1 is assignable to U2 within the budget; the pin needs a relation that spends it")
	}
	if !load.RelationOverflowMarked(typeChecker) {
		t.Fatal("after the overflowing query, no relation cache holds an overflow mark")
	}
	if overflow := checked.RelationOverflow(typeChecker); !strings.Contains(overflow, "TS2859") {
		t.Fatalf("after the overflowing query, RelationOverflow gave %q, want a TS2859", overflow)
	}
}

// A program that lowers is NotYet once its checker holds an overflowed relation. The mark is planted in
// the assignable relation's cache, as checkTypeRelatedToEx leaves one (relater.go:382), with no TS2859
// beside it: the road a diagnostic dropped at TypeToString's deepest nesting leaves, and the one the
// guard must still see.
func TestAnOverflowedRelationMakesLoweringNotYet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	checked, err := load.Load([]string{writeAgreementSource(t, "main.ts", "const total: number = 1 + 2;\nconsole.log(String(total));\n")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	typeChecker, release := checked.Checker(ctx, checked.Files()[0])
	defer release()
	if _, err := Lower(ctx, checked); err != nil {
		t.Fatalf("the control, its checker holding no overflow: %v", err)
	}
	relation := reflect.ValueOf(typeChecker).Elem().FieldByName("assignableRelation")
	if !relation.IsValid() || relation.IsNil() {
		t.Fatal("the checker's assignable relation isn't where the guard reads it")
	}
	field := relation.Elem().FieldByName("results")
	results := reflect.NewAt(field.Type(), field.Addr().UnsafePointer()).Elem()
	if results.IsNil() {
		results.Set(reflect.MakeMap(field.Type()))
	}
	mark := reflect.New(field.Type().Elem()).Elem()
	mark.SetUint(1<<1 | 1<<5) // RelationComparisonResultFailed | RelationComparisonResultComplexityOverflow
	results.SetMapIndex(reflect.New(field.Type().Key()).Elem(), mark)
	_, err = Lower(ctx, checked)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.Error(), "gave up on as too complex") || !strings.Contains(notYet.Error(), "marked as overflowed") {
		t.Fatalf("lowering with an overflow marked: got %v, want the guard's NotYet", err)
	}
}
