// The sites' shape for ruling 2 (#bbssdcy): tsc's Expression & ArrowFunction, an intersection one of whose
// members declares name never, handed a value whose real object has no name slot and one whose object
// has one, and read only by its other fields. Before the ruling the parameter stopped at the intersection.
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
function kindOf(node: Expression & Arrow): string {
	return "kind " + String(node.kind);
}
console.log(kindOf(makeArrow({ kind: 1 })));
console.log(kindOf(makeArrow({ kind: 2, name: "named" })));
