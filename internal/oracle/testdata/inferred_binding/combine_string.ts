function isArray(value: unknown): value is readonly unknown[] {
	return Array.isArray(value);
}
function append<T extends {}>(to: T[], value: T | undefined): T[];
function append<T extends {}>(to: T[] | undefined, value: T): T[];
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
	if (value === undefined) return to;
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
function combine<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined): T | T[] | undefined {
	if (xs === undefined) return ys;
	if (ys === undefined) return xs;
	if (isArray(xs)) return isArray(ys) ? [...xs, ...ys] : append(xs, ys);
	if (isArray(ys)) return append(ys, xs);
	return [xs, ys];
}
const out = combine<string>(['b', 'c'], 'a');
if (out !== undefined && Array.isArray(out)) {
	for (const x of out) {
		console.log(x);
	}
}
