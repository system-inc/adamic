// Two members of a three-member enum into a union's tag slots (#egmxf20): excluding Kind.Triangle by
// !== leaves 7 in, so k is as open as the whole enum.
enum Kind { Circle, Square, Triangle }
interface Circle { readonly kind: Kind.Circle; readonly radius: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius);
	return "square " + String(shape.side * 10);
}
function made(k: Kind.Circle | Kind.Square): string {
	return area({ kind: k, radius: 1, side: 2 });
}
function other(): number { return 7; }
const kind: Kind = other();
if (kind !== Kind.Triangle) {
	console.log(made(kind));
}
