// Port of pinned Go cohere's no-ex-assign.
import type { RuleContext } from './rule_context.ts';

const messageUnexpectedExceptionAssignment =
    'This assigns to the caught exception, which discards the only reference to what actually went wrong. Anything after this point in the handler reports the new value rather than the error, so a rethrow rethrows the wrong thing and a log line describes something that never happened. The usual intent is a second binding: declare one instead of overwriting the parameter.';

function assignments(ctx: RuleContext, index: number, name: string): void {
    const node = ctx.node(index);
    let target = -1;
    if(ctx.assignment(index)) target = ctx.unwrap(node.children[0] ?? -1);
    if(
        (node.kind === 'PrefixUnaryExpression' || node.kind === 'PostfixUnaryExpression') &&
        (node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken')
    )
        target = ctx.unwrap(node.children[0] ?? -1);
    if(target >= 0 && ctx.node(target).kind === 'Identifier' && ctx.node(target).text === name)
        ctx.report(target, 'no-ex-assign', 'unexpectedExceptionAssignment', messageUnexpectedExceptionAssignment);
    for(const child of node.children) assignments(ctx, child, name);
}

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-ex-assign')) return;
    const node = ctx.node(index);
    if(node.kind !== 'CatchClause') return;
    const declaration = node.children[0] ?? -1;
    const block = node.children[1] ?? -1;
    if(declaration < 0 || block < 0 || ctx.node(declaration).kind !== 'VariableDeclaration') return;
    const name = ctx.node(declaration).children[0] ?? -1;
    if(name < 0 || ctx.node(name).kind !== 'Identifier') return;
    assignments(ctx, block, ctx.node(name).text);
}
