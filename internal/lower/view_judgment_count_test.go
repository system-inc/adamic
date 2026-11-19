package lower

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// viewJudgmentSource is a program of casts functions, each reading a field through a kind-union cast into
// checked views, and writes plain writes of that checked field at the top level, the shape Loom job
// J-9bt4dja timed: every cast's views walk every write.
func viewJudgmentSource(casts, writes int) string {
	var source strings.Builder
	source.WriteString(kindUnionWriteModel) // defined in cast_kind_union_test.go
	for index := range casts {
		fmt.Fprintf(&source, "function read%d(node: Node): number { return (node as PropertyAccessExpression | ElementAccessExpression).flags + %d; }\n", index, index)
	}
	for index := range writes {
		fmt.Fprintf(&source, "built.flags = %d;\n", index)
	}
	source.WriteString("let total = 0;\n")
	for index := range casts {
		fmt.Fprintf(&source, "total += read%d(node);\n", index)
	}
	source.WriteString("console.log(String(total));\n")
	return source.String()
}

// A checked access is judged about once per lowering however many views walk it: the indexed walk keeps its
// judgment and reads it back (keepsViewJudgments), where it judged every known access again at every view, so
// lowering cost casts times accesses (#715zjph). Twenty casts beside a hundred writes judge about 2,000 times
// without the keep, and this holds them to twice the program's casts and writes. A count, not a time, so a slow
// machine can't redden it. Mutant: the keep dropped, viewSiteJudged never filled, and the count is about 20 × 100.
//
// Not parallel: judgeViewAccessCalls counts every judgment in the package while it's set, and no parallel test
// runs beside a sequential one.
func TestViewJudgmentKeptPerCheckedAccess(t *testing.T) {
	const casts, writes = 20, 100
	var calls atomic.Int64
	judgeViewAccessCalls.Store(&calls)
	defer judgeViewAccessCalls.Store(nil)
	if _, err := lowerSourceAs(t, viewJudgmentSource(casts, writes), ".ts"); err != nil {
		t.Fatal(err)
	}
	judged := calls.Load()
	if judged == 0 {
		t.Fatal("the fixture's views judged no access, so the count says nothing")
	}
	if limit := int64(2 * (casts + writes)); judged > limit {
		t.Fatalf("%d casts beside %d writes judged %d accesses, more than %d: a known access is judged again at every view", casts, writes, judged, limit)
	}
}
