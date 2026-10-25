// Spec: A downcast to a member of a union is checkable
// Expect: lowers
type Shape = { readonly kind: 'Circle'; readonly radius: number } | { readonly kind: 'Square'; readonly side: number };
const shapes: readonly Shape[] = [{ kind: 'Circle', radius: 2 }];
for (const shape of shapes) {
	const circle = shape as Extract<Shape, { kind: 'Circle' }>;
	console.log(`${circle.radius}`);
}
