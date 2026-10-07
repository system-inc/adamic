type Shape =
	| { readonly kind: 'Circle'; readonly radius: number }
	| { readonly kind: 'Rectangle'; readonly width: number; readonly height: number }
	| { readonly kind: 'Triangle'; readonly base: number; readonly height: number };

function area(shape: Shape): number {
	switch (shape.kind) {
		case 'Circle':
			return Math.PI * shape.radius ** 2;
		case 'Rectangle':
			return shape.width * shape.height;
		case 'Triangle':
			return (shape.base * shape.height) / 2;
	}
}

const shapes: readonly Shape[] = [
	{ kind: 'Circle', radius: 1.5 },
	{ kind: 'Rectangle', width: 3, height: 4 },
	{ kind: 'Triangle', base: 6, height: 2.5 },
];

let total = 0;
for (const shape of shapes) {
	const shapeArea = area(shape);
	total += shapeArea;
	console.log(`${shape.kind}: ${shapeArea.toFixed(3)}`);
}
console.log(`Total: ${total.toFixed(3)} (${total})`);
