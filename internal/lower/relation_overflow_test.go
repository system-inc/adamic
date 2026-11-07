package lower

import (
	"context"
	"errors"
	"fmt"
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

// A relation query that spends its budget answers "not related", here wrongly, since U1 is assignable to
// U2, and the checker reports TS2859 once and holds it. tsc's own check never makes this query, so the
// program loads, and lowers while its checker holds no overflow (the control). Once a query of the
// lowering's checker overflows, lowering the same program is NotYet (#apyq5ve). Mutant: the guard in
// lowerCheckedWithCycleTypes removed, the second lowering succeeds.
func TestARelationTheCheckerGaveUpOnIsNotYet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	checked, err := load.Load([]string{writeAgreementSource(t, "main.ts", overflowingUnions())})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(checked.Files()) != 1 {
		t.Fatalf("Load gave %d files, want the entry alone", len(checked.Files()))
	}
	entry := checked.Files()[0]
	typeChecker, release := checked.Checker(ctx, entry)
	defer release()

	if overflow := checked.RelationOverflow(typeChecker); overflow != "" {
		t.Fatalf("before any query of its own, the checker holds %s", overflow)
	}
	if _, err := Lower(ctx, checked); err != nil {
		t.Fatalf("the control, its checker holding no overflow: %v", err)
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
	overflow := checked.RelationOverflow(typeChecker)
	if !strings.Contains(overflow, "TS2859") {
		t.Fatalf("after the overflowing query, RelationOverflow gave %q, want a TS2859", overflow)
	}

	_, err = Lower(ctx, checked)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.Error(), "TS2859") || !strings.Contains(notYet.Error(), "gave up on as too complex") {
		t.Fatalf("lowering with the checker holding %s: got %v, want a NotYet naming the TS2859", overflow, err)
	}
}
