function keepFirst<T>(to: T[]): void {
	const first = to[0];
	if (first !== undefined) {
		to.push(first);
	}
}
const xs: number[] = [1];
keepFirst(xs);
console.log(`${String(xs.length)} ${String(xs[1])}`);
