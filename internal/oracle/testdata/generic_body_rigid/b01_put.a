function put<T extends {}>(to: T[], value: T): T[] {
	const wide: {}[] = to;
	wide.push(1);
	return to;
}
const names: string[] = put(["a"], "b");
const second = names[1];
if (second !== undefined) {
	console.log(second.toUpperCase());
}
