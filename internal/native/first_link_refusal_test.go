package native

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// firstLinkTrapped is a program as a first-link lowering leaves one: main is one trap, and Traps holds its row.
func firstLinkTrapped() *ir.Program {
	return &ir.Program{
		Source:  "trapped.a",
		Strings: []string{"NotYet: Math.random at trapped.a:1:1"},
		Main:    []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: 0}, Trap: true}},
		Traps:   []ir.Trap{{Kind: "NotYet", Reason: "Math.random", Site: "trapped.a:1:1", Unit: "trapped.a module", Phase: "Statement"}},
	}
}

// firstLinkRefusal is what run panicked with, or "" when it returned.
func firstLinkRefusal(run func()) (refusal string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			refusal = fmt.Sprint(recovered)
		}
	}()
	run()
	return ""
}

// A program lowered in first-link trap mode becomes C only where the build's options set FirstLinkTraps, which
// only cmd/adamic-firstlink does (TestFirstLinkTrapsOptionHasOneDoor): every road from an ir.Program to C refuses
// it otherwise, naming its first trap, and the same program without its Traps is written as ever. Its build with
// the option is TestFirstLinkTrapRunsUpToTheStop's (internal/lower). Mutant: cProgram's refusal dropped, and every
// road writes the trapped program's C.
func TestFirstLinkTrapProgramIsRefusedWithoutTheOption(t *testing.T) {
	t.Parallel()
	const want = "native: compiler bug: a first-link program, lowered in trap mode, builds only through cmd/adamic-firstlink (first trap: trapped.a:1:1 Math.random)"
	roads := []struct {
		name string
		run  func()
	}{
		{"C", func() { C(firstLinkTrapped()) }},
		{"CWithOptions", func() { CWithOptions(firstLinkTrapped(), EmitOptions{Options: native_runtime.Options{Sanitize: true}}) }},
		{"WASI", func() { _, _ = WASI(firstLinkTrapped(), -1) }},
		{"WASIExports", func() { _, _ = WASIExports(firstLinkTrapped(), nil) }},
		{"TSGoC", func() { _, _ = TSGoC(firstLinkTrapped()) }},
	}
	for _, road := range roads {
		if got := firstLinkRefusal(road.run); got != want {
			t.Errorf("%s: panicked with %q, want %q", road.name, got, want)
		}
	}
	ordinary := firstLinkTrapped()
	ordinary.Traps = nil
	source := ""
	if got := firstLinkRefusal(func() { source = C(ordinary) }); got != "" || !strings.Contains(source, "adamic_first_link_trap(") {
		t.Errorf("the program without its Traps: panicked with %q, C %q; want its C, with the trap's call", got, source)
	}
}
