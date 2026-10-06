// Port of pinned Go cohere's base/consistency-require-pagination-argument-name.
import type { RuleContext } from './rule_context.ts';
import { args, text, child, last, kind, named, key } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('base/consistency-require-pagination-argument-name')) return;
    const rule = 'base/consistency-require-pagination-argument-name';
    if(kind(ctx, index) !== 'Decorator') return;
    const call = child(ctx, index, 0);
    if(kind(ctx, call) !== 'CallExpression' || !named(ctx, child(ctx, call, 0), 'GraphQlArgument')) return;
    const parent = ctx.parent(index);
    if(kind(ctx, parent) !== 'Parameter') return;
    const name = key(ctx, parent);
    if(kind(ctx, name) !== 'Identifier') return;
    let pagination = false;
    for(const argument of args(ctx, call)) {
        if(kind(ctx, argument) === 'ArrowFunction') {
            const body = last(ctx, argument);
            if(kind(ctx, body) === 'Identifier' && ctx.node(body).text.endsWith('PaginationInput')) pagination = true;
        }
    }
    for(const part of ctx.node(parent).children)
        if(
            kind(ctx, part) === 'TypeReference' &&
            kind(ctx, child(ctx, part, 0)) === 'Identifier' &&
            ctx.node(child(ctx, part, 0)).text.endsWith('PaginationInput')
        )
            pagination = true;
    if(!pagination) return;
    let reported = name;
    const first = args(ctx, call)[0] ?? -1;
    if(['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(kind(ctx, first))) reported = first;
    const spelling = ctx.node(reported).text;
    if(spelling === 'pagination' || spelling.endsWith('Pagination')) return;
    ctx.report(reported, rule, 'invalidName', text('base_consistency_require_pagination_argument_name:invalidName'));
}
