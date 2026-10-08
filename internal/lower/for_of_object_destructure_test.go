package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestForOfObjectBindingChecks(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`for (const { value = 2 } of [{ value: 1 }]) { console.log(value.toFixed(0)); }`, "a destructured name that isn't plain"},
		{`for (const { value, ...rest } of [{ value: 1, other: 2 }]) { console.log(rest.other.toFixed(0)); }`, "a destructured name that isn't plain"},
		{`for (const { ['value']: value } of [{ value: 1 }]) { console.log(value.toFixed(0)); }`, "a computed field name"},
		{`const values = new Map<string, { readonly value: number }>(); for (const { value } of values.values()) { console.log(value.toFixed(0)); }`, "for...of over a map's values"},
		{`const values = new Map<string, { readonly value: number }>(); const iterator = values.values(); for (const { value } of iterator) { console.log(value.toFixed(0)); }`, "object destructuring over a stored collection iterator"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var stop *NotYet
			if !errors.As(err, &stop) || !strings.Contains(stop.What, probe.reason) {
				t.Fatalf("want %s, got %v", probe.reason, err)
			}
		})
	}
	t.Run("method receiver", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSource(t, `class Row { value(): number { return 1; } } for (const { value } of [new Row()]) { console.log(value().toFixed(0)); }`)
		var refusal *Refused
		if !errors.As(err, &refusal) || refusal.What != "a method in object destructuring" {
			t.Fatalf("want method receiver refusal, got %v", err)
		}
	})
}
