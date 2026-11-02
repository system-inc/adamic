package native_test

import "testing"

// && on a record is its right side where the record is present and undefined where it's missing:
// a record is an object, so it's falsy only when it's missing, and a bare record pointer is never
// null (expression.go). This is tsc's getLocaleSpecificMessage (utilities.ts:8608), a record of
// localized messages that may not be loaded, read with && and then || for the fallback. The strings
// are built at run time so a dropped retain or a doubled release reaches the sanitizer.
const logicalRecordOperandSource = `let table: Record<string, string> | undefined;
function message(key: string, fallback: string): string {
    return table && table[key] || fallback;
}
console.log(message('a', 'fallback-a'));
table = { a: ['localized', 'a'].join('-') };
console.log(message('a', 'fallback-a'));
console.log(message('b', 'fallback-b'));
const empty: Record<string, string> = {};
console.log(empty && 'value');
function made(present: boolean, key: string): Record<string, string> | undefined {
    if (!present) {
        return undefined;
    }
    const record: Record<string, string> = {};
    record[key] = ['made', 'here'].join(' ');
    return record;
}
const again = made(true, 'built');
console.log(` + "`${again && again['built']}`" + `);
console.log(` + "`${again && again['absent']}`" + `);
console.log(` + "`${made(false, 'built') && 'never'}`" + `);
const kept = made(true, 'built') && ['joined', 'late'].join(' ');
console.log(` + "`${kept}`" + `);`

// Node prints fallback-a, localized-a, fallback-b, value, made here, undefined, undefined and
// joined late. Lowered with the missing arm as the right side instead of undefined, native prints
// never where Node prints undefined, and this goes red.
func TestLogicalRecordOperand(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, logicalRecordOperandSource, nil)
}
