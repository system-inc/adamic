interface Statement { kind: number }
interface ExpressionStatement extends Statement { kind: 1; expression: string }
interface NodeArray<T extends Statement> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Statement> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface StaticBlock { statements: NodeArray<Statement> }
function createClassThisAssignmentBlock(statements: NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }]): StaticBlock & { statements: NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }] } {
	return { statements };
}
function isThisAssignment(a: NodeArray<Statement>): a is NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }] {
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
	const block = createClassThisAssignmentBlock(list);
	console.log("block " + block.statements[0].expression);
}
