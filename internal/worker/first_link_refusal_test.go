package worker

import (
	"fmt"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

// A Worker's handler is emitted with workerBackend's options, which never set FirstLinkTraps, so a program lowered
// in first-link trap mode is refused at the emit and never becomes a Worker, naming its first trap. Mutant:
// JavaScriptWith's refusal dropped, and the trapped program is emitted as a handler.
func TestFirstLinkTrapProgramIsRefusedAsAWorker(t *testing.T) {
	t.Parallel()
	trapped := &ir.Program{
		Source:  "trapped.a",
		Strings: []string{"NotYet: Math.random at trapped.a:1:1"},
		Main:    []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: 0}, Trap: true}},
		Traps:   []ir.Trap{{Kind: "NotYet", Reason: "Math.random", Site: "trapped.a:1:1", Unit: "trapped.a module", Phase: "Statement"}},
	}
	refusal := func() (refusal string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				refusal = fmt.Sprint(recovered)
			}
		}()
		javascript.JavaScriptWith(trapped, workerBackend())
		return ""
	}()
	const want = "javascript: compiler bug: a first-link program, lowered in trap mode, builds only through cmd/adamic-firstlink (first trap: trapped.a:1:1 Math.random)"
	if refusal != want {
		t.Fatalf("the worker's emit panicked with %q, want %q", refusal, want)
	}
}
