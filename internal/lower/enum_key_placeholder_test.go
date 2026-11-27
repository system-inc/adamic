package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A placeholder field (undefined!) under an enum key, on each road that reads the key's member before the field's
// value. The numeric road (ruling (A), numericKeyRead) stops it, since no fixture holds that read to Node yet
// (#kgngvxj). Road 2 (orderedEnumFieldName, tail family 40) keeps it: its member read runs before the unset value,
// which both backends evaluate. Mutant: the stop no longer scoped to the numeric road (numericKey dropped from its
// condition) stops road 2's row too, in this stop's words.
func TestEnumKeyPlaceholderFieldByRoad(t *testing.T) {
	t.Parallel()
	const stop = "an uninitialized object field whose computed key reads an enum member"

	t.Run("numeric road stops", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, `function read(): number {
    const table = { [Pending.Low]: undefined! as number };
    return table[250];
}
enum Pending { Low = 250 }
console.log(String(read()));
`, ".ts")
		var notYet *NotYet
		if !errors.As(err, &notYet) || notYet.What != stop || !strings.Contains(notYet.Where, "main.ts:2:") {
			t.Fatalf("want NotYet %q on line 2, got %v", stop, err)
		}
	})

	t.Run("road 2 keeps its read", func(t *testing.T) {
		t.Parallel()
		program, err := lowerSourceAs(t, `enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [PollingInterval.Low]: undefined! as number, [PollingInterval.Medium]: 64 };
}
console.log(Object.keys(make()).join(","));
`, ".ts")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, function := range program.Functions {
			walk(function.Body, func(node any) bool {
				literal, ok := node.(ir.ObjectLiteral)
				if !ok || len(literal.Fields) != 2 || literal.Fields[0].Name != "250" {
					return true
				}
				_, sequenced := literal.Fields[0].Value.(ir.Comma)
				found = found || sequenced && literal.Fields[0].Uninitialized
				return true
			})
		}
		if !found {
			t.Fatal("road 2's placeholder field must stay uninitialized with its key's read before it")
		}
	})
}
