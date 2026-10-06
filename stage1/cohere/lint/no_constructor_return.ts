// Port of pinned Go cohere's no-constructor-return.
import type { RuleContext } from './rule_context.ts';

// Normalize the two constructor shapes without changing the shared parser.
function constructor(ctx: RuleContext, index: number): boolean {
    const node = ctx.node(index);
    if(ctx.hasKind(index, 'AsteriskToken')) return false;
    if(node.kind === 'Constructor') {
        let position = node.pos;
        for(const child of node.children) {
            if(ctx.node(child).kind.endsWith('Keyword')) position = Math.max(position, ctx.node(child).end);
        }
        const token = ctx.scanAt(position);
        return token !== 'GetKeyword' && token !== 'SetKeyword';
    }
    if(node.kind !== 'MethodDeclaration') return false;
    const owner = ctx.parent(index);
    if(owner < 0 || !['ClassDeclaration', 'ClassExpression'].includes(ctx.node(owner).kind)) return false;
    const name = node.children.find((child) => !ctx.node(child).kind.endsWith('Keyword')) ?? -1;
    return name >= 0 && ctx.node(name).kind === 'StringLiteral' && ctx.node(name).text === 'constructor';
}

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-constructor-return')) return;
    if(ctx.node(index).kind !== 'ReturnStatement' || ctx.node(index).children.length === 0) return;
    for(let parent = ctx.parent(index); parent >= 0; parent = ctx.parent(parent)) {
        if(constructor(ctx, parent)) {
            if(!ctx.hasKind(parent, 'StaticKeyword'))
                ctx.report(
                    index,
                    'no-constructor-return',
                    'noConstructorReturn',
                    `This returns a value from a constructor, where it is either discarded or worse. \`new C()\` evaluates to the new instance and throws the returned value away, unless that value is an object, in which case the language substitutes it for the instance and every field this constructor assigned is lost. Neither outcome is one a reader expects from a \`new\`, so return nothing and assign to \`this\` instead.`,
                );
            return;
        }
        if(ctx.functionLike(parent)) return;
    }
}
