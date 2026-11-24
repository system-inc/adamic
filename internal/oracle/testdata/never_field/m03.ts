interface Named {
	readonly kind: 1;
	readonly name: string;
}
interface Arrow {
	readonly kind: 2;
	readonly name: never;
	readonly body: string;
}
type Fn = Named | Arrow;
function nameOf(f: Fn): string {
	return f.kind === 1 ? f.name : "arrow";
}
function build(n: number): Fn {
	return n === 1 ? { kind: 1, name: "n" } : ({ kind: 2, body: "b" } as unknown as Arrow);
}
console.log(nameOf(build(1)) + " " + nameOf(build(2)));
