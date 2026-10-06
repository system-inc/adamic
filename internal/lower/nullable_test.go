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
