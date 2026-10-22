package lower

import "testing"

// Lower only: executing this reproducer would allocate more than 4 GiB.
func TestCounterBeyondSafeIntegerFromTypedArrayLengthStaysDouble(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function probe(): void {
 const values = new Uint8Array(4294967297);
 let iterations = 0;
 for (let counter = 9007199254740990; counter <= values.length * 2097152 - 2; counter++) {
  console.log(`+"`${counter}`"+`);
  iterations++;
  if (iterations === 5) { break; }
 }
}
probe();
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, local := range program.Locals {
		if local.Name == "counter" {
			if local.Counter || local.CounterGuard != nil {
				t.Fatalf("counter must stay double: integer=%t, guard=%+v", local.Counter, local.CounterGuard)
			}
			return
		}
	}
	t.Fatal("missing counter")
}
