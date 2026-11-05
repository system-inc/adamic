package lower

import "testing"

// A held type string is the checker's answer, asked again or not: typeString's first answer for each type is
// TypeToString's, and so is every later one, which comes from the lowering's map (#wy2dyf0). Mutant, red: a held answer
// returned altered.
func TestATypeStringIsTheCheckersAnswer(t *testing.T) {
	t.Parallel()
	finder, types := cycleIndexTypes(t, `declare const record: {count: number; names: string[]}; declare const choice: "left" | 1 | null; declare const rows: Array<{done: boolean}>; declare const pair: [number, string];`)
	for _, name := range []string{"record", "choice", "rows", "pair"} {
		typed := types[name]
		if typed == nil {
			t.Fatalf("no type for %s", name)
		}
		want := finder.l.checker.TypeToString(typed)
		for asked := 1; asked <= 3; asked++ {
			if got := finder.l.typeString(typed); got != want {
				t.Fatalf("%s, asked %d times: typeString %q, TypeToString %q", name, asked, got, want)
			}
		}
	}
	if len(finder.l.typeStrings) != 4 {
		t.Fatalf("held %d type strings for 4 types", len(finder.l.typeStrings))
	}
}
