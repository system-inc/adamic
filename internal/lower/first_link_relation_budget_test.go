package lower

import (
	"context"
	goast "go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// firstLinkBudgetSpentPlanted is the RelationOverflow text the pins plant through firstLinkOverflowCheck.
const firstLinkBudgetSpentPlanted = "main.a:1:1: a relation the pin marked as overflowed"

// firstLinkBudgetLowering is a first-link lowering of firstLinkTwoFunctions, past its baseline, whose checker
// holds no overflow, so a pin drives its passes through firstLinkPass with the budget planted where it says.
func firstLinkBudgetLowering(t *testing.T) (*lowering, string) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", firstLinkTwoFunctions)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	built, err := lowerWithElementKey(&viewElementKey{}, loaded, check, entry, nil, false, false, ModuleFormatEcmaScript, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !built.firstLinkBudgetRead || built.firstLinkOverflow != "" || len(built.result.TrapJudgmentsAbandoned) != 0 {
		t.Fatalf("before the pass: baseline read %t, overflow %q, TrapJudgmentsAbandoned %+v, want read, none and none", built.firstLinkBudgetRead, built.firstLinkOverflow, built.result.TrapJudgmentsAbandoned)
	}
	return built, built.program.Where(entry.Statements.Nodes[0].Body().AsBlock().Statements.Nodes[0])
}

// firstLinkBudgetRow is the Error row a pass abandoned past a spent budget has, overflow being RelationOverflow's
// text.
func firstLinkBudgetRow(t *testing.T, row ir.Trap, pass, overflow string) {
	t.Helper()
	if row.Kind != "Error" || row.Phase != pass || row.Reason != "the checker's relation budget is spent: "+overflow || row.Unit != "" || row.Site != "" {
		t.Fatalf("%s's row: got %+v, want the Error row naming the spent budget, with no unit", pass, row)
	}
}

// A judging pass that stops once in a first-link lowering is run again, but never past a spent relation budget:
// on tsc's first link the cache passed its budget during findCycles' second run, and the third never returned,
// every relation query overflowing and its TS2859 deduped against all before it (J-5cj56ea). The budget is
// planted between the first run and the rerun, once the stop's row is recorded, and a rerun made past it hangs,
// as tsc's did. The pass gets one Error row naming the budget, with no rerun, and is named in
// TrapJudgmentsAbandoned. A pass after it abandons with the same words and never runs, and the budget, seen
// spent, isn't read again. The pass runs on its own goroutine under a deadline, so a rerun into the planted
// budget reads as a red here, never as the package's timeout.
// Mutant (first_link_relation_budget.mutants.md, 1): firstLinkPass reads the budget after the run again rather
// than before it, and the rerun hangs past the deadline.
func TestFirstLinkPassNeverRunsPastASpentRelationBudget(t *testing.T) {
	t.Parallel()
	built, inFirst := firstLinkBudgetLowering(t)
	rows := len(built.traps)
	spent := func() bool { return len(built.traps) > rows }
	reads := 0
	built.firstLinkOverflowCheck = func() string {
		reads++
		if spent() {
			return firstLinkBudgetSpentPlanted
		}
		return ""
	}
	hang := make(chan struct{})
	t.Cleanup(func() { close(hang) })
	reruns := 0
	done := make(chan error, 1)
	go func() {
		done <- built.firstLinkPass("checkSomething", false, &NotYet{Where: inFirst, What: "a stop in first"}, func() error {
			reruns++
			if spent() {
				<-hang
			}
			return nil
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("the pass: got %v, want its stop and the spent budget recorded and nil", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the pass ran again past the relation budget planted after its first stop, and that run hung: firstLinkPass must read the budget before each run again")
	}
	if reruns != 0 {
		t.Fatalf("the pass ran again %d times past the spent budget, want none", reruns)
	}
	recorded := built.traps[rows:]
	if len(recorded) != 2 || recorded[0].Kind != "NotYet" || recorded[0].Unit != "main.a first" {
		t.Fatalf("the pass's rows: got %+v, want first's NotYet, then the spent budget's Error", recorded)
	}
	firstLinkBudgetRow(t, recorded[1], "checkSomething", firstLinkBudgetSpentPlanted)
	want := []ir.TrapJudgmentAbandoned{{Pass: "checkSomething", Why: "Error"}}
	if got := built.result.TrapJudgmentsAbandoned; !slices.Equal(got, want) {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want %+v", got, want)
	}
	ran := false
	if err := built.firstLinkPass("checkAfter", false, nil, func() error { ran = true; return nil }); err != nil {
		t.Fatalf("the pass after: got %v, want it abandoned and nil", err)
	}
	if ran || len(built.traps) != rows+3 {
		t.Fatalf("the pass after: ran %t, rows %+v, want it never run and one row more", ran, built.traps[rows:])
	}
	firstLinkBudgetRow(t, built.traps[rows+2], "checkAfter", firstLinkBudgetSpentPlanted)
	want = append(want, ir.TrapJudgmentAbandoned{Pass: "checkAfter", Why: "Error"})
	if got := built.result.TrapJudgmentsAbandoned; !slices.Equal(got, want) {
		t.Fatalf("TrapJudgmentsAbandoned after the pass after: got %+v, want %+v", got, want)
	}
	if reads != 2 {
		t.Fatalf("the budget was read %d times, want twice: on the pass's entry and before its rerun, and never once seen spent", reads)
	}
}

// A pass that never stops can still spend the budget, and it's named on that pass, never on the next one. The
// budget is planted before a pass that doesn't stop (its err nil, as lower.go hands every pass to firstLinkPass
// in a first-link lowering), and that pass gets the Error row and is abandoned first. The pass after it stops at
// first, and abandons on entry with the same words, its stop never trapped, since its run was past the budget.
func TestFirstLinkSpentBudgetIsNamedOnThePassThatSpentIt(t *testing.T) {
	t.Parallel()
	built, inFirst := firstLinkBudgetLowering(t)
	rows := len(built.traps)
	built.firstLinkOverflowCheck = func() string { return firstLinkBudgetSpentPlanted }
	runs := 0
	run := func() error { runs++; return nil }
	if err := built.firstLinkPass("checkQuiet", false, nil, run); err != nil {
		t.Fatalf("the quiet pass: got %v, want it abandoned and nil", err)
	}
	if err := built.firstLinkPass("checkNext", false, &NotYet{Where: inFirst, What: "a stop in first"}, run); err != nil {
		t.Fatalf("the pass after: got %v, want it abandoned and nil", err)
	}
	if runs != 0 {
		t.Fatalf("a pass ran again %d times past the spent budget, want none", runs)
	}
	recorded := built.traps[rows:]
	if len(recorded) != 2 {
		t.Fatalf("the rows: got %+v, want the quiet pass's Error, then the next's", recorded)
	}
	firstLinkBudgetRow(t, recorded[0], "checkQuiet", firstLinkBudgetSpentPlanted)
	firstLinkBudgetRow(t, recorded[1], "checkNext", firstLinkBudgetSpentPlanted)
	want := []ir.TrapJudgmentAbandoned{{Pass: "checkQuiet", Why: "Error"}, {Pass: "checkNext", Why: "Error"}}
	if got := built.result.TrapJudgmentsAbandoned; !slices.Equal(got, want) {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want %+v", got, want)
	}
	if _, trapped := built.firstLinkStubs[built.firstLinkPassUnit(inFirst)]; trapped {
		t.Fatal("first was trapped by a stop met past the spent budget")
	}
}

// A budget spent before the first judging pass runs is the program's: a real first-link lowering of a checker
// holding an overflow mark records it once, as the RelationOverflow row with no unit, read at the baseline, and
// every whole-program pass lower.go hands firstLinkPass, read from its AST in order, then abandons with the
// budget's Error row: a judging pass in TrapJudgmentsAbandoned, a rewriting one in TrapRewritesSkipped. No second
// RelationOverflow row follows, and every row is Suspect. The mark is planted as
// TestAnOverflowedRelationMakesLoweringNotYet plants it.
// Mutant (first_link_relation_budget.mutants.md, 2): checkSlotWrites' site in lower.go hands firstLinkPass only a
// stop, and its row and its abandonment are missing.
func TestFirstLinkBudgetSpentAtTheBaselineAbandonsEveryPass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	checked, err := load.Load([]string{writeAgreementSource(t, "main.a", "const total: number = 1 + 2;\nconsole.log(String(total));\n")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	typeChecker, release := checked.Checker(ctx, checked.Files()[0])
	defer release()
	markRelationOverflow(t, typeChecker)
	overflow := checked.RelationOverflow(typeChecker)
	if overflow == "" {
		t.Fatal("the planted mark isn't read as an overflow")
	}
	program, err := LowerWithOptions(ctx, checked, Options{FirstLinkTraps: true})
	if err != nil {
		t.Fatalf("a first-link lowering with an overflow marked: %v, want its rows", err)
	}
	passes, rewrites := firstLinkLowerPasses(t)
	rows := program.Traps
	if len(rows) != 1+len(passes) {
		t.Fatalf("Traps: got %d rows %+v, want the RelationOverflow row and one for each of %v", len(rows), rows, passes)
	}
	if rows[0].Kind != "NotYet" || rows[0].Phase != "RelationOverflow" || rows[0].Unit != "" {
		t.Fatalf("the first row: got %+v, want the program's RelationOverflow row", rows[0])
	}
	var abandoned []ir.TrapJudgmentAbandoned
	var skipped []string
	for index, pass := range passes {
		firstLinkBudgetRow(t, rows[1+index], pass, overflow)
		if rewrites[pass] {
			skipped = append(skipped, pass)
		} else {
			abandoned = append(abandoned, ir.TrapJudgmentAbandoned{Pass: pass, Why: "Error"})
		}
	}
	if !slices.Equal(program.TrapJudgmentsAbandoned, abandoned) {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want %+v", program.TrapJudgmentsAbandoned, abandoned)
	}
	if !slices.Equal(program.TrapRewritesSkipped, skipped) {
		t.Fatalf("TrapRewritesSkipped: got %q, want %q", program.TrapRewritesSkipped, skipped)
	}
	for _, row := range rows {
		if !row.Suspect {
			t.Errorf("a row that isn't Suspect past the spent budget: %+v", row)
		}
	}
}

// firstLinkLowerPasses is every pass lowerWithElementKey hands firstLinkPass, by its name, in source order, and
// which of them rewrite.
func firstLinkLowerPasses(t *testing.T) ([]string, map[string]bool) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "lower.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var passes []string
	rewrites := map[string]bool{}
	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*goast.FuncDecl)
		if !isFunction || function.Name.Name != "lowerWithElementKey" {
			continue
		}
		goast.Inspect(function.Body, func(node goast.Node) bool {
			call, isCall := node.(*goast.CallExpr)
			if !isCall || len(call.Args) != 4 {
				return true
			}
			if selector, isSelector := call.Fun.(*goast.SelectorExpr); !isSelector || selector.Sel.Name != "firstLinkPass" {
				return true
			}
			literal, isLiteral := call.Args[0].(*goast.BasicLit)
			flag, isIdent := call.Args[1].(*goast.Ident)
			if !isLiteral || !isIdent {
				t.Fatalf("a firstLinkPass call whose name or rewrites isn't written out: %#v", call)
			}
			name, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			passes = append(passes, name)
			rewrites[name] = flag.Name == "true"
			return true
		})
	}
	if len(passes) == 0 {
		t.Fatal("no firstLinkPass call read from lowerWithElementKey")
	}
	return passes, rewrites
}
