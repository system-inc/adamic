// Port of pinned Go cohere's no-useless-concat.
import type { RuleContext } from './rule_context.ts';

const messageUnexpectedConcat =
    'These two literals are joined at runtime where the source could have written one literal. Nothing is computed between them, so the concatenation only adds an operator for a reader to follow and a chance for the two halves to drift apart.';

function concat(ctx: RuleContext, index: number): boolean {
    return (
        index >= 0 &&
        ctx.node(index).kind === 'BinaryExpression' &&
        ctx.node(ctx.node(index).children[1] ?? -1).kind === 'PlusToken'
    );
}
function literal(ctx: RuleContext, index: number): boolean {
    return (
        index >= 0 &&
        ['StringLiteral', 'NoSubstitutionTemplateLiteral', 'TemplateExpression'].includes(ctx.node(index).kind)
    );
}

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-useless-concat')) return;
    if(!concat(ctx, index)) return;
    let left = ctx.unwrap(ctx.node(index).children[0] ?? -1);
    let right = ctx.unwrap(ctx.node(index).children[2] ?? -1);
    while(concat(ctx, left)) left = ctx.unwrap(ctx.node(left).children[2] ?? -1);
    while(concat(ctx, right)) right = ctx.unwrap(ctx.node(right).children[0] ?? -1);
    if(!literal(ctx, left) || !literal(ctx, right)) return;
    const gap = ctx.source.slice(ctx.node(left).end, ctx.start(right));
    if(gap.includes('\n') || gap.includes('\r')) return;
    ctx.report(ctx.node(index).children[1] ?? -1, 'no-useless-concat', 'unexpectedConcat', messageUnexpectedConcat);
}
