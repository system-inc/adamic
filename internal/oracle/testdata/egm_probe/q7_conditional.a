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
console.log(String(corners({ kind: Kind.Circle })));
function other(): number { return 7; }
const kind: Kind = other();
const flag = other() > 1;
console.log(String(corners(flag ? { kind } : { kind: Kind.Square })));
