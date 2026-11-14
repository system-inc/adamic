function putMap<T extends {}>(to: Map<string, T>, value: T): Map<string, T> {
	const wide: Map<string, {}> = to;
	wide.set("x", 1);
	to.set("v", value);
	return to;
}
const names: Map<string, string> = putMap(new Map<string, string>(), "b");
for (const name of names.values()) {
	console.log(typeof name);
}
