package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsSysPresenceMembers are the optional members of System that oracle/testdata/tssys_presence tests for:
// tryEnableSourceMapsForHost and setBlocking straight off the barrel's namespace, as tsc.ts:16 and :20
// do, and writeOutputIsTTY, getMemoryUsage, base64encode and clearScreen off inner's parameter, as
// executeCommandLine.ts does. writeOutputIsTTY, tryEnableSourceMapsForHost and setBlocking are methods
// getNodeSystem's literal gives, getMemoryUsage one it omits, base64encode a field it gives an arrow and
// clearScreen a field it omits.
var tsSysPresenceMembers = []string{"tryEnableSourceMapsForHost", "setBlocking", "writeOutputIsTTY", "getMemoryUsage", "base64encode", "clearScreen"}

// tsc's ts.sys as tsc builds it (#p2n17c4 T1): an exported let made by an arrow called where it's
// written, which nests getNodeSystem and fills its own let from it, stores a field onto it and hands it
// back with !, beside a setSys nothing calls; read through a barrel's namespace and handed down a chain
// of plain calls, one from an arrow. reachSystem (reach_roads_test.go) pins the reach of a flatter twin;
// this is the shape tsc has, with getNodeSystem nested, a literal whose method shorthand reads the
// literal, the patch's store and the chain, and its presence tests: every one lowers, which is
// checkOptionalMemberPresence admitting it, a known reach of plain literals, and each tested member is
// read off its slot as it runs. setBlocking is the 15th site (tsc.ts:20), judged by the check as the
// others are, since its receiver is ts.sys, not stdout's handle. The oracle holds what it runs to
// (internal/oracle/tssys_presence_test.go). Mutants are in tssys_presence.mutants.md.
func TestTsSysPresenceLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			loaded, err := load.Load([]string{"../oracle/testdata/tssys_presence/main" + extension})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			program, err := Lower(context.Background(), loaded)
			if err != nil {
				t.Fatalf("want it to lower, every presence test admitted, got %v", err)
			}
			for _, member := range tsSysPresenceMembers {
				reads := optionalPresenceSlotReads(program, member)
				if reads == 0 {
					t.Errorf("no read of %s off its slot, want the test read as it runs, never folded", member)
				}
				t.Logf("%s: %d reads off its slot", member, reads)
			}
			wantNestedGetNodeSystemCall(t, program)
		})
	}
}

// wantNestedGetNodeSystemCall fails unless the arrow's call of getNodeSystem, the declaration nested in
// it (sys.ts:1468), is a call through a function value whose one target is getNodeSystem, and logs how
// the call is carried, which says whether its Direct is the only thing that bounds it (M-b).
func wantNestedGetNodeSystemCall(t *testing.T, program *ir.Program) {
	t.Helper()
	calls := 0
	visit := func(node any) bool {
		call, isCall := node.(ir.CallClosure)
		if !isCall {
			return true
		}
		targets := program.ClosureTargets(call)
		if targets.Unknown || len(targets.Functions) != 1 || program.Functions[targets.Functions[0]].Name != "getNodeSystem" {
			return true
		}
		calls++
		t.Logf("getNodeSystem is called with Direct %d through %T", call.Direct, call.Closure)
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	if calls != 1 {
		t.Errorf("getNodeSystem is called through a bounded function value %d times, want once: the nested call lowered some other way", calls)
	}
}
