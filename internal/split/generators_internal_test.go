package split

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A generator's step is one of the values its call stands for, so what it yields is what its call
// returns (Lower, 04:44 on #qk8rztp): each(boxes) yields boxes' elements, and the consumer's step
// aliases them, so the write through it is a write to the parameter. Without the yields in returns,
// the step aliased nothing and the write went unseen. Mutant: the Yield case taken out of walk's
// switch; returns[each] is empty and both checks go red.
func TestGeneratorYieldsAreItsCallsReturns(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `class Box {
	count = 0;
}
function* each(boxes: Box[]): Generator<Box, void, undefined> {
	for (const box of boxes) {
		yield box;
	}
}
function bump(boxes: Box[]): void {
	for (const box of each(boxes)) {
		box.count += 1;
	}
}
const first = new Box();
bump([first, new Box()]);
console.log(String(first.count));
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		test.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		test.Fatal(err)
	}
	each, bump := -1, -1
	for index, function := range program.Functions {
		switch function.Name {
		case "each":
			each = index
		case "bump":
			bump = index
		}
	}
	if each < 0 || bump < 0 || !program.Functions[each].Generator {
		test.Fatalf("each %d, bump %d: want the generator and its consumer", each, bump)
	}
	a := summarize(program)
	parameter := program.Functions[each].Parameters[0]
	if !a.reachable(a.returns[each])[parameter] && !a.returns[each][parameter] {
		test.Errorf(
			"each's returns %v don't reach its parameter %d: a step aliases nothing",
			a.returns[each],
			parameter,
		)
	}
	consumed := program.Functions[bump].Parameters[0]
	if !a.writes[bump][consumed] {
		test.Errorf("bump's writes %v miss its parameter %d, which it writes through a step", a.writes[bump], consumed)
	}
}
