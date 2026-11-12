function* first(): Generator<boolean | undefined, void, unknown> { yield true; yield undefined; }
function show(values: Iterable<boolean | undefined>): void {
	for (const value of values) console.log(value === undefined ? 'none' : value ? 'yes' : 'no');
}
const flags = new Map<string, boolean | undefined>([['a', false], ['b', undefined]]);
show(first());
show(flags.values());
