package lower

import (
	"strings"
	"testing"
)

func TestArrayTailDefaultSortBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const values: (number | undefined)[] = [1, undefined]; values.sort();`,
		`const values = [{value: 1}, {value: 2}]; values.sort();`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "default sort requires dense primitive") {
			t.Fatalf("want default sort representation refusal, got %v", err)
		}
	}
	for _, source := range []string{`[1, , 2].sort();`, `new Array(3).sort();`} {
		if _, err := lowerSource(t, source); err == nil {
			t.Fatalf("sparse array compiled: %s", source)
		}
	}
}

func TestArrayTailReduceRightBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const values: number[] = []; values.reduceRight((sum, value) => sum + value);`,
		`const values = [1,2]; values.reduceRight((sum, value) => sum + value);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "initial value") {
			t.Fatalf("want missing-initial refusal, got %v", err)
		}
	}
	for _, source := range []string{
		`const values=[1,2]; values.sort=()=>values; values.sort();`,
		`const values=[1,2]; values.reduceRight=()=>{throw new Error("override");}; values.reduceRight((sum,value)=>sum+value,0);`,
	} {
		if _, err := lowerSource(t, source); err == nil {
			t.Fatalf("intrinsic override compiled: %s", source)
		}
	}
}
