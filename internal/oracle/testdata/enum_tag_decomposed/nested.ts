// A whole open enum into a union's tag slots through a literal nested in another (#egmxf20).
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
interface Holder { readonly shape: Shape }
const holder: Holder = { shape: { kind, radius: 1, side: 2 } };
console.log(area(holder.shape));
