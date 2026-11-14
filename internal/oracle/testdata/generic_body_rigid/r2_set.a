function f<T extends {}>(xs: T[]): Set<T> {
	const set = new Set<T>(xs);
	const wide: Set<{}> = set;
	wide.add(42);
	return set;
}
const names: Set<string> = f(["a"]);
for (const name of names) {
	console.log(typeof name);
}
