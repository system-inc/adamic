// Port of pinned Go cohere's no-compare-neg-zero.
import type { RuleContext } from './rule_context.ts';

function negativeZero(ctx: RuleContext, index: number): boolean {
    const outer = ctx.unwrap(index);
    if(outer < 0 || ctx.node(outer).kind !== 'PrefixUnaryExpression' || ctx.node(outer).operator !== 'MinusToken')
        return false;
    const value = ctx.unwrap(ctx.node(outer).children[0] ?? -1);
    return value >= 0 && ctx.node(value).kind === 'NumericLiteral' && ctx.node(value).text === '0';
}

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-compare-neg-zero')) return;
    const operator = ctx.comparison(index);
    if(operator === '') return;
    const children = ctx.node(index).children;
    if(!negativeZero(ctx, children[0] ?? -1) && !negativeZero(ctx, children[2] ?? -1)) return;
    ctx.report(
        index,
        'no-compare-neg-zero',
        'unexpected',
        `Comparing against \`-0\` with \`${operator}\`: every comparison operator treats \`-0\` and \`0\` as the same number, so this is true for positive zero too and does not test what it looks like it tests. \`Object.is(x, -0)\` is the only comparison that distinguishes them.`,
    );
}
