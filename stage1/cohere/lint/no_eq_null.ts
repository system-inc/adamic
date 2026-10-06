// Port of pinned Go cohere's no-eq-null.
import type { RuleContext } from './rule_context.ts';

const messageNoEqNull =
    'This compares against `null` with `==` or `!=`, which also matches `undefined`. That coercion is usually what the author wanted and is never what the code says, so a reader cannot tell a deliberate both-values check from a typo. Write `=== null` when only null is meant, or `== null` deliberately somewhere the convention says so.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-eq-null')) return;
    if(ctx.node(index).kind !== 'BinaryExpression') return;
    const children = ctx.node(index).children;
    const operator = ctx.node(children[1] ?? -1).kind;
    if(operator !== 'EqualsEqualsToken' && operator !== 'ExclamationEqualsToken') return;
    if(
        ctx.node(ctx.unwrap(children[0] ?? -1)).kind === 'NullKeyword' ||
        ctx.node(ctx.unwrap(children[2] ?? -1)).kind === 'NullKeyword'
    )
        ctx.report(index, 'no-eq-null', 'unexpected', messageNoEqNull);
}
