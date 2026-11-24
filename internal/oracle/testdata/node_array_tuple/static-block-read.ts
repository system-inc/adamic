interface Statement { kind: number }
interface ExpressionStatement extends Statement { kind: 1; expression: string }
interface NodeArray<T extends Statement> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Statement> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
type ThisAssignment = NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }];
function expressionOf(statements: ThisAssignment): string {
	return statements[0].expression;
}
function isThisAssignment(a: NodeArray<Statement>): a is ThisAssignment {
	return a.length === 1 && a[0]?.kind === 1;
}
const first: Statement = { kind: 1, expression: "value" } as ExpressionStatement;
const made = [first] as MutableNodeArray<Statement>;
made.pos = 0;
made.end = 0;
made.hasTrailingComma = false;
made.transformFlags = 0;
const list: NodeArray<Statement> = made;
if (isThisAssignment(list)) {
	console.log("statement " + expressionOf(list));
}
