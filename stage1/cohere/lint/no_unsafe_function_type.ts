import type { Batch4Context } from './batch4_context.ts';
const messageBannedFunctionType =
    'The `Function` type accepts any function-like value, so it constrains nothing beyond callability: it says nothing about parameters, nothing about the return type, and a value typed this way can be called with any arguments and produces `any`. That is the opposite of what a type annotation is for. Write the signature the value actually has, such as `(input: string) => number`, or `() => void` when it takes nothing and returns nothing.';

export function visit(ctx: Batch4Context, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-unsafe-function-type') ||
        !['TypeReference', 'ExpressionWithTypeArguments'].includes(ctx.node(index).kind)
    )
        return;
    const name = ctx.node(index).children[0] ?? -1;
    if(name < 0 || ctx.node(name).kind !== 'Identifier' || ctx.node(name).text !== 'Function') return;
    for(let statement = 0; statement < ctx.parser.nodes.length; statement++) {
        if(
            !['TypeAliasDeclaration', 'InterfaceDeclaration', 'ClassDeclaration', 'EnumDeclaration'].includes(
                ctx.node(statement).kind,
            )
        )
            continue;
        const local = ctx.name(statement);
        if(local >= 0 && ctx.node(local).text === 'Function') return;
    }
    ctx.report(name, '@typescript-eslint/no-unsafe-function-type', 'bannedFunctionType', messageBannedFunctionType);
}
