import { panic } from 'adamic';
import type { Batch4Context } from './batch4_context.ts';
const messageUseTopLevelQualifier =
    "Every name in this import carries its own inline `type` qualifier, so TypeScript erases the names and leaves the bare `import 'module'` behind, which still runs the module for its side effects. The import reads as type-only and is not. Move the qualifier to the top level, where it removes the whole statement.";

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-import-type-side-effects') || ctx.node(index).kind !== 'ImportDeclaration')
        return;
    const clause = ctx.child(index, 'ImportClause');
    if(clause < 0 || ctx.node(clause).semantic === 'TypeKeyword' || ctx.child(clause, 'Identifier') >= 0) return;
    const named = ctx.child(clause, 'NamedImports');
    if(
        named < 0 ||
        ctx.node(named).children.length === 0 ||
        !ctx.node(named).children.every((child) => ctx.node(child).semantic === '1')
    )
        return;
    const first = ctx.node(named).children[0] ?? panic('specifier');
    const imported = ctx.node(first).children[0] ?? panic('imported');
    const finding = ctx.report(
        index,
        '@typescript-eslint/no-import-type-side-effects',
        'useTopLevelQualifier',
        messageUseTopLevelQualifier,
        'fix',
    );
    finding.editStart = ctx.start(first);
    finding.editEnd = ctx.start(imported);
    for(const specifier of ctx.node(named).children.slice(1)) {
        const name = ctx.node(specifier).children[0] ?? panic('imported');
        ctx.edit(ctx.start(specifier), ctx.start(name), '');
    }
    ctx.edit(ctx.start(index) + 6, ctx.start(index) + 6, ' type');
}
