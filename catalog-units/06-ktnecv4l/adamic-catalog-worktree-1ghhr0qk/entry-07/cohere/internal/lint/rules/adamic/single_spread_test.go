package adamic

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestSingleSpreadFiresOnEverySpreadAfterAnything: probes h06 and h07, and the shapes between them.
func TestSingleSpreadFiresOnEverySpreadAfterAnything(t *testing.T) {
	t.Parallel()
	const declarations = `declare const a: { x: number }; declare const b: { y: number }; declare const c: { z: number };` + "\n"
	for name, fixture := range map[string]struct {
		source string
		spans  []string
	}{
		// h06
		"two spreads": {declarations + `const merged = { ...a, ...b };`, []string{"...b"}},
		// h07
		"a field, then a spread":  {declarations + `const merged = { x: 1, ...b };`, []string{"...b"}},
		"three spreads":           {declarations + `const merged = { ...a, ...b, ...c };`, []string{"...b", "...c"}},
		"after a spread literal":  {declarations + `const merged = { ...(Math.random() > 0.5 ? { y: 1 } : {}), ...b };`, []string{"...b"}},
		"a method, then a spread": {declarations + `const merged = { run(): void {}, ...b };`, []string{"...b"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, SingleSpread, fixture.source)
			rule_testing.ExpectFindings(t, result, repeated("laterSpread", len(fixture.spans))...)
			expectSpans(t, fixture.source, result, fixture.spans...)
		})
	}
}

// TestSingleSpreadStaysCleanOnTheOneSafeForm: one leading spread, then fields (h08), and spreads of
// literals, which carry exactly their type's keys.
func TestSingleSpreadStaysCleanOnTheOneSafeForm(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// h08
		"a leading spread, then fields": `declare const b: { y: number }; const merged = { ...b, x: 1 };`,
		"a conditional literal":         `declare const b: { y: number }; const merged = { ...b, ...(Math.random() > 0.5 ? { x: 1 } : {}) };`,
		"a literal spread":              `declare const b: { y: number }; const merged = { ...b, ...{ x: 1 } };`,
		"no spread":                     `const point = { x: 1, y: 2 };`,
		"an array spread":               `declare const xs: number[]; declare const ys: number[]; const all = [...xs, ...ys];`,
		"a destructuring rest":          `declare const point: { x: number; y: number }; const { x, ...rest } = point;`,
		"a destructuring assignment":    `declare const point: { x: number; y: number }; let x = 0; let rest = {}; ({ x, ...rest } = point);`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, SingleSpread, source))
		})
	}
}

// TestSingleSpreadStaysCleanOnAdamicsPrograms: program 9's `{ ...tree, left }` is the form Adamic allows.
func TestSingleSpreadStaysCleanOnAdamicsPrograms(t *testing.T) {
	t.Parallel()
	expectProgramsClean(t, SingleSpread)
}
