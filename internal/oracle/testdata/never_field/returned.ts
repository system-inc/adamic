// Ruling 2 past tsc's shape (#bbssdcy): a function returning the intersection, the shape of family 13's
// ImmediatelyInvokedArrowFunction, whose result asks the same representation.
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
function wrap(made: Base): Expression & Arrow {
	return makeArrow(made);
}
const wrapped = wrap({ kind: 5 });
console.log("kind " + String(wrapped.kind));
