// The open value compared by === with one member before it's written (#egmxf20): the comparison of a const
// local proves the member (enumComparedMember, #snypn8c c3), so it lowers.
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
