package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestAssignmentValuePreservesCycleChecks(t *testing.T) {
	_, err := lowerSource(t, `interface Item { next: Item | undefined; }
 function make(): Item { return { next: undefined }; }
 function probe(): void {
 const parent = make();
 const child = make();
 const result = (parent.next = child);
 result.next = parent;
 }
 probe();
 `)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("want assignment-result cycle refused, got %v", err)
	}
}

func TestAssignmentValueRepresentationBoundaries(t *testing.T) {
	for _, source := range []string{
		`const values: (string | number)[] = [1]; const result = (values[0] = 'new');`,
		`const values: (boolean | undefined)[] = [true]; const result = (values[0] = false);`,
		`const holder: { item: string | number } = { item: 1 }; const result = (holder.item = 'new');`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Fatalf("want represented-storage boundary, got %v", err)
		}
	}
}
