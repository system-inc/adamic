// An exhaustive switch's never handed to a never-returning call, with nothing between the switch and
// the read (#p6vawef, class (a) in Lower's scout): the read is trusted, and lowers as a trap no valid
// run reaches. Every member runs, so Node and both backends print the same. fail takes unknown, since a
// parameter typed never is not yet on its own (a value of type never, at the parameter).
type Shape = "circle" | "square";
function fail(shape: unknown): never { throw new Error("unexpected shape " + String(shape)); }
function corners(shape: Shape): number {
	switch (shape) {
		case "circle": return 0;
		case "square": return 4;
	}
	return fail(shape);
}
console.log(String(corners("circle")) + " " + String(corners("square")));
