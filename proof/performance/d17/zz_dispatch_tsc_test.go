package lower

import (
	"context"
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// TestZZDispatchScanOverTsc is a proof-only probe (#d1chbm7, never merged): ruling 17's scan over tsc's adapted
// program, ADAMIC_DISPATCH_TSC naming its src/tsc/tsc.ts, printing every admitted site and every declined
// candidate with its reason, and whether condition 3's program-wide accessors flag is set.
func TestZZDispatchScanOverTsc(t *testing.T) {
	root := os.Getenv("ADAMIC_DISPATCH_TSC")
	if root == "" {
		t.Skip("ADAMIC_DISPATCH_TSC unset")
	}
	program, err := load.Load([]string{root})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	roads := l.dispatchCasts()
	for assertion, site := range roads.sites {
		t.Logf("ADMITTED %s keys=%d calls=%d presence=%d", l.program.Where(assertion), len(site.keys), len(site.calls), len(site.presenceTests))
		for _, key := range site.keys {
			if len(key.name) > 0 && (key.name == "301" || key.name == "302") {
				t.Logf("KEY %s members %s", key.name, l.checker.TypeToString(key.members))
			}
		}
	}
	for assertion, reason := range roads.declines {
		t.Logf("DECLINED %s: %s", l.program.Where(assertion), reason)
	}
	writes := l.propertyWritesOf()
	t.Logf("ACCESSORS %t, accessorName(kind) %t", writes.accessors, l.accessorName("kind"))
}
