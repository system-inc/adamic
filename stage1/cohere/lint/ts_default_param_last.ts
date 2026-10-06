// Port of pinned Go cohere's @typescript-eslint/default-param-last.
import type { RuleContext } from './rule_context.ts';
import { text, defaultParameters } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/default-param-last')) return;
    const rule = '@typescript-eslint/default-param-last';
    defaultParameters(ctx, index, rule, text('ts_default_param_last'), false);
}
