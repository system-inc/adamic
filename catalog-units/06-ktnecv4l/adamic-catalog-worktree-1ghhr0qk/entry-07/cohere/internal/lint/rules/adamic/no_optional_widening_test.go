package adamic

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoOptionalWideningFiresWhereAnOptionalPropertyIsAdded: probe h05's second step, and its shapes.
func TestNoOptionalWideningFiresWhereAnOptionalPropertyIsAdded(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		// h05
		"a re-widened value": {`const full = { x: 1, y: 'surprise' }; const narrow: { x: number } = full; const wide: { x: number; y?: number } = narrow;`, "narrow"},
		"an argument":        {`declare const narrow: { x: number }; function draw(point: { x: number; y?: number }): void {} draw(narrow);`, "narrow"},
		// t5 on #drbrp8c: Object.assign can hand a function any member, and width subtyping drops it.
		"a function's own member": {`declare const plain: (value: string) => string; const component: { (value: string): string; displayName?: string } = plain;`, "plain"},
		"a nested property":       {`declare const narrow: { readonly at: { x: number } }; const wide: { readonly at: { x: number; y?: number } } = narrow;`, "narrow"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, NoOptionalWidening, fixture.source)
			rule_testing.ExpectFindings(t, result, "optionalWidening")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestNoOptionalWideningStaysCleanWhereTheKeysAreKnown: literals, const literal aliases, and a source
// that declares the property.
func TestNoOptionalWideningStaysCleanWhereTheKeysAreKnown(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"a fresh literal":           `const wide: { x: number; y?: number } = { x: 1 };`,
		"a const literal alias":     `const point = { x: 1 }; const wide: { x: number; y?: number } = point;`,
		"the property declared":     `declare const point: { x: number; y?: number }; const wide: { x: number; y?: number } = point;`,
		"a required-only target":    `declare const point: { x: number; y: number }; const flat: { x: number } = point;`,
		"an argument literal":       `function draw(point: { x: number; y?: number }): void {} draw({ x: 1 });`,
		"an optional tuple element": `declare const none: []; const some: [number?] = none;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, NoOptionalWidening, source))
		})
	}
}

// TestNoOptionalWideningStaysCleanOnAdamicsPrograms: Adamic 0.1 compiles all ten.
func TestNoOptionalWideningStaysCleanOnAdamicsPrograms(t *testing.T) {
	t.Parallel()
	expectProgramsClean(t, NoOptionalWidening)
}
