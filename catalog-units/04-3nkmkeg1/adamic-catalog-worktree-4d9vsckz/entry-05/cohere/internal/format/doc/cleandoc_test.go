package doc

import (
	"math/rand"
	"testing"
)

// TestCleanDocMatchesMapDoc checks CleanDoc's own walk against upstream's definition, mapDoc(doc,
// cleanDocFn), on the random specs the printer differential uses, without the fork (#r89mksm). Both
// results are compared structurally and printed, and the doc CleanDoc was given must read the same
// afterwards, since an unchanged part is now handed back rather than copied.
func TestCleanDocMatchesMapDoc(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(20261005))
	changed := 0
	for index := range 3000 {
		generator := &specGenerator{random: random}
		spec := generator.node()
		document := buildSpec(spec)
		before := showDoc(document)

		cleaned := CleanDoc(document)
		// Read before printing: the printer marks groups broken in place, and the cleaned doc shares its
		// unchanged groups with the input. A mark follows from the group's contents alone, so a shared
		// group is marked as each copy would have been.
		if after := showDoc(document); after != before {
			t.Fatalf("case %d: CleanDoc changed its input from\n%s\nto\n%s", index, before, after)
		}
		expected := MapDoc(buildSpec(spec), cleanDocFn)
		if got, want := showDoc(cleaned), showDoc(expected); got != want {
			t.Fatalf("case %d: CleanDoc built\n%s\nmapDoc(cleanDocFn) built\n%s", index, got, want)
		}
		if showDoc(cleaned) != before {
			changed++
		}
		if got, want := printOrError(cleaned), printOrError(expected); got != want {
			t.Fatalf("case %d: CleanDoc printed %q, mapDoc(cleanDocFn) %q", index, got, want)
		}
	}
	if changed == 0 {
		t.Fatal("no spec was changed by cleaning, so the comparison never left the unchanged path")
	}
}

// printOrError prints a doc as the differentials do, or the panic it raises.
func printOrError(document Doc) (result string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = "error"
		}
	}()
	return Print(document, Options{PrintWidth: 20, TabWidth: 2})
}
