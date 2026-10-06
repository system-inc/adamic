import type { Batch4Context } from './batch4_context.ts';
const messageUselessEmptyExport =
    'An `export {}` exists to make a script into a module, and this file is already a module because something else in it imports or exports. The statement therefore changes nothing, and it reads as though it were load-bearing to anyone deciding whether they may delete it. Remove it.';

export function visit(ctx: Batch4Context, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-useless-empty-export') ||
        ctx.node(index).kind !== 'SourceFile' ||
        ctx.parser.path.endsWith('.d.ts')
    )
        return;
    const empty: number[] = [];
    let module = false;
    for(const statement of ctx.node(index).children) {
        const node = ctx.node(statement);
        const named = ctx.child(statement, 'NamedExports');
        if(
            node.kind === 'ExportDeclaration' &&
            named >= 0 &&
            ctx.node(named).children.length === 0 &&
            ctx.child(statement, 'StringLiteral') < 0
        ) {
            empty.push(statement);
            continue;
        }
        if(
            node.kind === 'ImportDeclaration' ||
            node.kind === 'ExportDeclaration' ||
            (node.kind === 'ExportAssignment' && node.semantic !== '1') ||
            (node.kind === 'ImportEqualsDeclaration' && ctx.child(statement, 'ExternalModuleReference') >= 0) ||
            ctx.hasKind(statement, 'ExportKeyword')
        )
            module = true;
    }
    if(module)
        for(const statement of empty)
            ctx.report(
                statement,
                '@typescript-eslint/no-useless-empty-export',
                'uselessEmptyExport',
                messageUselessEmptyExport,
                'fix',
            );
}
