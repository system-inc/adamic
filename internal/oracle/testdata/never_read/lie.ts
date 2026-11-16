// trusted.ts's corners handed a shape whose kind lies (#p6vawef): a numeric enum
// holds any number, so other's 7 is a Kind to the checker, and { kind } is a Shape only by tsc
// decomposing the union. Node reaches fail and throws. Adamic refuses the program at 20:30, where the
// whole Kind goes into Circle's or Square's tag slot (#egmxf20), so the never read never sees the lie.
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
console.log(String(corners({ kind })));
