import type { Batch4Context } from './batch4_context.ts';
const messageNoNonNullOptionalChain =
    'An optional chain returns `undefined` by design when a link is nullish, so asserting the result non-null contradicts the reason the `?.` was written. The assertion is erased at compile time and the `undefined` still arrives at runtime, which turns a guard the author asked for into a crash the types promised could not happen. Handle the `undefined` case instead, or drop the `?.` if the value cannot actually be nullish.';
const messageSuggestRemovingNonNull = 'Remove the non-null assertion.';

function chain(ctx: Batch4Context, index: number): boolean {
    if(index < 0) return false;
    const node = ctx.node(index);
    if(
        !['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(
            node.kind,
        )
    )
        return false;
    return ctx.hasKind(index, 'QuestionDotToken') || chain(ctx, node.children[0] ?? -1);
}
export function visit(ctx: Batch4Context, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-non-null-asserted-optional-chain') ||
        ctx.node(index).kind !== 'NonNullExpression'
    )
        return;
    const parent = ctx.parent(index);
    // A continued unparenthesized chain retains its own short circuit.
    if(
        parent >= 0 &&
        ctx.node(ctx.node(index).children[0] ?? index).kind !== 'ParenthesizedExpression' &&
        ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(
            ctx.node(parent).kind,
        ) &&
        ctx.node(parent).children[0] === index
    )
        return;
    const inner = ctx.unwrap(ctx.node(index).children[0] ?? -1);
    if(!chain(ctx, inner)) return;
    const finding = ctx.report(
        index,
        '@typescript-eslint/no-non-null-asserted-optional-chain',
        'noNonNullOptionalChain',
        messageNoNonNullOptionalChain,
        'suggestion',
        '',
        messageSuggestRemovingNonNull,
    );
    finding.editStart = ctx.node(index).end - 1;
}
