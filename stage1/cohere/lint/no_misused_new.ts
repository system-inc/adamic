import type { Batch4Context } from './batch4_context.ts';
const messageNoMisusedNewInterfaceConstruct =
    'This construct signature returns the interface that declares it, which describes a constructor whose instances are more constructors. An interface describing a static class object should have its `new` signature return the instance type being constructed, not the constructor interface itself.';
const messageNoMisusedNewInterfaceConstructor =
    'An interface cannot be constructed, so a member named `constructor` declares an ordinary method that happens to share a name with the class keyword and constructs nothing. Use a `new` construct signature if a constructor is meant, and rename the method otherwise.';
const messageNoMisusedNewClassNew =
    'This method is named `new` and declares that it returns the class it belongs to, which reads as a constructor and is not one. Calling it does not construct anything, so the name misleads every reader; a constructor is spelled `constructor`.';

export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-misused-new')) return;
    const node = ctx.node(index);
    if(node.kind === 'MethodSignature') {
        const key = ctx.name(index);
        if(key >= 0 && ctx.node(key).kind === 'Identifier' && ctx.node(key).text === 'constructor')
            ctx.report(
                key,
                '@typescript-eslint/no-misused-new',
                'interfaceConstructor',
                messageNoMisusedNewInterfaceConstructor,
            );
    }
    if(!['InterfaceDeclaration', 'ClassDeclaration', 'ClassExpression'].includes(node.kind)) return;
    const name = ctx.name(index);
    if(name < 0 || ctx.node(name).kind !== 'Identifier') return;
    for(const member of node.children) {
        const kind = ctx.node(member).kind;
        const construct = kind === 'ConstructSignature' && node.kind === 'InterfaceDeclaration';
        if(
            !construct &&
            !(
                node.kind !== 'InterfaceDeclaration' &&
                ['MethodDeclaration', 'GetAccessor'].includes(kind) &&
                ctx.body(member) < 0
            )
        )
            continue;
        const key = ctx.name(member);
        if(!construct && (key < 0 || ctx.node(key).kind !== 'Identifier' || ctx.node(key).text !== 'new')) continue;
        const type = ctx.returnType(member);
        if(type < 0 || ctx.node(type).kind !== 'TypeReference') continue;
        const reference = ctx.node(type).children[0] ?? -1;
        if(
            reference < 0 ||
            ctx.node(reference).kind !== 'Identifier' ||
            ctx.node(reference).text !== ctx.node(name).text
        )
            continue;
        if(construct) {
            const finding = ctx.report(
                member,
                '@typescript-eslint/no-misused-new',
                'interfaceConstruct',
                messageNoMisusedNewInterfaceConstruct,
            );
            ctx.replaceRange(finding, ctx.start(member), ctx.start(member) + 3);
        }
        else ctx.report(key, '@typescript-eslint/no-misused-new', 'classNew', messageNoMisusedNewClassNew);
    }
}
