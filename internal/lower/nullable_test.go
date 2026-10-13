package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNullableReferencesNeedAnEmptyCaseTag(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function f(v: string | null | undefined): string { return typeof v; }",
		"function f(v?: string | null): string { return typeof v; }",
		"function f(v: string | null = 'default'): string { return typeof v; }",
		"function f(v: readonly string[] | null | undefined): string { return typeof v; }",
		"function f(v: { readonly a: string } | null | undefined): string { return typeof v; }",
		"class C { readonly a = 'a'; } function f(v: C | null | undefined): string { return typeof v; }",
		"function f(v: Map<string,string> | null | undefined): string { return typeof v; }",
		"function f<T>(v: T): string { return typeof v; } f<string | null | undefined>(null);",
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), "empty-case tag") {
			t.Errorf("%s: got %v, want the named empty-case tag reason", source, err)
		}
	}
}

func TestNullableReferenceViewsCannotChangeTheEmptyCase(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function f(v: string | null): void { const wider: string | null | undefined = v; }",
		"function f(v: string | undefined): void { const wider: string | null | undefined = v; }",
		"function f(v: { readonly x: string | null }): void { const wider: { readonly x?: string | null } = v; }",
		"function f(v: { readonly x: string | undefined }): void { const wider: { readonly x?: string | null | undefined } = v; }",
		"function f(v: readonly (string | null)[]): void { const wider: readonly (string | null | undefined)[] = v; }",
		"function f<T>(v: { readonly x: T }): void { const wider: { readonly x?: T } = v; } f<{readonly a:string}|null>({x:null});",
		"function f<T>(v: readonly T[]): void { const wider: readonly (T | undefined)[] = v; } f<string|null>([null]);",
		"function f<T>(v: T): void { const wider: T | undefined = v; } f<string|null>(null);",
		"function f<T>(v: T): void { const wider: T | null = v; } f<string|undefined>(undefined);",
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) {
			t.Errorf("%s: got %v, want NotYet", source, err)
		}
	}
}

func TestNullableReferencesKeepAssertionsAndLooseEqualityRefused(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function f(v: string|null): boolean { return v == null; }",
		"function f(v: string|undefined): boolean { return v != undefined; }",
		"function f(v: string|null): string { return v!; }",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Errorf("%s: got %v, want Refused", source, err)
		}
	}
}

func TestNullableReferenceReadsNeedATagInEveryExpression(t *testing.T) {
	t.Parallel()
	prelude := "const seen = new Map<string, RegExpExecArray | null>(); seen.set('miss', /z/.exec('abc')); const found: (RegExpExecArray | null)[] = []; found.push(/z/.exec('abc')); "
	for _, observed := range []string{
		"console.log(`absent: undefined=${seen.get('absent') === undefined} null=${seen.get('absent') === null}`); console.log(`miss: undefined=${seen.get('miss') === undefined} null=${seen.get('miss') === null}`); console.log(`past end: undefined=${found[3] === undefined} null=${found[3] === null}`); console.log(`stored: undefined=${found[0] === undefined} null=${found[0] === null}`);",
		"console.log(`${seen.get('absent') === null}`);",
		"console.log(`${seen.get('miss') === undefined}`);",
		"console.log(`${found[3] === null}`);",
		"console.log(`${found[0] === undefined}`);",
		"console.log(`${seen.get('miss')}`);",
		"function pass<T>(v: T): string { return typeof v; } console.log(pass(seen.get('miss')));",
		"if (seen.has('miss')) { console.log(`${seen.get('miss') === null}`); }",
		"if (seen.get('miss') !== undefined) { console.log(`${seen.get('miss') === null}`); }",
	} {
		_, err := lowerSource(t, prelude+observed)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), "empty-case tag") {
			t.Errorf("%s: got %v, want the named empty-case tag reason", observed, err)
		}
	}
}

// The former typeof refusal predated typeof-null-2. The oracle holds the full
// typeof_null_slots.a program to Node with the nullable pointer layout.
func TestNullableReferenceTypeOfKeepsLookupPresence(t *testing.T) {
	t.Parallel()
	for _, observed := range []string{
		"typeof found[0]",
		"typeof (found[3])",
		"typeof found.at(0)",
		"typeof found.pop()",
		"typeof seen.get('miss')",
		"typeof (seen.get('absent'))",
	} {
		source := "const seen = new Map<string, RegExpExecArray | null>(); const found: (RegExpExecArray | null)[] = []; console.log(" + observed + ");"
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("%s: got %v, want acceptance with lookup presence", observed, err)
		}
	}
}
