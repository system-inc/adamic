/*
 * A presence test of an optional field holding a function, which a program's interface declares
 * (#p2n17c4 shape 1), read off its slot as it runs: typeof, !, ?: and the left of && on an arrow, a
 * literal's method shorthand and a literal without the field; a condition guarding the call, which keeps
 * its receiver; and ?. on a receiver that's missing, one with the field and one without.
 */
interface Probed {
	readonly n: number;
	m?: () => number;
}

function probe(h: Probed): string {
	const said = h.m && 'and';
	return `${typeof h.m} ${String(!h.m)} ${h.m ? 'has' : 'none'} ${said ?? 'neither'}`;
}

function called(h: Probed): string {
	if (h.m) {
		return String(h.m());
	}
	return 'none';
}

function maybe(h: Probed | undefined): string {
	return h?.m ? 'has' : 'none';
}

console.log(probe({ n: 1, m: (): number => 1 }));
console.log(probe({ n: 2, m() { return 2; } }));
console.log(probe({ n: 3 }));
console.log(called({ n: 3, m() { return this.n; } }));
console.log(called({ n: 4 }));
console.log(maybe(undefined));
console.log(maybe({ n: 5, m: (): number => 5 }));
console.log(maybe({ n: 6 }));
