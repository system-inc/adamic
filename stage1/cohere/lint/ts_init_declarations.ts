// Port of pinned Go cohere's @typescript-eslint/init-declarations.
import type { RuleContext } from './rule_context.ts';
import { child, kind } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/init-declarations')) return;
    const rule = '@typescript-eslint/init-declarations';
    if(kind(ctx, index) !== 'VariableDeclarationList') return;
    const mode = ctx.settings.read('mode', '');
    if(mode !== 'always' && mode !== 'never') return;
    let ancestor = index;
    while(ancestor >= 0) {
        if(ctx.hasKind(ancestor, 'DeclareKeyword')) return;
        ancestor = ctx.parent(ancestor);
    }
    const parent = ctx.parent(index);
    const loop = parent >= 0 && ['ForStatement', 'ForInStatement', 'ForOfStatement'].includes(kind(ctx, parent));
    for(const declaration of ctx.node(index).children) {
        const name = child(ctx, declaration, 0);
        if(kind(ctx, name) !== 'Identifier') continue;
        const initialized = loop || ctx.initializer(declaration) >= 0;
        if(mode === 'always' && !initialized)
            ctx.report(
                name,
                rule,
                'initialized',
                `Variable '${ctx.node(name).text}' should be initialized on declaration.`,
            );
        if(
            mode === 'never' &&
            initialized &&
            !['2', '4', '6'].includes(ctx.node(index).semantic) &&
            !(loop && ctx.settings.read('ignoreforloopinit', 'false') === 'true')
        )
            ctx.report(
                declaration,
                rule,
                'notInitialized',
                `Variable '${ctx.node(name).text}' should not be initialized on declaration.`,
            );
    }
}
