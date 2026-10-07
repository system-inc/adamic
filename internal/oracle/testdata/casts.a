type Shape =
	| { readonly kind: 'Circle'; readonly radius: number }
	| { readonly kind: 'Square'; readonly side: number }
	| { readonly kind: 'Line'; readonly length: number };

const shapes: readonly Shape[] = [
	{ kind: 'Circle', radius: 2 },
	{ kind: 'Line', length: 3 },
];
const first = shapes[0];
if (first !== undefined) {
	const circle = first as Extract<Shape, { kind: 'Circle' }>;
	console.log(`${circle.radius}`);
}
for (const shape of shapes) {
	const notSquare = shape as Exclude<Shape, { kind: 'Square' }>;
	console.log(notSquare.kind);
}
const point = { x: 1, y: 2 } as const;
console.log(`${point.x + point.y}`);

interface Named {
	readonly name: string;
}

const named = { name: 'Ken', age: 83 } as Named;
console.log(named.name);
