package lower

import (
	"strings"
	"testing"
)

func TestMapSetKeyViewsStayRefused(t *testing.T) {
	t.Parallel()
	for index, source := range []string{
		"const narrow = new Map<number, string>([[1, 'one']]); const wide: ReadonlyMap<number | string, string> = narrow; console.log(`${wide.get(1)}`);",
		"const narrow = new Set<number>([1]); const wide: ReadonlySet<number | string> = narrow; console.log(`${wide.has(1)}`);",
	} {
		name := "Map"
		if index == 1 {
			name = "Set"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			if err == nil || !strings.Contains(err.Error(), "boxed collection key view") {
				t.Fatalf("want key-view refusal, got %v", err)
			}
		})
	}
}
