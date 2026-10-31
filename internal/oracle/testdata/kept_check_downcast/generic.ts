// kept_check_downcast_refused/generic.a as .ts, where the cast stays a checked cast: its test passes, so the
// program prints what Node prints.
type Circle<T> = { readonly kind: 'Circle'; readonly radius: T };
type Square<T> = { readonly kind: 'Square'; readonly side: T };
type Shape<T> = Circle<T> | Square<T>;
function radius<T>(shape: Shape<T>): T {
	return (shape as Circle<T>).radius;
}
console.log(`radius ${radius<number>({ kind: 'Circle', radius: 2 })}`);
