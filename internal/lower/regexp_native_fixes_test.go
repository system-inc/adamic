package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRegExpGroupCompoundRefusals(t *testing.T) {
	// These four operators accept string | undefined in TypeScript. Numeric compound
	// assignments to this dictionary are rejected by tsc, so they are not stage 0 work.
	for _, target := range []string{"g.y", "g['y']", "g.x"} {
		for _, operator := range []string{"+=", "&&=", "||=", "??="} {
			source := "const m = /(?<x>a)/.exec('a'); if (m !== null && m.groups !== undefined) { const g = m.groups; " + target + " " + operator + " 'q'; }"
			_, err := lowerSource(t, source)
			if operator == "&&=" || operator == "||=" {
				// Adamic 0.1's logical-assignment refusal precedes regex lowering.
				var refused *Refused
				if !errors.As(err, &refused) {
					t.Errorf("want 0.1 refusal for %s: %v", operator, err)
				}
				continue
			}
			var notYet *NotYet
			// ??= is already NotYet before either operand is lowered.
			if !errors.As(err, &notYet) || (operator != "??=" && !strings.Contains(err.Error(), "overriding a RegExp or iterator property")) {
				t.Errorf("want regex assignment NotYet for %s %s, got %v", target, operator, err)
			}
		}
	}
	if _, err := lowerSource(t, "const r = /a/g; r.lastIndex += 1; console.log(`${r.lastIndex}`);"); err != nil {
		t.Fatalf("supported lastIndex compound update: %v", err)
	}
}
