// Port of pinned Go cohere's default-param-last.
import type { RuleContext } from './rule_context.ts';
import { text, defaultParameters } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('default-param-last')) return;
    const rule = 'default-param-last';
    defaultParameters(ctx, index, rule, text('default_param_last'), true);
}
