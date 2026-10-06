// Port of pinned Go cohere's guard-for-in.
import type { RuleContext } from './rule_context.ts';
import { text, child, last, kind } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('guard-for-in')) return;
    const rule = 'guard-for-in';
    if(kind(ctx, index) !== 'ForInStatement') return;
    const body = last(ctx, index);
    if(body < 0 || ['EmptyStatement', 'IfStatement'].includes(kind(ctx, body))) return;
    if(kind(ctx, body) === 'Block') {
        const statements = ctx.node(body).children;
        if(statements.length === 0) return;
        const first = statements[0] ?? -1;
        if(kind(ctx, first) === 'IfStatement') {
            if(statements.length === 1) return;
            const consequent = child(ctx, first, 1);
            if(kind(ctx, consequent) === 'ContinueStatement') return;
            if(
                kind(ctx, consequent) === 'Block' &&
                ctx.node(consequent).children.length === 1 &&
                kind(ctx, child(ctx, consequent, 0)) === 'ContinueStatement'
            )
                return;
        }
    }
    ctx.report(index, rule, 'wrap', text('guard_for_in'));
}
