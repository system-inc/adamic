package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
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

// plantedCompiler stands in for adamic c: it writes C that clang refuses for each marker the program
// still holds, ALPHA's error before BETA's, and C clang accepts once neither is left. Adamic's own C has
// only one way clang refuses it today (typeof 'a', a literal's address compared with NULL), and a test
// that leaned on that bug would break the day it's fixed, so the two refusals are planted here.
const plantedCompiler = `#!/bin/sh
[ "$1" = c ] || exit 0
echo 'int main(void) {'
if grep -q ALPHA "$2"; then echo '	int alpha = 0; if (&alpha == 0) { return 1; }'; fi
if grep -q BETA "$2"; then echo '	int beta = 0; if (beta = 1) { return 2; }'; fi
echo '	return 0;'
echo '}'
`

// plantedClang holds both markers. The reducer's first cut, the program's first half gone, leaves
// BETA alone: a reduction that took any refusal clang makes would keep that cut and drift to it.
const plantedClang = `console.log('ALPHA');
console.log('one');
console.log('two');
console.log('BETA');
`

// Reducing C clang refuses keeps clang's first error, with its numbers left out: it ends with ALPHA
// alone, never at BETA, which a reduction that took any refusal would keep instead.
func TestReduceKeepsClangsError(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	runtime, err := native.RuntimeLibrary(filepath.Join("..", "native", "runtime"), native.Options{Sanitize: true})
	if err != nil {
		t.Fatal(err)
	}
	compiler := filepath.Join(directory, "adamic")
	if err := os.WriteFile(compiler, []byte(plantedCompiler), 0o755); err != nil {
		t.Fatal(err)
	}
	checkout := &Checkout{Root: "../..", directory: directory, adamic: compiler, runtime: runtime}
	original := checkout.Observe(plantedClang, "program.a", filepath.Join(directory, "original"), "")
	signature, err := Derive(original)
	if err != nil {
		t.Fatal(err)
	}
	want := "error: comparison of address of 'alpha' equal to a null pointer is always false [-Werror,-Wtautological-pointer-compare]"
	if signature.Kind != "cc" || signature.Text != want || !signature.Exact {
		t.Fatalf("derived %s, want cc:%s", signature, want)
	}
	for written, has := range map[string]bool{"cc:tautological-pointer-compare": true, "cc:-Wparentheses": true, "cc:no such warning": false} {
		parsed, err := ParseSignature(written)
		if err != nil || original.Has(parsed) != has {
			t.Errorf("%s matches what clang said: %v, want %v (%v)", written, !has, has, err)
		}
	}
	observe := func(source string, slot int) Observation {
		return checkout.Observe(source, "program.a", filepath.Join(directory, fmt.Sprintf("slot%d", slot)), signature.Kind)
	}
	reduced := Reduce("planted.a", plantedClang, signature, original, observe, 4, 200)
	if want := "console.log('ALPHA');\n"; reduced.Source != want {
		t.Errorf("reduced to\n%s\nwant\n%s", reduced.Source, want)
	}
	if again := checkout.Observe(reduced.Source, "program.a", filepath.Join(directory, "again"), ""); !again.Has(signature) {
		t.Errorf("the reduced program fails as %q, not as %s", again.Lines["cc"], signature)
	}
}

// A crash's signature is the crash, the program's own: the crash kind, its whole line. A candidate is
// kept only when it crashes the same way, never when it has drifted to a plain output difference or
// to another crash. compiler-panic, the compiler's own panic, is a different kind that a crash never
// matches.
func TestReduceKeepsTheCrash(t *testing.T) {
	t.Parallel()
	original := Observation{Verdict: Crash, Lines: map[string]string{"crash": "native AddressSanitizer: SEGV"}}
	signature, err := Derive(original)
	if err != nil {
		t.Fatal(err)
	}
	if signature.Kind != "crash" || signature.Text != "native AddressSanitizer: SEGV" || !signature.Exact {
		t.Fatalf("derived %s, want the exact crash", signature)
	}
	for _, test := range []struct {
		name      string
		candidate Observation
		want      bool
	}{
		{"the same crash", Observation{Verdict: Crash, Lines: map[string]string{"crash": "native AddressSanitizer: SEGV"}}, true},
		{"a plain output difference instead", Observation{Verdict: Finding, Lines: map[string]string{"mismatch": "native exit: node 0, native 1"}}, false},
		{"the crash's words as a finding", Observation{Verdict: Finding, Lines: map[string]string{"crash": "native AddressSanitizer: SEGV"}}, false},
		{"another crash", Observation{Verdict: Crash, Lines: map[string]string{"crash": "native signal: segmentation fault"}}, false},
		{"no failure", Observation{Verdict: Agreed, Lines: map[string]string{}}, false},
	} {
		if got := keeps(signature, original, test.candidate); got != test.want {
			t.Errorf("%s: kept %v, want %v", test.name, got, test.want)
		}
	}
	for written, has := range map[string]bool{"crash:SEGV": true, "crash:AddressSanitizer": true, "compiler-panic:SEGV": false, "mismatch:SEGV": false} {
		parsed, err := ParseSignature(written)
		if err != nil || original.Has(parsed) != has {
			t.Errorf("%s matches the crash: %v, want %v (%v)", written, !has, has, err)
		}
	}
	compiler := Observation{Verdict: Finding, Lines: map[string]string{"compiler-panic": "panic: runtime error: index out of range [3] with length 2"}}
	if derived, err := Derive(compiler); err != nil || derived.Kind != "compiler-panic" {
		t.Errorf("the compiler's panic derived %s (%v), want compiler-panic", derived, err)
	}
}
