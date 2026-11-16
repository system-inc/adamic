// A whole open enum into a union's tag slots through a property (#egmxf20): tsc splits Kind into
// Kind.Circle | Kind.Square and builds the literal as either member, but kind holds 7.
enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius);
	return "square " + String(shape.side * 10);
}
function other(): number { return 7; }
const kind: Kind = other();
console.log(area({ kind: kind, radius: 1, side: 2 }));
