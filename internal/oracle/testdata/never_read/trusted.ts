// A switch over a shape's kind that leaves it never, handed to a never-returning call with nothing
// between the switch and the read (#p6vawef, class (a) in Lower's scout): the read is trusted, and
// lowers as a trap no valid run reaches. Every member runs, so Node and both backends print the same.
// fail takes unknown, since a parameter typed never is not yet on its own (a value of type never, at the
// parameter).
enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle }
interface Square { readonly kind: Kind.Square }
type Shape = Circle | Square;
function fail(shape: unknown): never { throw new Error("unexpected shape " + String(shape)); }
function corners(shape: Shape): number {
	switch (shape.kind) {
		case Kind.Circle: return 0;
		case Kind.Square: return 4;
	}
	return fail(shape);
}
console.log(String(corners({ kind: Kind.Circle })) + " " + String(corners({ kind: Kind.Square })));
