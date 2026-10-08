package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestElementTypeObjectArraysCannotCloseCycles(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function run(): void { const values: object[] = []; values.push(values); } run();",
		"function run(): void { const values: object[] = []; values.push(() => values.length); } run();",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "cycle") {
			t.Errorf("object array cycle must be refused, got %v", err)
		}
	}
}
