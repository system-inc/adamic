import type { Batch4Context } from './batch4_context.ts';

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/init-declarations') || ctx.node(index).kind !== 'VariableDeclarationList')
        return;
    const mode = ctx.settings.read('mode', '');
    if(mode !== 'always' && mode !== 'never') return;
    for(let parent = ctx.parent(index); parent >= 0; parent = ctx.parent(parent)) {
        if(ctx.hasKind(parent, 'DeclareKeyword')) return;
    }
    const parent = ctx.parent(index);
    const loop = parent >= 0 && ['ForStatement', 'ForInStatement', 'ForOfStatement'].includes(ctx.node(parent).kind);
    for(const declaration of ctx.node(index).children) {
        const parts = ctx.parts(declaration);
        const name = parts[0] ?? -1;
        if(name < 0 || ctx.node(name).kind !== 'Identifier') continue;
        const initialized = loop || (parts[2] ?? -1) >= 0;
        if(mode === 'always' && !initialized)
            ctx.report(
                name,
                '@typescript-eslint/init-declarations',
                'initialized',
                `Variable '${ctx.node(name).text}' should be initialized on declaration.`,
            );
        if(
            mode === 'never' &&
            initialized &&
            ctx.node(index).semantic !== '2' &&
            ctx.node(index).semantic !== '4' &&
            ctx.node(index).semantic !== '6' &&
            !(loop && ctx.settings.read('ignoreforloopinit', 'false') === 'true')
        )
            ctx.report(
                declaration,
                '@typescript-eslint/init-declarations',
                'notInitialized',
                `Variable '${ctx.node(name).text}' should not be initialized on declaration.`,
            );
    }
}
