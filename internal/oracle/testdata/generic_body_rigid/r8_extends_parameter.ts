function f<T extends U, U>(xs: T[], u: U): T[] {
	const wide: U[] = xs;
	wide.push(u);
	return xs;
}
const names: "a"[] = ["a"];
f<"a", string>(names, "b");
for (const name of names) {
	console.log(name);
}
