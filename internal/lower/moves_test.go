package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMoveConsumesOnlyProvenOwner(test *testing.T) {
	test.Parallel()
	source := `import { parallelMap } from 'adamic';
function run(): void {
 const items = [{ value: 1 }];
 const results = parallelMap(items, (item) => { item.value += 1; return item; });
 console.log(results.length.toString());
}
run();
`
	program, lowerSourceError := lowerSource(test, source)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	found := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if mapped, ok := node.(ir.ParallelMap); ok {
				found++
				if !mapped.Moved || len(mapped.Shared) != 0 {
					test.Errorf("unproven transfer: %+v", mapped)
				}
			}
			return true
		})
	}
	if found != 1 {
		test.Fatalf("want one move, got %d", found)
	}
}

func TestMoveRejectsLaterSourceUse(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `import { parallelMap } from 'adamic';
function run(): void {
 const items = [{ value: 1 }];
 const results = parallelMap(items, (item) => { item.value += 1; return item; });
 const observed = items.length;
 console.log(observed.toString());
 console.log(results.length.toString());
}
run();
`)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) ||
		refused.What != "use after move: items.length; items was moved into parallelMap" ||
		refused.Fix != "don't use it after the parallelMap" {
		test.Fatalf("want exact use-after-move refusal, got %v", lowerSourceError)
	}
}
