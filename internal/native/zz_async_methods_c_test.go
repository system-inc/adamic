package native

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A probe, never for landing (#kha2ea0): async_methods.a's C, written where ZZ_C_OUT says.
func TestZZAsyncMethodsC(t *testing.T) {
	loaded, err := load.Load([]string{"../oracle/testdata/async_methods.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("ZZ_C_OUT"), []byte(C(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	header, units, err := splitC(C(program))
	t.Logf("split: header %d bytes, %d units, error %v", len(header), len(units), err)
	for _, unit := range append([]compilationUnit{{"units.h", header}}, units...) {
		t.Logf("unit %s: %d bytes", unit.name, len(unit.source))
		for _, line := range strings.Split(unit.source, "\n") {
			if strings.Contains(line, "Worker_work_771fdf0e0a64745c66e078ad537670ab(") || strings.Contains(line, "a_module_676b6af4303fdbd7e969e09d5f2b150d(") || strings.Contains(line, "_in(adamic_region") {
				if len(line) > 230 {
					line = line[:230]
				}
				t.Logf("  %s: %s", unit.name, line)
			}
		}
	}
}
