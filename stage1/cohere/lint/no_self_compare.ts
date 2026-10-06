// Port of pinned Go cohere's no-self-compare.
import type { RuleContext } from './rule_context.ts';

const messageNoSelfCompareComparingToSelf =
    'Both sides of this comparison are written identically, so the answer does not depend on the data: it is true, or it is false, every time. The two shapes this is almost always hiding are a typo in one operand, and a deliberate `x !== x` reaching for a NaN check, which `Number.isNaN(x)` says out loud instead.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-self-compare')) return;
    if(ctx.node(index).kind !== 'BinaryExpression' || ctx.comparison(index) === '') return;
    const children = ctx.node(index).children;
    if(ctx.tokens(children[0] ?? -1) === ctx.tokens(children[2] ?? -1))
        ctx.report(index, 'no-self-compare', 'comparingToSelf', messageNoSelfCompareComparingToSelf);
}
