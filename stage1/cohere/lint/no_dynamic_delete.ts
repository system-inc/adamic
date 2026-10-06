import type { Batch4Context } from './batch4_context.ts';

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-dynamic-delete') || ctx.node(index).kind !== 'DeleteExpression') return;
    const access = ctx.unwrap(ctx.node(index).children[0] ?? -1);
    if(access < 0 || ctx.node(access).kind !== 'ElementAccessExpression' || ctx.hasKind(access, 'QuestionDotToken'))
        return;
    const key = ctx.unwrap(ctx.node(access).children[ctx.node(access).children.length - 1] ?? -1);
    if(key < 0) return;
    const node = ctx.node(key);
    if(node.kind === 'StringLiteral' || node.kind === 'NumericLiteral') return;
    if(node.kind === 'PrefixUnaryExpression' && node.operator === 'MinusToken') {
        const operand = ctx.unwrap(node.children[0] ?? -1);
        if(operand >= 0 && ctx.node(operand).kind === 'NumericLiteral') return;
    }
    ctx.report(
        key,
        '@typescript-eslint/no-dynamic-delete',
        'dynamicDelete',
        'Do not delete dynamically computed property keys.',
    );
}
