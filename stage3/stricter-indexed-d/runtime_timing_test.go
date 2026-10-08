package indexed_d

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Not parallel: before/after timing must not compete with this unit's other builds.
func TestNullableGuardTiming(t *testing.T) {
	if os.Getenv("ADAMIC_NULL_BENCH") == "" {
		t.Skip("explicit before/after runtime measurement")
	}
	const source = `const values: (string | null)[] = ["abc", null];
let total = 0;
for (let i = 0; i < 100000000; i++) {
 const value: string | null = values[i % 2];
 if (value !== null) { total += value.length; }
}
console.log(String(total));
`
	program, _, node := nullableProgram(t, source, false)
	if checks := ir.InsertedChecks(program); len(checks) != 1 || checks[0].Kind != "indexed-presence" {
		t.Fatalf("guarded loop inventory: %+v", checks)
	}
	binary := filepath.Join(t.TempDir(), "loop")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 5; trial++ {
		start := time.Now()
		got := run(binary)
		elapsed := time.Since(start)
		if got != node {
			t.Fatalf("timed loop: %+v, Node %+v", got, node)
		}
		t.Logf("%s trial=%d iterations=100000000 elapsed=%.6fs Node=%q", os.Getenv("ADAMIC_NULL_BENCH"), trial+1, elapsed.Seconds(), node.stdout)
	}
}
