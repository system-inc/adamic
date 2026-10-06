// Port of pinned Go cohere's default-case-last.
import type { RuleContext } from './rule_context.ts';
import { text, last, kind } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('default-case-last')) return;
    const rule = 'default-case-last';
    if(kind(ctx, index) !== 'SwitchStatement') return;
    const block = last(ctx, index);
    if(block < 0) return;
    const clauses = ctx.node(block).children;
    for(let i = 0; i < clauses.length; i++) {
        const clause = clauses[i] ?? -1;
        if(kind(ctx, clause) === 'DefaultClause') {
            if(i < clauses.length - 1) ctx.report(clause, rule, 'notLast', text('default_case_last'));
            return;
        }
    }
}
