import type { RuleContext } from './rule_context.ts';
const message =
    'Declaring a variable named `module` shadows the CommonJS `module` object that the bundler relies on, so code compiled into a CommonJS context loses its reference to its own exports. The failure surfaces at runtime as an export that is silently missing rather than as a build error. Rename the variable.';
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@next/next/no-assign-module-variable') || ctx.node(index).kind !== 'VariableStatement') return;
    const list = ctx.child(index, 'VariableDeclarationList');
    if(list < 0) return;
    for(const declaration of ctx.node(list).children) {
        const name = ctx.name(declaration);
        if(name >= 0 && ctx.node(name).kind === 'Identifier' && ctx.node(name).text === 'module') {
            ctx.report(index, '@next/next/no-assign-module-variable', 'noAssignModuleVariable', message);
            return;
        }
    }
}
