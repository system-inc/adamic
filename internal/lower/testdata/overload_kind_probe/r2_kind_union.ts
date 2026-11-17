// Road r2, a union member whose kind is itself a union of two literals. Operator's kind is Plus | Minus,
// so Extract by an Operator's kind is Operator whichever the value is, and the body's flip of Plus to
// Minus is honest: a check that the result's kind equals the argument's stops an honest program, and a
// body proof by equal kind literals refuses it. PlusOperator, a subtype of Operator with kind Plus
// alone, makes Extract<Shape, { kind: Plus }> never, which no member serves: the call still returns an
// Operator, typed never, and Node prints its symbol.
enum SyntaxKind {
	Unknown = 0,
	Plus = 1,
	Minus = 2,
	Name = 3,
}

interface Node {
	readonly kind: SyntaxKind;
}

interface Operator extends Node {
	readonly kind: SyntaxKind.Plus | SyntaxKind.Minus;
	readonly symbol: string;
}

interface Name extends Node {
	readonly kind: SyntaxKind.Name;
	readonly text: string;
}

interface PlusOperator extends Operator {
	readonly kind: SyntaxKind.Plus;
}

type Shape = Operator | Name;

function flip<T extends Shape>(node: T): Extract<Shape, Pick<T, "kind">>;
function flip(node: Shape) {
	switch (node.kind) {
		case SyntaxKind.Plus:
			return makeOperator(SyntaxKind.Minus, "-");
		case SyntaxKind.Minus:
			return makeOperator(SyntaxKind.Plus, "+");
		case SyntaxKind.Name:
			return makeName(`${node.text}'`);
	}
}

function makeOperator(kind: SyntaxKind.Plus | SyntaxKind.Minus, symbol: string): Operator {
	return { kind, symbol };
}

function makeName(text: string): Name {
	return { kind: SyntaxKind.Name, text };
}

function describe(operator: Operator): string {
	return `${operator.kind} ${operator.symbol}`;
}

const plus: Operator = { kind: SyntaxKind.Plus, symbol: "+" };
const name: Name = { kind: SyntaxKind.Name, text: "n" };
const onlyPlus: PlusOperator = { kind: SyntaxKind.Plus, symbol: "+" };

const flipped = flip(plus);
console.log(describe(flipped));
const renamed = flip(name);
console.log(renamed.text);
const gone = flip(onlyPlus);
console.log(describe(gone));
console.log("after never");
