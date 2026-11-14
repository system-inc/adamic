function inner<U>(ys: U[], value: U): void {
	ys.push(value);
}
function f<T extends {}>(xs: T[], value: T): T[] {
	inner<{}>(xs, 42);
	xs.push(value);
	return xs;
}
const names: string[] = f(["a"], "b");
for (const name of names) {
	console.log(typeof name);
}
