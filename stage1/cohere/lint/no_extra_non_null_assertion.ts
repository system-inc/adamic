import type { RuleContext } from './rule_context.ts';
const messageNoExtraNonNullAssertion =
    'This non-null assertion is redundant: the value it asserts on has already been asserted non-null, or is about to be checked at runtime by the `?.` that follows it. Either way the second `!` narrows nothing the first one did not, so it is noise that reads as a claim. Remove it.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-extra-non-null-assertion') || ctx.node(index).kind !== 'NonNullExpression')
        return;
    let outer = index;
    while(ctx.parent(outer) >= 0 && ctx.node(ctx.parent(outer)).kind === 'ParenthesizedExpression')
        outer = ctx.parent(outer);
    const parent = ctx.parent(outer);
    if(parent < 0) return;
    const node = ctx.node(parent);
    if(
        node.kind !== 'NonNullExpression' &&
        !(
            ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(node.kind) &&
            node.children[0] === outer &&
            ctx.hasKind(parent, 'QuestionDotToken')
        )
    )
        return;
    const finding = ctx.report(
        index,
        '@typescript-eslint/no-extra-non-null-assertion',
        'noExtraNonNullAssertion',
        messageNoExtraNonNullAssertion,
        'fix',
    );
    finding.editStart = ctx.node(index).end - 1;
}
