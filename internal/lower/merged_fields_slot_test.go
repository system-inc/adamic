package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestMergedFieldUnknownSlot(t *testing.T) {
	t.Parallel()
	const prefix = "class Merged { a = 1; }\ninterface Merged { f?: string | bigint; }\nconst m = new Merged();\n"
	for _, probe := range []struct{ name, access string }{
		{"read", "console.log(typeof m.f);"},
		{"bracket read", "console.log(typeof m['f']);"},
		{"write", "m.f = 'x';"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, prefix+probe.access, ".ts")
			var stop *NotYet
			if !errors.As(err, &stop) || !strings.Contains(stop.What, "accessing f from the interface merge of class Merged without a known slot representation") {
				t.Fatalf("want the merged slot diagnostic, got %v", err)
			}
		})
	}
	t.Run("unused", func(t *testing.T) {
		t.Parallel()
		program := lowersAndAgreesWithNodeAs(t, prefix+"console.log(String(m.a));", ".ts")
		if len(program.Classes) != 1 || program.Classes[0].Name != "Merged" || holdsSlot(program.Classes[0].Fields, "f") {
			t.Fatal("want the Merged layout without a slot for f")
		}
	})
}
