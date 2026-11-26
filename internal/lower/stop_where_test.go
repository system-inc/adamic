package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
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
