type Shape = { readonly kind: 'Circle'; readonly radius: number } | { readonly kind: 'Square'; readonly side: number };

const shapes: readonly Shape[] = [{ kind: 'Square', side: 1 }];
for (const shape of shapes) {
	console.log('casting');
	const circle = shape as Extract<Shape, { kind: 'Circle' }>;
	console.log(`${circle.radius}`);
}
