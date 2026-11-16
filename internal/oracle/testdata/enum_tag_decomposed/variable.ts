// A whole open enum into a union's tag slots through a variable bound to a literal (#egmxf20): no one
// member takes the variable's type, only the union, by decomposing it.
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
const made = { kind, radius: 1, side: 2 };
console.log(area(made));
