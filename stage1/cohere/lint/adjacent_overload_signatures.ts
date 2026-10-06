import type { Batch4Context } from './batch4_context.ts';

import { isIdentifierStart, isIdentifierPart } from '../../typescript/scanner/characters.ts';
function identifier(text: string): boolean {
    if(text.length === 0 || !isIdentifierStart(text.codePointAt(0) ?? -1)) return false;
    for(let index = 1; index < text.length; index++) if(!isIdentifierPart(text.codePointAt(index) ?? -1)) return false;
    return true;
}
export function visit(ctx: Batch4Context, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/adjacent-overload-signatures') ||
        ![
            'ClassDeclaration',
            'ClassExpression',
            'InterfaceDeclaration',
            'TypeLiteral',
            'SourceFile',
            'Block',
            'ModuleBlock',
        ].includes(ctx.node(index).kind)
    )
        return;
    const seen = new Set<string>();
    let previous = '';
    for(const member of ctx.node(index).children) {
        const kind = ctx.node(member).kind;
        if(
            kind.endsWith('Keyword') ||
            kind === 'TypeParameter' ||
            kind === 'HeritageClause' ||
            ((ctx.node(index).kind === 'InterfaceDeclaration' ||
                ctx.node(index).kind === 'ClassDeclaration' ||
                ctx.node(index).kind === 'ClassExpression') &&
                member === ctx.name(index) &&
                ctx.node(member).kind === 'Identifier')
        )
            continue;
        if(
            ![
                'FunctionDeclaration',
                'MethodSignature',
                'MethodDeclaration',
                'GetAccessor',
                'SetAccessor',
                'Constructor',
                'CallSignature',
                'ConstructSignature',
            ].includes(kind) ||
            ctx.hasKind(member, 'AbstractKeyword')
        ) {
            previous = '';
            continue;
        }
        let name =
            kind === 'Constructor'
                ? 'constructor'
                : kind === 'CallSignature'
                  ? 'call'
                  : kind === 'ConstructSignature'
                    ? 'new'
                    : '';
        let nameKind = 'Normal';
        if(name === '') {
            let key = ctx.name(member);
            if(key < 0) {
                previous = '';
                continue;
            }
            if(ctx.node(key).kind === 'ComputedPropertyName') key = ctx.node(key).children[0] ?? key;
            const node = ctx.node(key);
            if(node.kind === 'PrivateIdentifier') {
                name = node.text;
                nameKind = 'Private';
            }
            else if(node.kind === 'Identifier') name = node.text;
            else if(node.kind === 'StringLiteral' || node.kind === 'NumericLiteral') {
                name = identifier(node.text) ? node.text : `"${node.text}"`;
                nameKind = identifier(node.text) ? 'Normal' : 'Quoted';
            }
            else {
                name = ctx.text(key);
                nameKind = 'Expression';
            }
        }
        const staticness = ['FunctionDeclaration', 'CallSignature', 'ConstructSignature'].includes(kind)
            ? 'Unset'
            : ctx.hasKind(member, 'StaticKeyword')
              ? 'Static'
              : 'Instance';
        const key = `${nameKind} ${staticness} ${kind === 'CallSignature'} ${name}`;
        if(seen.has(key) && previous !== key) {
            const display = staticness === 'Static' ? `static ${name}` : name;
            ctx.report(
                member,
                '@typescript-eslint/adjacent-overload-signatures',
                'adjacentSignature',
                `This signature for ${display} is separated from the others by unrelated members, so a reader scrolling the declaration sees an overload set that looks complete and is not. Move every signature for ${display} together, in the order the compiler resolves them, so the whole set can be read at once.`,
            );
        }
        seen.add(key);
        previous = key;
    }
}
