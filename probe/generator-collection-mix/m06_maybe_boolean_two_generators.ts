function* first(): Generator<boolean | undefined, void, unknown> { yield true; yield undefined; }
function* second(): Generator<boolean | undefined, void, unknown> { yield false; }
function show(values: Iterable<boolean | undefined>): void {
	for (const value of values) console.log(value === undefined ? 'none' : value ? 'yes' : 'no');
}
show(first());
show(second());
