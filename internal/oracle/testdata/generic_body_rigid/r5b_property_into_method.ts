function f<T extends {}>(xs: T[]): T[] {
	const sink: { m: (x: T) => void } = {
		m: (x: T): void => {
			xs.push(x);
		},
	};
	const w: { m(x: {}): void } = sink;
	w.m(42);
	return xs;
}
const names: string[] = f(["a"]);
for (const name of names) {
	console.log(typeof name);
}
