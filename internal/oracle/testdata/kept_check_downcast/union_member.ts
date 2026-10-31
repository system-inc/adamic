// kept_check_downcast_refused/union_member.a as .ts, where the cast stays a checked cast: its test passes, so the
// program prints what Node prints.
type Circle = { readonly kind: 'Circle'; readonly radius: number };
type Square = { readonly kind: 'Square'; readonly side: number };
type Shape = Circle | Square;
function radius(shape: Shape): number {
	const circle = shape as Circle;
	return circle.radius;
}
console.log(`radius ${radius({ kind: 'Circle', radius: 2 })}`);
