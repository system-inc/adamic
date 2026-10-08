package lower

import (
	"errors"
	"testing"
)

func TestOptionalSortKeepsNumericDefaultLimit(t *testing.T) {
	_, err := lowerSource(t, `function sorted(values:number[],comparer?: (a:number,b:number)=>number):number[]{return values.sort(comparer);} console.log(sorted([10,2]).join(","));`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "an optional sort comparator on non-string elements; narrow the comparator to a present function before sorting" {
		t.Fatalf("want existing non-string default-sort limit, got %v", err)
	}
}
