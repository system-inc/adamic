import type { Batch4Context } from './batch4_context.ts';
const messageNoThisAlias =
    'This binding holds `this` under another name. The alias exists to carry the receiver into a nested `function`, which rebinds `this` to something else, and an arrow function closes over `this` directly and needs no carrier. Every later reader now has to establish which of the two names is the real receiver and whether they are still the same object. Use an arrow function and write `this`.';
const messageNoThisDestructure =
    'This pattern pulls members off `this` into local bindings. Each one is read once, at the moment of destructuring, so a later write to the property is invisible here, and a method taken this way has lost its receiver and throws when called. Read through `this` at the point of use instead.';

export function visit(ctx: Batch4Context, index: number): void {
    if(!['.ts', '.tsx', '.mts', '.cts'].some((suffix) => ctx.parser.path.endsWith(suffix))) return;
    if(!ctx.enabled('@typescript-eslint/no-this-alias')) return;
    const node = ctx.node(index);
    let name = -1;
    let value = -1;
    if(node.kind === 'VariableDeclaration') {
        const parts = ctx.parts(index);
        name = parts[0] ?? -1;
        value = parts[2] ?? -1;
    }
    else if(ctx.assignment(index)) {
        name = node.children[0] ?? -1;
        value = node.children[2] ?? -1;
        if(name >= 0 && !['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(ctx.node(name).kind))
            name = ctx.outer(name);
    }
    if(name < 0 || value < 0 || ctx.node(value).kind !== 'ThisKeyword') return;
    const target = ctx.node(name);
    if(target.kind === 'Identifier') {
        if(ctx.settings.list('allowednames', ctx.settings.list('allownames', [])).includes(target.text)) return;
        ctx.report(name, '@typescript-eslint/no-this-alias', 'thisAssignment', messageNoThisAlias);
    }
    else if(
        ['ArrayBindingPattern', 'ObjectBindingPattern', 'ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(
            target.kind,
        ) &&
        ctx.settings.read('reportdestructuring', 'false') === 'true'
    ) {
        ctx.report(name, '@typescript-eslint/no-this-alias', 'thisDestructure', messageNoThisDestructure);
    }
}
