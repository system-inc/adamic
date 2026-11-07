// In .ts, as in .a: tsc's shape (core.ts:2116, #9tw2b9z): compare read off a new collator as a value, held in a const
// with no annotation as tsc holds it, and handed to compareWithCallback, a small copy of tsc's. Node
// prints every line, a10 after a2 under numeric collation. Both backends print the first two, the
// second never calling compare, and stop at the third's call in head 1's one ruled form (#zjndrx5),
// never an ordinal answer.
function compareWithCallback(a: string | undefined, b: string | undefined, comparer: (a: string, b: string) => number): number {
	return a === b ? 0 : a === undefined ? -1 : b === undefined ? 1 : comparer(a, b);
}

const comparer = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true }).compare;
console.log(`comparer ${typeof comparer}`);
console.log(`same string ${String(compareWithCallback('a', 'a', comparer))}`);
const result = compareWithCallback('a10', 'a2', comparer);
console.log(`compared ${result < 0 ? 'before' : result > 0 ? 'after' : 'equal'}`);
