package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fusedHarness computes what a fused multiply-add would get wrong. The operands come from volatile
// variables, so clang can't fold the arithmetic away before it chooses the instructions.
const fusedHarness = `#include <stdio.h>

int main(void) {
	volatile double tenth = 0.1, ten = 10, one = 1, third = 1.0 / 3;
	double a = tenth, b = ten, c = one, x = third;
	printf("%.17g %.17g %.17g\n", a * b - c, x * 3 - c, x * x - x * x);
	return 0;
}
`

// JavaScript rounds each multiply and each add on its own, so a native program must never fuse them,
// on a processor that could. This compiles for one that can (Haswell, the first x86 with FMA) and
// runs it here, which works only where this machine has FMA too.
func TestArithmeticIsNeverFused(t *testing.T) {
	t.Parallel()
	cpuinfo, err := os.ReadFile("/proc/cpuinfo")
	if err != nil || !strings.Contains(string(cpuinfo), " fma ") {
		t.Skip("this machine has no fused multiply-add to test against (or no /proc/cpuinfo to say)")
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "fused")
		if err := Build(fusedHarness, binary, Options{Sanitize: sanitize, cpu: "haswell"}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != "0 0 0\n" {
			t.Errorf("sanitize %v: got %q, want \"0 0 0\\n\", as JavaScript computes it", sanitize, got)
		}
	}
}
