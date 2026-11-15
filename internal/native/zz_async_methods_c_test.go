package native

import (
	"context"
	"os"
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
	for _, unit := range units {
		t.Logf("unit %s: %d bytes", unit.name, len(unit.source))
	}
}
