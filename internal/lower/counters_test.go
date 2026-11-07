package lower

import "testing"

// Lower only: executing this reproducer would allocate more than 4 GiB.
func TestCounterBeyondSafeIntegerFromTypedArrayLengthStaysDouble(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, `function probe(): void {
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
		test.Fatal(err)
	}
	for _, local := range program.Locals {
		if local.Name == "counter" {
			if local.Counter || local.CounterGuard != nil {
				test.Fatalf("counter must stay double: integer=%t, guard=%+v", local.Counter, local.CounterGuard)
			}
			return
		}
	}
	test.Fatal("missing counter")
}
