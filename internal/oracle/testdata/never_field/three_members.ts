// Ruling 2 past tsc's shape (#bbssdcy): three members, two of which declare name never and one of which
// doesn't declare it, so every member that declares it declares never.
interface Expression {
	readonly kind: number;
}
interface Silent {
	readonly kind: number;
	readonly name: never;
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
function kindOf(node: Expression & Arrow & Silent): string {
	return "kind " + String(node.kind);
}
console.log(kindOf(makeArrow({ kind: 3 })));
console.log(kindOf(makeArrow({ kind: 4, name: "named" })));
