package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// dictionaryViewDeclarations is tsc's MapLike (core.ts) and an OptionsBase-shaped dictionary whose values are held
// as a Union, for each probe to cast to.
const dictionaryViewDeclarations = "interface MapLike<T> { [index: string]: T } interface OptionsBase { [option: string]: string | number | boolean | undefined } "

// TestDictionaryViewCasts holds where the checked dictionary view is taken (#2v3xtgm, re-ruled Nov 23) and where a
// cast keeps castIndexSignature's NotYet at the cast: only object, undefined beside it where the target keeps it,
// cast to a dictionary of exactly unknown values, in a program that writes through no dictionary the target is
// assignable to. An admitted cast lowers to its one helper, a checked narrow to a record that also tests the record
// holds references, in the two run-time NotYet's words. Mutant: dictionaryViewWritten answers false, and
// written_through_a_dictionary and deleted_through_a_dictionary go red, admitted.
func TestDictionaryViewCasts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		// stop is the NotYet's words at the cast, or "" where the view is admitted.
		stop string
	}{
		{"maplike of unknown", "function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; }", ""},
		{"record of unknown", "function f(x: object): Record<string, unknown> { return x as Record<string, unknown>; }", ""},
		{"undefined kept", "function f(x: object | undefined): MapLike<unknown> | undefined { return x as MapLike<unknown> | undefined; }", ""},
		{"undefined dropped", "function f(x: object | undefined): MapLike<unknown> { return x as MapLike<unknown>; }",
			"a checked dictionary view of MapLike<unknown> over object | undefined"},
		{"number values", "function f(x: object): Record<string, number> { return x as Record<string, number>; }",
			"a checked dictionary view of Record<string, number> over object"},
		{"union-held values", "function f(x: object): OptionsBase { return x as OptionsBase; }",
			"a checked dictionary view of OptionsBase over object"},
		{"written through a dictionary", "function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; } function g(m: MapLike<unknown>): void { m['a'] = 1; }",
			"a checked dictionary view of MapLike<unknown> over object"},
		{"deleted through a dictionary", "function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; } function g(m: MapLike<unknown>, k: string): void { delete m[k]; }",
			"a checked dictionary view of MapLike<unknown> over object"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, dictionaryViewDeclarations+probe.source, ".ts")
			if probe.stop != "" {
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "stage 0 can't lower "+probe.stop+" yet") {
					t.Fatalf("want the NotYet %q at the cast, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want the view admitted, got %v", err)
			}
			found := false
			for _, function := range program.Functions {
				if function.Name != "dictionary_view" {
					continue
				}
				found = true
				returned, isReturn := function.Body[len(function.Body)-1].(ir.Return)
				narrow, isNarrow := returned.Value.(ir.Narrow)
				if !isReturn || !isNarrow || !narrow.Checked || narrow.To != ir.Record || narrow.Message != ir.DictionaryViewRecordStop || narrow.ReferencesMessage != ir.DictionaryViewReferencesStop {
					t.Fatalf("want the helper to return a checked narrow to a record with both stops, got %#v", function.Body)
				}
			}
			if !found {
				t.Fatal("no dictionary_view helper lowered")
			}
		})
	}
}

// A source with fields of its own keeps a stop, whichever rule makes it: the view takes object alone, so a fixed
// object's static type never reaches a record's read through it (#mft3zc6).
func TestDictionaryViewKeepsAFixedSourceStopped(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, dictionaryViewDeclarations+"interface Package { readonly name: string } function f(x: Package): MapLike<unknown> { return x as MapLike<unknown>; }", ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "MapLike<unknown>") {
		t.Fatalf("want a NotYet naming MapLike<unknown>, got %v", err)
	}
}

// .a keeps main's refusal for the view's own cast, as item 8 refuses every checked downcast there.
func TestDictionaryViewRefusedInAdamic(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, dictionaryViewDeclarations+"function f(x: object): MapLike<unknown> { return x as MapLike<unknown>; }", ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Adamic 0.1 refuses a cast the runtime can't check; ") {
		t.Fatalf(".a: want main's refusal word for word, got %v", err)
	}
}
