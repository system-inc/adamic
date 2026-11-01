package lower

import (
	"errors"
	"strings"
	"testing"
)

// Only a type reference has type arguments, and the checker dereferences nil asking anything else for
// them. Each site here asked a type a program writes, a union or a string, and panicked (#egxccg8, the
// sweep after the rest parameter's): a destructuring declaration judged for nominal ancestry, and
// Object.entries's pairs written into a wider array.
func TestTypeArgumentsAreReadOnlyFromReferences(t *testing.T) {
	t.Parallel()
	const pick = "function pick(flag: boolean): [string, number] | [number, string] {\n    return flag ? [\"a\", 1] : [1, \"a\"];\n}\n"
	for _, probe := range []struct{ name, source, where, what string }{
		{"destructured union of tuples", "class Box {}\n" + pick + "const [a, b]: [string, number] | [number, string] = pick(true);\nconsole.log(`${a}${b}`);\n", ":5:7", "NotYet destructuring anything but a tuple into [names]"},
		{"destructured string", "class Box {}\nconst [c, d]: string = \"xy\";\nconsole.log(`${c}${d}`);\n", ":2:7", "NotYet destructuring anything but a tuple into [names]"},
		// A union of tuples is still judged for nominal ancestry, member by member, as one tuple is.
		{"destructured union seen as a class", "class Box {\n    value = 1;\n}\nfunction plain(flag: boolean): [{ value: number }] | [{ value: number }, number] {\n    return flag ? [{ value: 2 }] : [{ value: 3 }, 4];\n}\nconst [box]: [Box] | [Box, number] = plain(true);\nconsole.log(`${box.value}`);\n", ":7:8", "Refused a destructured value without nominal ancestry seen as Box"},
		{"entries seen as unknown", "const shape = { a: 1, b: 2 };\nconst pairs: readonly unknown[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n", ":2:35", "Refused Object.entries: tsc's entries must have a proven value type"},
		{"entries seen as a union of tuples", "const shape = { a: 1, b: 2 };\nconst pairs: readonly ([string, number] | [string, string])[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n", ":2:65", "Refused Object.entries: tsc's entries must have a proven value type"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				defer func() {
					if failure := recover(); failure != nil {
						t.Fatalf("panicked: %v", failure)
					}
				}()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				var refused *Refused
				got, where := "", ""
				switch {
				case errors.As(err, &notYet):
					got, where = "NotYet "+notYet.What, notYet.Where
				case errors.As(err, &refused) && strings.HasPrefix(refused.What, "Object."):
					got, where = "Refused "+refused.What+": "+refused.Fix, refused.Where
				case errors.As(err, &refused):
					got, where = "Refused "+refused.What, refused.Where
				}
				if got != probe.what || !strings.HasSuffix(where, probe.where) {
					t.Fatalf("want %s at %s, got %v", probe.what, probe.where, err)
				}
			})
		}
	}
}
