// Port of pinned Go cohere's no-useless-catch.
import type { RuleContext } from './rule_context.ts';

const messageUnnecessaryCatch =
    'This `catch` does nothing but rethrow what it caught, so the whole `try` has no effect: the error propagates exactly as it would with no `try` at all. It reads as if the error is being handled, which is the harm, since the next reader has to check. Delete the `try` and `catch` and keep the body, or handle the error here.';
const messageUnnecessaryCatchClause =
    'This `catch` does nothing but rethrow what it caught, so it changes nothing: the error propagates as it would without it, and the `finally` still runs either way. Delete the `catch` clause and keep `try`/`finally`, or handle the error here.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-useless-catch')) return;
    const node = ctx.node(index);
    if(node.kind !== 'CatchClause') return;
    const declaration = node.children[0] ?? -1;
    if(declaration < 0 || ctx.node(declaration).kind !== 'VariableDeclaration') return;
    const name = ctx.node(declaration).children[0] ?? -1;
    const block = node.children[1] ?? -1;
    if(name < 0 || block < 0 || ctx.node(name).kind !== 'Identifier') return;
    const first = ctx.node(block).children[0] ?? -1;
    if(first < 0 || ctx.node(first).kind !== 'ThrowStatement') return;
    let value = ctx.node(first).children[0] ?? -1;
    while(value >= 0 && ['ParenthesizedExpression', 'NonNullExpression', 'AsExpression'].includes(ctx.node(value).kind))
        value = ctx.node(value).children[0] ?? -1;
    if(value < 0 || ctx.node(value).kind !== 'Identifier' || ctx.node(value).text !== ctx.node(name).text) return;
    const parent = ctx.parent(index);
    if(parent < 0 || ctx.node(parent).kind !== 'TryStatement') return;
    const hasFinally = ctx.node(parent).children.length === 3;
    ctx.report(
        hasFinally ? index : parent,
        'no-useless-catch',
        hasFinally ? 'unnecessaryCatchClause' : 'unnecessaryCatch',
        hasFinally ? messageUnnecessaryCatchClause : messageUnnecessaryCatch,
    );
}
