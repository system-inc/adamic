function f<T extends {}>(xs: T[]): T[] {
	const widen = (): void => {
		const wide: {}[] = xs;
		wide.push(42);
	};
	widen();
	return xs;
}
const names: string[] = f(["a"]);
for (const name of names) {
	console.log(typeof name);
}
