package fuzz

import (
	"fmt"
	"path/filepath"
	"testing"
)

// A list's items go with their own commas, the trailing one too, and nothing else goes with them.
func TestSourceTreeCutsAtItems(t *testing.T) {
	t.Parallel()
	tree := newSourceTree("cut.ts", "f(a, b, c,);\n[1, 2];\n")
	var arguments, elements = -1, -1
	for list := range tree.Lists() {
		switch tree.Items(list) {
		case 3:
			arguments = list
		case 2:
			elements = list
		}
	}
	if arguments < 0 || elements < 0 {
		t.Fatalf("lists not found: %d of them", tree.Lists())
	}
	for _, test := range []struct {
		list, start, end int
		want             string
	}{
		{arguments, 0, 1, "f(b, c,);\n[1, 2];\n"},
		{arguments, 1, 2, "f(a, c,);\n[1, 2];\n"},
		{arguments, 1, 3, "f(a);\n[1, 2];\n"},
		{arguments, 0, 3, "f();\n[1, 2];\n"},
		{elements, 1, 2, "f(a, b, c,);\n[1];\n"},
		{0, 0, 1, "\n[1, 2];\n"},
	} {
		if got := tree.Without(test.list, test.start, test.end).Source(); got != test.want {
			t.Errorf("without %d [%d, %d): %q, want %q", test.list, test.start, test.end, got, test.want)
		}
	}
}

// planted fails two ways: the non-null assertion comes first, so it's the signature, and the type
// predicate after it is a second refusal that a reduction which let go of the first would drift to.
const planted = `const values: number[] = [1, 2, 3];
let total = 0;
for (const value of values) {
	total += value;
}
// The one failure that counts.
function first(list: number[], unused: string): number {
	return list[0]!;
}
console.log(` + "`${total} ${first(values, 'x')}`" + `);
function isText(value: string | number): value is string {
	return typeof value === 'string';
}
console.log(` + "`${isText('a')}`" + `);
`

// plantedMinimum is all the non-null refusal needs: the function that asserts, with the parameter it
// reads.
const plantedMinimum = `function first(list: number[]): number {
	return list[0]!;
}
`

// Reducing the planted program keeps exactly its first refusal: it ends at the known minimum, still
// refused the same way, never at the type predicate a looser reduction would keep instead.
func TestReduceKeepsTheSignature(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	original := checkout.Observe(planted, "program.a", filepath.Join(directory, "original"), "")
	signature, err := Derive(original)
	if err != nil {
		t.Fatal(err)
	}
	want := "Adamic 0.1 refuses the non-null assertion !"
	if signature.Kind != "refusal" || !signature.matches(signature.Text) || len(signature.Text) < len(want) || signature.Text[:len(want)] != want {
		t.Fatalf("derived %s, want the non-null refusal", signature)
	}
	observe := func(source string, slot int) Observation {
		return checkout.Observe(source, "program.a", filepath.Join(directory, fmt.Sprintf("slot%d", slot)), signature.Kind)
	}
	reduced := Reduce("planted.a", planted, signature, original, observe, 4, 500)
	if reduced.Source != plantedMinimum {
		t.Errorf("reduced to\n%s\nwant\n%s", reduced.Source, plantedMinimum)
	}
	if again := checkout.Observe(reduced.Source, "program.a", filepath.Join(directory, "again"), ""); !again.Has(signature) {
		t.Errorf("the reduced program fails as %v, not as %s", again.Lines, signature)
	}
}
