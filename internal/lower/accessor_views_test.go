package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestAccessorInferredMixedArrayView(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, elements string }{
		{"getter first", "new A(), { total: 2 }"},
		{"literal first", "{ total: 2 }, new A()"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `class A { get total(): number { return 1; } } const items = [`+probe.elements+`]; for (const item of items) { console.log(item.total.toString()); }`)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/accessor-field-view") {
				t.Fatalf("want accessor-field-view refusal, got %v", err)
			}
		})
	}
}
