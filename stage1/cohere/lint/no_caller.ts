// Port of pinned Go cohere's no-caller.
import type { RuleContext } from './rule_context.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-caller')) return;
    const node = ctx.node(index);
    if(node.kind !== 'PropertyAccessExpression') return;
    const receiver = ctx.unwrap(node.children[0] ?? -1);
    const key = node.children[node.children.length - 1] ?? -1;
    if(receiver < 0 || key < 0 || ctx.node(receiver).kind !== 'Identifier' || ctx.node(receiver).text !== 'arguments')
        return;
    const name = ctx.node(key).text;
    if(name !== 'caller' && name !== 'callee') return;
    ctx.report(
        key,
        'no-caller',
        'noCaller',
        `This uses \`arguments.${name}\`. Both properties throw in strict mode, which every module and class body already is, and they prevent engines from optimizing functions whose callers would otherwise be statically known. Name the function and refer to it by name instead.`,
    );
}
