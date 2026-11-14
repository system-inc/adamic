function f<T extends {}>(xs: T[]): Record<string, T> {
	const record: Record<string, T> = {};
	for (const x of xs) {
		record["k"] = x;
	}
	const wide: Record<string, {}> = record;
	wide["n"] = 42;
	return record;
}
const names: Record<string, string> = f(["a"]);
console.log(typeof names["n"]);
