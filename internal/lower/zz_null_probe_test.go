package lower

import (
	"testing"
)

// A probe, never for landing: does a checked closure's membership test on a union's null sentinel
// reach adamic_instanceof's class read? The claim Box | number lies on null, so the check at the call
// must stop with exit 70; a sanitizer report instead is a read past adamic_null's end.
func TestProbeNullSentinelReachesInstanceof(t *testing.T) {
	t.Parallel()
	source := `class Box {
  readonly size: number;
  constructor(size: number) { this.size = size; }
}
function count(values: (Box | number | string | null)[], test: (value: Box | number | string | null) => value is Box | number): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count([new Box(1), 2, "s", null], (value): value is Box | number => typeof value !== "string")));
`
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Logf("PROBE lower: %v", err)
		return
	}
	code, stdout, stderr := brandedNative(t, program)
	t.Logf("PROBE native exit=%d stdout=%q stderr=%q", code, stdout, stderr)
}
