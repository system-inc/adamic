package adamic

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const shapes = `
type Circle = { readonly kind: 'Circle'; readonly radius: number };
type Square = { readonly kind: 'Square'; readonly side: number };
type Shape = Circle | Square;
interface User { readonly name: string }
declare const shape: Shape;
declare function load(): unknown;
`

// TestNoUncheckedCastFiresOnCastsNothingCouldCheck: adamic's refusal 2, and the hatches around it.
func TestNoUncheckedCastFiresOnCastsNothingCouldCheck(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		// adamic's refusal 2
		"unknown to an interface": {shapes + `const user = load() as User;`, "load() as User"},
		"through unknown":         {shapes + `const user = (shape as unknown) as User;`, "(shape as unknown) as User"},
		"from any":                {shapes + `declare const raw: any; const user = raw as User;`, "raw as User"},
		"from an any array":       {shapes + `declare const rows: any[]; const users = rows as User[];`, "rows as User[]"},
		"an angle bracket":        {shapes + `const user = <User>load();`, "<User>load()"},
		"members with no tag":     {`type A = { readonly a: number }; type B = { readonly b: number }; declare const either: A | B; const a = either as A;`, "either as A"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, NoUncheckedCast, fixture.source)
			rule_testing.ExpectFindings(t, result, "uncheckedCast")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestNoUncheckedCastStaysCleanOnChecksAndUpcasts: what Adamic compiles, checked or not needing a check.
func TestNoUncheckedCastStaysCleanOnChecksAndUpcasts(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"a discriminated member":    shapes + `const circle = shape as Circle;`,
		"typeof apart":              `declare const value: string | number; const text = value as string;`,
		"instanceof apart":          `class Cat { meow(): string { return ''; } } class Dog { bark(): string { return ''; } } declare const pet: Cat | Dog; const cat = pet as Cat;`,
		"undefined apart":           shapes + `declare const maybe: User | undefined; const user = maybe as User;`,
		"an upcast":                 shapes + `const any = shape as Shape | undefined;`,
		"as const":                  `const pair = [1, 2] as const;`,
		"to unknown":                shapes + `const opaque = shape as unknown;`,
		"an any array to unknown[]": `declare const rows: any[]; const opaque = rows as unknown[];`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, NoUncheckedCast, source))
		})
	}
}

// TestNoUncheckedCastStaysCleanOnAdamicsPrograms: Adamic 0.1 compiles all ten.
func TestNoUncheckedCastStaysCleanOnAdamicsPrograms(t *testing.T) {
	t.Parallel()
	expectProgramsClean(t, NoUncheckedCast)
}
