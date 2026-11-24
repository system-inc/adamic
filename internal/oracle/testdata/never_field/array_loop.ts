// Ruling 2 past tsc's shape (#bbssdcy): the intersection held as an array's element and read in a loop by
// its other field.
interface Expression {
	readonly kind: number;
}
interface Base {
	readonly name?: string;
	readonly kind: number;
}
interface Arrow extends Base {
	readonly name: never;
}
function makeArrow(made: Base): Arrow {
	return made as Arrow;
}
function total(nodes: readonly (Expression & Arrow)[]): number {
	let sum = 0;
	for (const node of nodes) {
		sum += node.kind;
	}
	return sum;
}
console.log(String(total([makeArrow({ kind: 1 }), makeArrow({ kind: 2, name: "named" }), makeArrow({ kind: 4 })])));
