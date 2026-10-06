import type { Batch4Context } from './batch4_context.ts';
const messageConsistentTypeDefinitionsInterfaceOverType =
    'Use an `interface` instead of a `type` here. Both describe the same object shape, but an interface is the one this project writes: it can be reopened by a later declaration, it shows up under its own name in compiler errors rather than being expanded inline, and mixing the two spellings for the same job makes a reader ask whether the difference was meant.';
const messageConsistentTypeDefinitionsTypeOverInterface =
    'Use a `type` instead of an `interface` here. Both describe the same object shape, but a type alias is the one this project writes: it cannot be reopened by a declaration somewhere else in the program, so what you read at the definition is all there is, and mixing the two spellings for the same job makes a reader ask whether the difference was meant.';

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/consistent-type-definitions')) return;
    const style = ctx.settings.read('style', ctx.settings.read('option', 'interface'));
    const kind = ctx.node(index).kind;
    if(
        (style === 'interface' && kind !== 'TypeAliasDeclaration') ||
        (style === 'type' && kind !== 'InterfaceDeclaration')
    )
        return;
    const name = ctx.name(index);
    if(name < 0) return;
    let body = ctx.node(index).children[ctx.node(index).children.length - 1] ?? -1;
    if(style === 'interface') {
        while(body >= 0 && ctx.node(body).kind === 'ParenthesizedType') body = ctx.node(body).children[0] ?? -1;
        if(body < 0 || ctx.node(body).kind !== 'TypeLiteral') return;
    }
    const start = ctx.start(index);
    const end = ctx.node(index).end;
    const keyword = ctx.source.indexOf(style === 'interface' ? 'type' : 'interface', start);
    const message =
        style === 'interface'
            ? messageConsistentTypeDefinitionsInterfaceOverType
            : messageConsistentTypeDefinitionsTypeOverInterface;
    const id = style === 'interface' ? 'interfaceOverType' : 'typeOverInterface';
    let global = false;
    for(let parent = ctx.parent(index); parent >= 0; parent = ctx.parent(parent))
        if(
            ctx.node(parent).kind === 'ModuleDeclaration' &&
            ctx.hasKind(parent, 'DeclareKeyword') &&
            ctx.node(parent).operator === 'GlobalKeyword'
        )
            global = true;
    const finding = ctx.report(
        name,
        '@typescript-eslint/consistent-type-definitions',
        id,
        message,
        global ? '' : 'fix',
        global ? '' : style === 'interface' ? 'interface' : 'type',
    );
    if(global) return;
    finding.editStart = keyword;
    finding.editEnd = keyword + (style === 'interface' ? 4 : 9);
    if(style === 'interface') {
        const bodyStart = ctx.start(body);
        const equals = ctx.source.slice(start, bodyStart).lastIndexOf('=') + start;
        // The pinned implementation trims whitespace only, retaining a preceding comment.
        let previous = equals;
        while(previous > start && ' \t\n\r'.includes(ctx.source[previous - 1] ?? '')) previous--;
        ctx.edit(previous, bodyStart, ' ');
        ctx.edit(ctx.node(body).end, end, '');
    }
    else {
        let headEnd = ctx.node(name).end;
        const parameters = ctx.node(index).children.filter((child) => ctx.node(child).kind === 'TypeParameter');
        if(parameters.length > 0) {
            const last = parameters[parameters.length - 1] ?? name;
            headEnd = ctx.source.indexOf('>', ctx.node(last).end) + 1;
        }
        const brace = ctx.source.indexOf('{', headEnd);
        ctx.edit(headEnd, brace, ' = ');
        const heritage = ctx.child(index, 'HeritageClause');
        if(heritage >= 0) {
            let suffix = '';
            for(const type of ctx.node(heritage).children) suffix += ` & ${ctx.text(type)}`;
            ctx.edit(end, end, suffix);
        }
        if(ctx.hasKind(index, 'ExportKeyword') && ctx.hasKind(index, 'DefaultKeyword')) {
            ctx.edit(start, keyword, '');
            ctx.edit(end, end, `\nexport default ${ctx.node(name).text}`);
        }
    }
}
