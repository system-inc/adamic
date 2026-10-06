// Port of pinned Go cohere's nexus/consistency-no-for-in.
import type { RuleContext } from './rule_context.ts';
import { text, kind } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('nexus/consistency-no-for-in')) return;
    const rule = 'nexus/consistency-no-for-in';
    if(kind(ctx, index) === 'ForInStatement')
        ctx.report(index, rule, 'forIn', text('nexus_consistency_no_for_in:forIn'));
}
