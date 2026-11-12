function* rows(): Generator<[string, string], void, unknown> { yield ['a', 'b']; }
function count(items: Iterable<[string, string]>): number {
	let n = 0;
	for (const item of items) n += item.length;
	return n;
}
const seen = new Set<string>(['x']);
console.log(String(count(rows())));
console.log(String(count(seen.entries())));
