package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// stopWhereStatics is a program with a static constructor object, Counter's, whose spread of a parameter
// checkAccessorSpreads stops at, at main.ts:10:14, inside copy: the shape that named only the file on tsc
// (#nxb3xh1). body is the spread's function and the rest of the program.
func stopWhereStatics(body string) string {
	return "class Counter {\n    static start(): number {\n        return 1;\n    }\n}\ninterface Point {\n    readonly x: number;\n}\n" + body
}

// A judging pass's stop names the spread it's about (stopWhere), so a first-link lowering traps the unit holding
// it and judges on, where checkAccessorSpreads' stop had named only the file, a Gap, and the pass's judgment was
// abandoned (#nxb3xh1). The spread lies in copy, whose body becomes the trap: it's the one row, no judgment is
// abandoned, and the program runs up to copy's call and exits 75 there, on native and on the JavaScript backend.
// The same spread in a top-level statement traps that statement. Without the option the stop is at the spread.
// Mutant (stop_where.mutants.md): stopWhere returns l.result.Source, and each is a Gap again, its judgment
// abandoned, and the stop names the file alone.
func TestFirstLinkAccessorSpreadStopTrapsItsUnit(t *testing.T) {
	t.Parallel()
	const stop = "spreading in a program with static constructor objects"
	source := stopWhereStatics("function copy(point: Point): Point {\n    return { ...point };\n}\nconsole.log(String(copy({ x: 1 }).x + Counter.start()));\n")
	program, err := lowerFirstLinkTS(t, source, true)
	firstLinkOneTrap(t, program, err, ir.Trap{Kind: "NotYet", Reason: stop, Unit: "main.ts copy", Phase: "checkAccessorSpreads"})
	if rows := firstLinkLocalSites(program); len(rows) != 1 || rows[0].Site != "main.ts:10:14" {
		t.Fatalf("Traps: got %+v, want the one row at the spread, main.ts:10:14", rows)
	}
	firstLinkTrapped(t, program, "", "NotYet: "+stop, "main.ts:10:14")
	_, err = lowerFirstLinkTS(t, source, false)
	var notYet *NotYet
	if !errors.As(err, &notYet) || firstLinkFromFile(err, "main.ts") != "main.ts:10:14: stage 0 can't lower "+stop+" yet" {
		t.Errorf("without the option: got %v, want the stop at the spread, main.ts:10:14", err)
	}
	statement := stopWhereStatics("const point: Point = { x: 1 };\nconst copied: Point = { ...point };\nconsole.log(String(copied.x + Counter.start()));\n")
	program, err = lowerFirstLinkTS(t, statement, true)
	firstLinkOneTrap(t, program, err, ir.Trap{Kind: "NotYet", Reason: stop, Unit: "main.ts module", Phase: "checkAccessorSpreads"})
	if rows := firstLinkLocalSites(program); len(rows) != 1 || !strings.HasSuffix(rows[0].Site, "main.ts:10:25") {
		t.Errorf("Traps of the top-level spread: got %+v, want the one row at the spread, main.ts:10:25", rows)
	}
}

// checkRangeUnionReads' stop names the function or the top-level statement holding the read (stopWhere), since
// the IR keeps no node for it, so a first-link lowering traps that unit and judges on, where the stop had named
// only the file, a Gap, and the pass's judgment was abandoned (#nxb3xh1). No source program reaches the stop: each
// road it guards stops in lowering first (Object.keys, Object.entries, Object.hasOwn, Object.assign and a for-in
// over a range type, in a function and at the top level, each tried on J-d6gkyvx). So this drives a real
// first-link lowering's pass, as TestFirstLinkJudgingPassGivingUpIsRecorded does, with the narrow of an untyped
// union local the pass stops (TestRangeUnionNarrowStops) put in width's body, which traps width at its
// declaration, then in the top-level statement's IR, which traps the statement. Neither abandons the judgment.
// Mutant (stop_where.mutants.md): stopWhere returns l.result.Source, and the first stop is a Gap, its judgment
// abandoned.
func TestFirstLinkRangeUnionStopTrapsItsUnit(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", rangeTextRange+rangeNodeArray+"export function width(range: TextRange): number { return range.end - range.pos; }\nconsole.log('ok');\n")})
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
	if len(built.traps) != 0 || !built.rangeUnionsHeld {
		t.Fatalf("before the narrows: Traps %+v and a range type held as a union %t, want none and held", built.traps, built.rangeUnionsHeld)
	}
	narrow := func(function int) ir.Statement {
		held := len(built.result.Locals)
		built.result.Locals = append(built.result.Locals, ir.Local{Name: "held", Type: ir.Union, Function: function})
		return ir.Evaluate{Value: ir.Narrow{Value: ir.Read{Local: held, Of: ir.Union}, To: ir.Object}}
	}
	declaration, statement := entry.Statements.Nodes[3], entry.Statements.Nodes[4]
	judged := func(name string, want ir.Trap) {
		t.Helper()
		rows := len(built.traps)
		if err := built.firstLinkPass("checkRangeUnionReads", false, built.checkRangeUnionReads(), built.checkRangeUnionReads); err != nil {
			t.Fatalf("%s: got %v, want its stop recorded and nil", name, err)
		}
		recorded := built.traps[rows:]
		for index := range recorded {
			recorded[index].Site = firstLinkLocalSite.ReplaceAllString(recorded[index].Site, "$1")
		}
		if len(recorded) != 1 || recorded[0].Kind != want.Kind || !strings.HasPrefix(recorded[0].Reason, want.Reason) || recorded[0].Site != want.Site || recorded[0].Unit != want.Unit || recorded[0].Phase != want.Phase || recorded[0].Suspect {
			t.Fatalf("%s: rows %+v, want the one %s at %s in %s", name, recorded, want.Reason, want.Site, want.Unit)
		}
		if got := built.result.TrapJudgmentsAbandoned; len(got) != 0 {
			t.Fatalf("%s: TrapJudgmentsAbandoned %+v, want none", name, got)
		}
	}
	width := -1
	for index, node := range built.functionNodes {
		if node == declaration {
			width = index
		}
	}
	if width < 0 {
		t.Fatal("width's body wasn't lowered from its declaration")
	}
	built.result.Functions[width].Body = append([]ir.Statement{narrow(width)}, built.result.Functions[width].Body...)
	judged("in width", ir.Trap{Kind: "NotYet", Reason: "a narrow of a union taken unasked in function ", Site: "main.ts:4:1", Unit: "main.ts width", Phase: "checkRangeUnionReads"})
	body := built.result.Functions[width].Body
	if trap, isPanic := body[0].(ir.Panic); len(body) != 1 || !isPanic || !trap.Trap {
		t.Fatalf("width's body: got %+v, want its trap alone", body)
	}
	start, length, placed := built.firstLinkMainAt(statement)
	if !placed || length == 0 {
		t.Fatalf("console.log's statement isn't placed in Main (%d, %d, %t)", start, length, placed)
	}
	built.result.Main[start] = narrow(-1)
	judged("at the top level", ir.Trap{Kind: "NotYet", Reason: "a narrow of a union taken unasked in the program's body", Site: "main.ts:5:1", Unit: "main.ts module", Phase: "checkRangeUnionReads"})
	if trap, isPanic := built.result.Main[start].(ir.Panic); !isPanic || !trap.Trap {
		t.Fatalf("console.log's statement in Main: got %+v, want its trap", built.result.Main[start])
	}
}
