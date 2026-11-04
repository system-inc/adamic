// return fail(), fail returning never, from a function whose result is a number, a string, a string
// enum, an array or an object: the call runs, and its throw reaches the caller's try. The value after
// it is a placeholder held as the function returns, never as the call's own representation
// (#xv8kybc, tsc's `return Debug.fail(...)`).
function fail(message: string): never {
	throw new Error("unreachable " + message);
}
const enum Tag {
	Div = "div",
	Span = "span",
}
function count(kind: number): number {
	if (kind === 1) return 1;
	if (kind === 2) return 3;
	return fail("count " + String(kind));
}
function name(kind: number): string {
	if (kind === 1) return "one" + "!";
	return fail("name " + String(kind));
}
function tag(kind: number): Tag {
	if (kind === 1) return Tag.Div;
	if (kind === 2) return Tag.Span;
	return fail("tag " + String(kind));
}
function list(kind: number): number[] {
	if (kind === 1) return [kind, kind + 1];
	return fail("list " + String(kind));
}
interface Box {
	readonly value: string;
}
function box(kind: number): Box {
	if (kind === 1) return { value: "box" + String(kind) };
	return fail("box " + String(kind));
}
function attempt(label: string, run: () => string): void {
	try {
		console.log(label + " " + run());
	} catch (error) {
		if (error instanceof Error) console.log(label + " caught " + error.message);
	}
}
for (const kind of [1, 2, 3]) {
	attempt("count", () => String(count(kind)));
	attempt("name", () => name(kind));
	attempt("tag", () => tag(kind));
	attempt("list", () => list(kind).join("+"));
	attempt("box", () => box(kind).value);
}
