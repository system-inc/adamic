package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsSysPresenceMembers are the optional members of System that oracle/testdata/tssys_presence_twin tests
// for: tryEnableSourceMapsForHost and setBlocking straight off the barrel's namespace, as tsc.ts:16 and :20
// do, and writeOutputIsTTY, getMemoryUsage, base64encode and clearScreen off inner's parameter, as
// executeCommandLine.ts does. writeOutputIsTTY, tryEnableSourceMapsForHost and setBlocking are methods
// getNodeSystem's literal gives, getMemoryUsage one it omits, base64encode a field it gives an arrow and
// clearScreen a field it omits.
var tsSysPresenceMembers = []string{"tryEnableSourceMapsForHost", "setBlocking", "writeOutputIsTTY", "getMemoryUsage", "base64encode", "clearScreen"}

// tsc's ts.sys as tsc builds it (#p2n17c4 T1), oracle/testdata/tssys_presence: an exported let made by an
// arrow called where it's written, which nests getNodeSystem and fills its own let from it and hands it
// back, beside a setSys nothing calls; read through a barrel's namespace and handed down a chain of plain
// calls, one from an arrow. getNodeSystem's literal has a method shorthand, createDirectory, that reads
// the literal itself (sys.ts:1546), and that stops lowering at the literal's declaration before any
// presence test is judged: a function value that captures the variable its own initializer declares,
// which is #dv99xzy. This pins that stop, so the day #dv99xzy lands this goes red, and tsc's shape
// should then lower as its twin does (TestTsSysPresenceLowers) and join the oracle's run of it.
func TestTsSysSelfReadingLiteralStops(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			loaded, err := load.Load([]string{"../oracle/testdata/tssys_presence/main" + extension})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), loaded)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != "a function value that captures the variable its own initializer declares" || !strings.HasSuffix(notYet.Where, "system"+extension+":41:9") {
				t.Fatalf("want a not yet at nodeSystem's declaration (system%s:41:9), a function value that captures the variable its own initializer declares (#dv99xzy), got %v", extension, err)
			}
		})
	}
}

// tsc's ts.sys's twin (#p2n17c4 T1), oracle/testdata/tssys_presence_twin: tsc's shape above with
// createDirectory deciding without reading the literal, so it gets past #dv99xzy. reachSystem
// (reach_roads_test.go) pins the reach of a flatter twin; this is the shape tsc has, with getNodeSystem
// nested, a literal of method shorthands and shorthand properties, and the chain, and its presence tests
// (tsc's patch of writeFile is left out: it reads the method as a value first, which stays refused, so
// tsc stops there anyway): every one lowers, which is checkOptionalMemberPresence admitting it, a known
// reach of plain literals, and each tested member is read off its slot as it runs. setBlocking is the
// 15th site (tsc.ts:20), judged by the check as the others are, since its receiver is ts.sys, not
// stdout's handle. The oracle holds what it runs to (internal/oracle/tssys_presence_test.go). Mutants are
// in tssys_presence.mutants.md.
func TestTsSysPresenceLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			loaded, err := load.Load([]string{"../oracle/testdata/tssys_presence_twin/main" + extension})
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
// it (sys.ts:1468), is a call through a function value whose one target is getNodeSystem, and logs
// whether ClosureTargets still bounds the call with its Direct taken away, which says whether Direct is
// the only thing that bounds it (M-b). It asks ClosureTargets of a copy rather than reading the call's
// Closure, which only internal/ir's allow list may read.
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
		undirected := call
		undirected.Direct = 0
		t.Logf("getNodeSystem is called with Direct %d, bounded without it: %t", call.Direct, !program.ClosureTargets(undirected).Unknown)
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
