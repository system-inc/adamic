// Port of pinned Go cohere's base/boundary-no-global-container.
import type { RuleContext } from './rule_context.ts';
import { text, named } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('base/boundary-no-global-container')) return;
    const rule = 'base/boundary-no-global-container';
    if(named(ctx, index, 'getGlobalContainer'))
        ctx.report(
            index,
            rule,
            'boundaryNoGlobalContainer',
            text('base_boundary_no_global_container:boundaryNoGlobalContainer'),
        );
}
