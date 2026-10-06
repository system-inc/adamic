import type { Batch4Context } from './batch4_context.ts';

export function visit(ctx: Batch4Context, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-dupe-class-members') ||
        !['ClassDeclaration', 'ClassExpression'].includes(ctx.node(index).kind)
    )
        return;
    const seen = new Map<string, string>();
    for(const member of ctx.node(index).children) {
        const kind = ctx.node(member).kind;
        if(!['PropertyDeclaration', 'MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(kind)) continue;
        if(kind !== 'PropertyDeclaration' && ctx.body(member) < 0) continue;
        const name = ctx.name(member);
        if(name < 0 || ctx.node(name).kind === 'ComputedPropertyName') continue;
        const text = `${ctx.node(name).kind === 'NumericLiteral' ? 'number' : 'string'}:${ctx.node(name).text}`;
        const key = `${ctx.hasKind(member, 'StaticKeyword')} ${ctx.node(name).kind === 'PrivateIdentifier'} ${text}`;
        const earlier = seen.get(key);
        if(earlier === undefined) {
            seen.set(key, kind);
            continue;
        }
        if(
            ['GetAccessor', 'SetAccessor'].includes(kind) &&
            ['GetAccessor', 'SetAccessor'].includes(earlier) &&
            earlier !== kind
        )
            continue;
        ctx.report(
            name,
            '@typescript-eslint/no-dupe-class-members',
            'noDupeClassMembers',
            `This class already declares a member named ${text}. The later declaration wins silently, so the earlier one is dead code that reads as live, and nothing in the language or at runtime tells the two apart.`,
        );
    }
}
