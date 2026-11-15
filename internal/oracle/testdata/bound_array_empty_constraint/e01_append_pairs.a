function append<T extends {}>(to: T[], value: T | undefined): T[];
function append<T extends {}>(to: T[] | undefined, value: T): T[];
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
	if (value === undefined) return to;
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
function pairsOf(names: string[]): [number, string][] | undefined {
	let pairs: [number, string][] | undefined;
	let id = 0;
	for (const name of names) {
		pairs = append(pairs, [id, name]);
		id++;
	}
	return pairs;
}
const pairs = pairsOf(["a", "b", "c"]);
if (pairs !== undefined) {
	for (const pair of pairs) {
		console.log(`${String(pair[0])} ${pair[1]}`);
	}
}
console.log(pairsOf([]) === undefined ? "none" : "some");
