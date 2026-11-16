// A whole open enum into a union's tag slots where the member it's built as has an optional field
// (#egmxf20).
enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius?: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius ?? -1);
	return "square " + String(shape.side * 10);
}
function other(): number { return 7; }
const kind: Kind = other();
console.log(area({ kind, side: 2 }));
