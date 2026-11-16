// One member literal into a union's tag slot is that member's own promise (#egmxf20), and lowers.
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
console.log(area({ kind: Kind.Circle, radius: 1 }) + " " + area({ kind: Kind.Square, side: 2 }) + " " + String(kind));
