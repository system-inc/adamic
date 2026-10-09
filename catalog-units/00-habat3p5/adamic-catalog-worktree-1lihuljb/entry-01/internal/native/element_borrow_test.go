package native

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Keep the actual benchmark's indexed declarations borrowed. The loop plan may also borrow
// bindings; count only the indexed declarations here.
func TestNbodyIndexedElementsBorrow(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../bench/nbody.ts")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	borrows, _ := planElementBorrows(program)
	declarations := 0
	for statement := range borrows {
		if _, ok := (*statement).(ir.Declare); ok {
			declarations++
		}
	}
	if declarations != 5 {
		t.Fatalf("want sun and both body/other pairs borrowed, got %d declarations", declarations)
	}
	generated := C(program)
	for statement := range borrows {
		declare, ok := (*statement).(ir.Declare)
		if !ok {
			continue
		}
		name := (&emitter{program: program}).localName(declare.Local)
		if !strings.Contains(generated, name+"_owner = NULL;") {
			t.Errorf("%s lost its borrowed element declaration", name)
		}
		if strings.Contains(generated, "adamic_release("+name+");") {
			t.Errorf("%s releases a count belonging to the array", name)
		}
	}
}

// A throw may end a borrow or be caught before its last use. Making its Error is harmless;
// calls in the message and removals in the catch still prevent borrowing.
func TestThrowElementBorrowPlan(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/borrow_element_throw.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	borrows, _ := planElementBorrows(program)
	got := map[string]bool{}
	for statement := range borrows {
		declare, ok := (*statement).(ir.Declare)
		if !ok {
			continue
		}
		local := program.Locals[declare.Local]
		got[program.Functions[local.Function].Name] = true
	}
	for _, name := range []string{"caughtInside", "leavesBorrowScope"} {
		if !got[name] {
			t.Errorf("%s: harmless Error creation prevented borrowing", name)
		}
	}
	for _, name := range []string{"invalidatesAfterCatch", "messageInvalidates"} {
		if got[name] {
			t.Errorf("%s: an element removed before its last use was borrowed", name)
		}
	}
}

// A bounded harmless closure can lend; an override, a bounded writer, or an
// unknown function value must keep the element counted across the call.
func TestCallTargetsElementBorrowPlan(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/call_targets_element.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	borrows, _ := planElementBorrows(program)
	borrowed := map[string]bool{}
	for statement := range borrows {
		local := program.Locals[(*statement).(ir.Declare).Local]
		borrowed[program.Functions[local.Function].Name] = true
	}
	if !borrowed["boundRead"] {
		t.Error("bounded harmless closure lost its element borrow")
	}
	for _, name := range []string{"virtualRead", "closureRead", "boundWrite"} {
		if borrowed[name] {
			t.Errorf("%s borrowed across a possible element write", name)
		}
	}
}

func TestDevirtualizeBorrowDocClaim(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/devirt_borrow_doc_claim.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	borrows, _ := planElementBorrows(program)
	found := map[string]bool{}
	for statement := range borrows {
		if declare, ok := (*statement).(ir.Declare); ok && program.Locals[declare.Local].Name == "first" {
			local := program.Locals[declare.Local]
			found[program.Functions[local.Function].Name] = true
		}
	}
	for _, name := range []string{"throughVirtual", "throughClosure"} {
		if !found[name] {
			t.Errorf("%s: length-only targets prevented borrowing", name)
		}
	}
}
