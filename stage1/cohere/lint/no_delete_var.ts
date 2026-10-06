// Port of pinned Go cohere's no-delete-var.
import type { RuleContext } from './rule_context.ts';

const messageUnexpectedDeleteVar =
    '`delete` removes a property from an object; it cannot remove a variable. Applied to a bare name this does nothing in strict mode and throws a SyntaxError where strict mode is enforced, which every module in this codebase is. So the line either has no effect or breaks the file, and neither is what it reads as doing. To drop a reference, assign undefined; to remove a property, delete the property.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-delete-var')) return;
    if(ctx.node(index).kind !== 'DeleteExpression') return;
    const target = ctx.unwrap(ctx.node(index).children[0] ?? -1);
    if(target >= 0 && ctx.node(target).kind === 'Identifier')
        ctx.report(index, 'no-delete-var', 'unexpectedDeleteVar', messageUnexpectedDeleteVar);
}
