function* names(): Generator<string, void, unknown> { yield 'g'; }
function join(values: Iterable<string>): string {
	let all = '';
	for (const value of values) all += value;
	return all;
}
const seen = new Set<string>(['x', 'y']);
console.log(join(names()));
console.log(join(seen.values()));
console.log(join(seen.keys()));
