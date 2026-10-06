import type { Batch4Context } from './batch4_context.ts';
const messageNoDuplicateEnumValues =
    'Two members of this enum are initialized to the same value. TypeScript permits it, but a reader expects members of one enum to name distinct things, and a reverse lookup by value can only return one of them, so the later member silently wins. The usual cause is a copied line whose value was never changed. Give each member its own value.';

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-duplicate-enum-values') || ctx.node(index).kind !== 'EnumDeclaration')
        return;
    const numbers = new Map<string, number>();
    const strings = new Map<string, number>();
    for(const member of ctx.node(index).children) {
        if(ctx.node(member).kind !== 'EnumMember') continue;
        const value = ctx.node(member).children[1] ?? -1;
        if(value < 0) continue;
        const node = ctx.node(value);
        if(node.kind !== 'StringLiteral' && node.kind !== 'NumericLiteral') continue;
        const seen = node.kind === 'StringLiteral' ? strings : numbers;
        const earlier = seen.get(node.text);
        if(earlier !== undefined)
            ctx.report(
                earlier,
                '@typescript-eslint/no-duplicate-enum-values',
                'noDuplicateEnumValues',
                messageNoDuplicateEnumValues,
            );
        if(node.kind === 'StringLiteral' || earlier === undefined) seen.set(node.text, value);
    }
}
