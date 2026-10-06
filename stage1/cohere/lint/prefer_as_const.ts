import type { RuleContext } from './rule_context.ts';
const messagePreferAsConst =
    'This literal type restates a value the compiler can already read off the literal itself. `as const` infers it, so the value and its type cannot drift apart; writing the type out means every edit to the value has to be made twice, and the compiler accepts the version where only one of them was.';

function matching(ctx: RuleContext, type: number, value: number): boolean {
    if(type < 0 || value < 0 || ctx.node(type).kind !== 'LiteralType') return false;
    const literal = ctx.node(type).children[0] ?? -1;
    return (
        literal >= 0 &&
        ['StringLiteral', 'NumericLiteral'].includes(ctx.node(literal).kind) &&
        ctx.node(literal).kind === ctx.node(value).kind &&
        ctx.node(literal).text === ctx.node(value).text
    );
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/prefer-as-const')) return;
    const node = ctx.node(index);
    if(node.kind === 'AsExpression') {
        const value = node.children[0] ?? -1;
        const type = node.children[1] ?? -1;
        if(matching(ctx, type, value))
            ctx.report(
                type,
                '@typescript-eslint/prefer-as-const',
                'preferAsConst',
                messagePreferAsConst,
                'fix',
                'const',
            );
        return;
    }
    if(
        !['VariableDeclaration', 'PropertyDeclaration'].includes(node.kind) ||
        ctx.hasKind(index, 'ExclamationToken') ||
        ctx.hasKind(index, 'QuestionToken')
    )
        return;
    const parts = ctx.parts(index);
    const name = parts[0] ?? -1;
    const type = parts[1] ?? -1;
    const value = parts[2] ?? -1;
    if(!matching(ctx, type, value)) return;
    const canFix = node.kind === 'PropertyDeclaration' || (name >= 0 && ctx.node(name).kind === 'Identifier');
    const finding = ctx.report(
        type,
        '@typescript-eslint/prefer-as-const',
        'preferAsConst',
        messagePreferAsConst,
        canFix ? 'fix' : '',
    );
    if(!canFix) return;
    ctx.scanAt(ctx.node(name).end);
    finding.editStart = ctx.scanStart();
    finding.editEnd = ctx.node(type).end;
    ctx.edit(ctx.node(value).end, ctx.node(value).end, ' as const');
}
