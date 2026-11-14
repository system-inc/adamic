function firstOf<T extends {}>(to: readonly T[], fallback: T): T {
	const wide: readonly {}[] = to;
	console.log(String(wide.length));
	const first = to[0];
	return first === undefined ? fallback : first;
}
const names: string[] = ["a", "b"];
console.log(firstOf(names, "z"));
console.log(firstOf([], "z"));
