function f<T extends {}>(xs: T[], fallback: T): [T, T] {
	const pair: [T, T] = [xs[0] ?? fallback, fallback];
	const wide: [{}, {}] = pair;
	wide[0] = 42;
	return pair;
}
const pair: [string, string] = f(["a"], "b");
console.log(typeof pair[0]);
