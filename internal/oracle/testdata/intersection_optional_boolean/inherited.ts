// tsc's expressionToTypeNode.ts shape (#tq6yzqy): NumericLiteral has isUnterminated from
// LiteralLikeNode, so the member holds the field through its interface base and declares nothing of
// it itself. Node prints: 1 true, 2 false, 3 undefined.
interface LiteralLike {
	isUnterminated?: boolean;
}
interface Literal extends LiteralLike {
	text: string;
}
type Numeric = Literal & { kind: 1 };
function probe(value: Numeric): void {
	console.log(`${value.text} ${String(value.isUnterminated)}`);
}
probe({ text: "1", kind: 1, isUnterminated: true });
probe({ text: "2", kind: 1, isUnterminated: false });
probe({ text: "3", kind: 1 });
