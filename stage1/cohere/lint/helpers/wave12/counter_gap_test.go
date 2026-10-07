package wave12

import (
	"bytes"
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestCounterExactPrimitiveGap(t *testing.T) {
	want := execute(t, "", upstreamFor(t, "counter_oracle.go.txt"))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("gaps/atomic-counter.a")
	source := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry)
	if !bytes.Equal(source, want) {
		t.Fatal("exact Node counter probe disagrees with real Go")
	}
	t.Logf("actual Go and exact Node support probe match nine counter observations:\n%s", want)
	program, err := load.Load([]string{entry})
	if err == nil {
		_, err = lower.Lower(context.Background(), program)
	}
	if err == nil {
		t.Fatal("exact counter primitive gap closed; implement and validate the helper")
	}
	message := err.Error()
	if !strings.Contains(message, "BigInt") && !strings.Contains(message, "bigint") && !strings.Contains(message, "Atomics") && !strings.Contains(message, "SharedArrayBuffer") {
		t.Fatalf("unexpected compiler refusal: %s", message)
	}
	t.Logf("exact native/emitted-IR support blocked before backend emission: %s", message)
	mutant, _ := filepath.Abs("gaps/number-counter.a")
	for i, got := range backends(t, mutant, "") {
		if bytes.Equal(got, want) {
			t.Fatalf("backend %d floating-point approximation survived", i)
		}
		t.Logf("backend %d number-counter approximation compiles, finishes and is caught only by Go comparison at row %d", i, difference(got, want))
	}
}
