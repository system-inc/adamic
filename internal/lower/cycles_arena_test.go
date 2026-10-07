package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// The empty backing literal has type never[], which is assignable to both a
// number[] and a record array but cannot own a record. The allocator really
// captures the arena, so accepting this fixture cannot rely on ignoring closures.
func TestNumericHandleArena(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../oracle/testdata/numeric_handle_arena.a")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if _, err := lowerSource(t, source); err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(source, "interface CollapseCopyNode {", "interface CollapseCopyNode {\n    parent: CollapseCopyNode[];", 1)
	mutant = strings.Replace(mutant, "contextPresent: true, context, children,", "contextPresent: true, context, children, parent: arena,", 1)
	_, err = lowerSource(t, mutant)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("parent-array mutant must be refused, got %v", err)
	}
	// Pin the whole diagnostic, normalizing only the temporary directory.
	message := refused.Error()
	directory := message[:strings.Index(message, "main.a:")]
	message = strings.ReplaceAll(message, directory, "")
	const want = "main.a:10:5: Adamic 0.1 refuses CollapseCopyNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free, and the write at main.a:27:9 may close one (the value written reaches something this function didn't make or let escape, and what it's written into wasn't made here); declare the elements weak, Weak<CollapseCopyNode>[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly CollapseCopyNode[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)"
	if message != want {
		t.Fatalf("parent-array mutant diagnostic changed:\ngot  %s\nwant %s", message, want)
	}
	t.Log(message)
	// A compatible concrete list can hide a genuine owner or closure. Those
	// structural edges remain followed even though its empty backing holds nothing.
	for _, hidden := range []string{"owner: arena", "callback: () => arena.length"} {
		real := strings.Replace(strings.Replace(source, "const empty: CollapseNodeList", "const empty", 1), "values: [] };", "values: backing, "+hidden+" };", 1)
		real = strings.Replace(real, "const empty =", "const backing: number[] = [];\n    const empty =", 1)
		_, err := lowerSource(t, real)
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/cycle-capable") {
			t.Errorf("hidden %s must remain refused, got %v", hidden, err)
		}
	}
	// The reach exception relies on the existing invariant view gate: retaining
	// a never[] alias while populating it through a number[] must be refused.
	_, err = lowerSource(t, "const bottom: never[] = [];\nconst values: number[] = bottom;\nvalues.push(1);\n")
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/invariant-mutable") {
		t.Errorf("a writable wider alias of never[] must be refused, got %v", err)
	}
}
