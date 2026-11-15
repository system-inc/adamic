// The open value narrowed by === to one member before it goes into the tag slot (#egmxf20), lowers.
enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius);
	return "square " + String(shape.side * 10);
}
function make(value: number): string {
	const kind: Kind = value;
	if (kind === Kind.Circle) {
		return area({ kind, radius: 1 });
	}
	return "open " + String(kind);
}
console.log(make(0) + " " + make(7));
