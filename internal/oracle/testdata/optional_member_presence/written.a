/*
 * A presence test of an optional field holding a function (#p2n17c4 shape 1), read off its slot as it
 * runs, after the slot is written: a literal that gives m as undefined, then a function written in, then
 * undefined written back, and a literal that gives a function, then undefined written in. Every literal
 * gives m, since a write to a field some literal omits is not yet, and the form is the field's, since a
 * write to an optional method stays refused. m's type holds undefined, as tsc's own host types write it,
 * since under exactOptionalPropertyTypes no other optional field takes undefined.
 */
interface Probed {
	readonly n: number;
	m?: (() => number) | undefined;
}

function probe(h: Probed): string {
	return `${typeof h.m} ${h.m ? 'has' : 'none'}`;
}

const h: Probed = { n: 1, m: undefined };
console.log(probe(h));
h.m = (): number => 2;
console.log(probe(h));
h.m = undefined;
console.log(probe(h));
const g: Probed = { n: 2, m: (): number => 3 };
console.log(probe(g));
g.m = undefined;
console.log(probe(g));
