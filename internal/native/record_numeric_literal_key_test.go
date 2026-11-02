package native_test

import (
	"strings"
	"testing"
)

// A numeric literal names a record literal's entry by its number's canonical string, as JavaScript's
// ToPropertyKey does: 0x1F is "31", 1e3 is "1000", 1.50 is "1.5" and 1e21 is "1e+21". The keys come back
// in Node's own-key order, array indices ascending and then the rest in insertion order, so 4294967295
// (one past the last index) stays where it was written. The second record is Language's G4 shape on
// #6c4swvt, a named member beside an index signature whose slot is a union. The values are built at run
// time so a dropped retain or a doubled release reaches the sanitizer.
const recordNumericLiteralKeySource = `const made = (text: string): string => [text, 'made'].join('-');
const r: Record<string, string> = { b: made('b'), 2: made('two'), 0x1F: made('hex'), 1e3: made('exp'), 1.50: made('frac'), 0: made('zero'), 4294967295: made('big'), 1e21: made('huge') };
console.log(Object.keys(r).join(','));
console.log(String(JSON.stringify(r)));
console.log(` + "`${r[31]} ${r[1000]} ${r[1.5]} ${r[2]} ${r[3]}`" + `);
interface Recovery { b: number; [key: string]: number | string }
const recovery: Recovery = { b: 1, 2: made('two') };
console.log(Object.keys(recovery).join(','));
console.log(` + "`${recovery.b} ${String(recovery[2])}`" + `);
`

func TestRecordNumericLiteralKeys(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, recordNumericLiteralKeySource, nil)
}

// Mutant: 0x1F's entry keyed by its source text instead of its canonical string, as a lowering that
// used the literal's spelling would. Node prints hex-made for r[31], and the mutant prints undefined and
// lists 0x1F among the keys.
func TestRecordNumericLiteralKeysMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, recordNumericLiteralKeySource, func(code string) string {
		return strings.Replace(code, `"31"`, `"0x1F"`, 1)
	})
}
