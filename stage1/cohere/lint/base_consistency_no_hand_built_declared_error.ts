// Port of pinned Go cohere's base/consistency-no-hand-built-declared-error.
import type { RuleContext } from './rule_context.ts';
import { args, text, child, kind, named, key } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('base/consistency-no-hand-built-declared-error')) return;
    const rule = 'base/consistency-no-hand-built-declared-error';
    if(kind(ctx, index) !== 'NewExpression' || !named(ctx, child(ctx, index, 0), 'BaseError')) return;
    if(ctx.parser.path.endsWith('/BaseError.ts') || ctx.parser.path.endsWith('/CreateBaseErrors.ts')) return;
    const options = args(ctx, index)[1] ?? -1;
    if(kind(ctx, options) !== 'ObjectLiteralExpression') return;
    for(const property of ctx.node(options).children)
        if(
            ['PropertyAssignment', 'ShorthandPropertyAssignment'].includes(kind(ctx, property)) &&
            named(ctx, key(ctx, property), 'identifier')
        ) {
            ctx.report(
                index,
                rule,
                'consistencyNoHandBuiltDeclaredError',
                text('base_consistency_no_hand_built_declared_error:consistencyNoHandBuiltDeclaredError'),
            );
            return;
        }
}
