package javascript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
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

// A program lowered in first-link trap mode is emitted only where the options set FirstLinkTraps, which only
// cmd/adamic-firstlink does (TestFirstLinkTrapsOptionHasOneDoor): JavaScript and JavaScriptWith refuse it
// otherwise, naming its first trap, and the same program without its Traps is emitted as ever. Its emit with the
// option is TestFirstLinkTrapRunsUpToTheStop's (internal/lower). Mutant: JavaScriptWith's refusal dropped, and
// both emit the trapped program.
func TestFirstLinkTrapProgramIsRefusedWithoutTheOption(t *testing.T) {
	t.Parallel()
	const want = "javascript: compiler bug: a first-link program, lowered in trap mode, builds only through cmd/adamic-firstlink (first trap: trapped.a:1:1 Math.random)"
	roads := []struct {
		name string
		run  func()
	}{
		{"JavaScript", func() { JavaScript(firstLinkTrapped()) }},
		{"JavaScriptWith", func() { JavaScriptWith(firstLinkTrapped(), Options{ModuleExports: true, RuntimeImport: "./adamic.mjs"}) }},
	}
	for _, road := range roads {
		if got := firstLinkRefusal(road.run); got != want {
			t.Errorf("%s: panicked with %q, want %q", road.name, got, want)
		}
	}
	ordinary := firstLinkTrapped()
	ordinary.Traps = nil
	code := ""
	if got := firstLinkRefusal(func() { code = JavaScript(ordinary) }); got != "" || !strings.Contains(code, "firstLinkTrap(") {
		t.Errorf("the program without its Traps: panicked with %q, JavaScript %q; want it emitted, with the trap's call", got, code)
	}
}
