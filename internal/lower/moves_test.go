package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMoveConsumesOnlyProvenOwner(t *testing.T) {
	t.Parallel()
	source := `import { parallelMap } from 'adamic';
function run(): void {
 const items = [{ value: 1 }];
 const results = parallelMap(items, (item) => { item.value += 1; return item; });
 console.log(results.length.toString());
}
run();
`
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if mapped, ok := node.(ir.ParallelMap); ok {
				found++
				if !mapped.Moved || len(mapped.Shared) != 0 {
					t.Errorf("unproven transfer: %+v", mapped)
				}
			}
			return true
		})
	}
	if found != 1 {
		t.Fatalf("want one move, got %d", found)
	}
}

func TestMoveRejectsLaterSourceUse(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `import { parallelMap } from 'adamic';
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
	if !errors.As(err, &refused) || refused.What != "use after move: items.length; items was moved into parallelMap" || refused.Fix != "don't use it after the parallelMap" {
		t.Fatalf("want exact use-after-move refusal, got %v", err)
	}
}
