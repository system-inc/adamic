package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNullReferenceUsesOnlyMigratedKinds(t *testing.T) {
	for _, kind := range []ir.Type{ir.String, ir.Object, ir.Array, ir.Record, ir.Map, ir.Closure, ir.Union, ir.Weak} {
		emitted := nullReference(kind)
		if kind.UsesNullSentinel() {
			if kind != ir.String || !strings.Contains(emitted, "adamic_reference_null(adamic_kind_string)") {
				t.Fatalf("migrated kind %d: %s", kind, emitted)
			}
		} else if emitted != "NULL" {
			t.Fatalf("unmigrated kind %d must not call the runtime default case: %s", kind, emitted)
		}
	}
}
