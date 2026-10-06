import type { Batch4Context } from './batch4_context.ts';

function call(ctx: Batch4Context, index: number, name: string, construct = false): boolean {
    const node = ctx.node(index);
    const target = node.children[0] ?? -1;
    return (
        node.kind === (construct ? 'NewExpression' : 'CallExpression') &&
        target >= 0 &&
        ctx.node(target).kind === 'Identifier' &&
        ctx.node(target).text === name
    );
}
function inferred(ctx: Batch4Context, type: number, value: number): string {
    const kind = ctx.node(type).kind;
    const node = ctx.node(value);
    let unwrapped = value;
    if(node.kind === 'PrefixUnaryExpression' && ['PlusToken', 'MinusToken'].includes(node.operator))
        unwrapped = node.children[0] ?? value;
    const inner = ctx.node(unwrapped);
    switch(kind) {
        case 'NumberKeyword':
            return inner.kind === 'NumericLiteral' ||
                (inner.kind === 'Identifier' && ['NaN', 'Infinity'].includes(inner.text)) ||
                call(ctx, unwrapped, 'Number')
                ? 'number'
                : '';
        case 'BigIntKeyword': {
            const big =
                node.kind === 'PrefixUnaryExpression' && node.operator === 'MinusToken'
                    ? ctx.node(node.children[0] ?? value)
                    : node;
            return [
                'NumericLiteral',
                'BigIntLiteral',
                'StringLiteral',
                'RegularExpressionLiteral',
                'TrueKeyword',
                'FalseKeyword',
                'NullKeyword',
            ].includes(big.kind) ||
                call(
                    ctx,
                    node.kind === 'PrefixUnaryExpression' && node.operator === 'MinusToken'
                        ? (node.children[0] ?? value)
                        : value,
                    'BigInt',
                )
                ? 'bigint'
                : '';
        }
        case 'BooleanKeyword':
            return ['TrueKeyword', 'FalseKeyword'].includes(node.kind) ||
                (node.kind === 'PrefixUnaryExpression' && node.operator === 'ExclamationToken') ||
                call(ctx, value, 'Boolean')
                ? 'boolean'
                : '';
        case 'StringKeyword':
            return ['StringLiteral', 'NoSubstitutionTemplateLiteral', 'TemplateExpression'].includes(node.kind) ||
                call(ctx, value, 'String')
                ? 'string'
                : '';
        case 'SymbolKeyword':
            return call(ctx, value, 'Symbol') ? 'symbol' : '';
        case 'UndefinedKeyword':
            return node.kind === 'VoidExpression' || (node.kind === 'Identifier' && node.text === 'undefined')
                ? 'undefined'
                : '';
        case 'LiteralType':
            return ctx.child(type, 'NullKeyword') >= 0 && node.kind === 'NullKeyword' ? 'null' : '';
        case 'TypeReference': {
            const name = ctx.node(type).children[0] ?? -1;
            return name >= 0 &&
                ctx.node(name).kind === 'Identifier' &&
                ctx.node(name).text === 'RegExp' &&
                (node.kind === 'RegularExpressionLiteral' ||
                    call(ctx, value, 'RegExp') ||
                    call(ctx, value, 'RegExp', true))
                ? 'RegExp'
                : '';
        }
        default:
            return '';
    }
}
export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-inferrable-types')) return;
    const node = ctx.node(index);
    if(!['VariableDeclaration', 'PropertyDeclaration', 'Parameter'].includes(node.kind)) return;
    if(node.kind === 'Parameter') {
        const parent = ctx.parent(index);
        if(parent < 0 || !ctx.functionLike(parent) || ctx.settings.read('ignoreparameters', 'false') === 'true') return;
    }
    if(
        node.kind === 'PropertyDeclaration' &&
        (ctx.settings.read('ignoreproperties', 'false') === 'true' ||
            ctx.hasKind(index, 'ReadonlyKeyword') ||
            ctx.hasKind(index, 'QuestionToken'))
    )
        return;
    const parts = ctx.parts(index);
    const name = parts[0] ?? -1;
    const type = parts[1] ?? -1;
    const value = parts[2] ?? -1;
    if(type < 0 || value < 0) return;
    const text = inferred(ctx, type, value);
    if(text === '') return;
    const parameterProperty =
        node.kind === 'Parameter' &&
        node.children.some((child) =>
            ['PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'OverrideKeyword'].includes(
                ctx.node(child).kind,
            ),
        );
    const token = ctx.child(index, node.kind === 'Parameter' ? 'QuestionToken' : 'ExclamationToken');
    const finding = ctx.report(
        parameterProperty ? name : index,
        '@typescript-eslint/no-inferrable-types',
        'noInferrableType',
        `Type \`${text}\` is trivially inferred from a \`${text}\` literal, so the annotation restates what the initializer already says. Remove it and the declaration means exactly the same thing with one fewer place to keep in sync.`,
        'fix',
    );
    let colon = ctx.start(type) - 1;
    while(colon >= 0 && ' \t\n\r\v\f'.includes(ctx.source[colon] ?? '')) colon--;
    if(colon < 0 || ctx.source[colon] !== ':') colon = ctx.start(type);
    if(token >= 0) {
        finding.editStart = ctx.start(token);
        finding.editEnd = ctx.node(token).end;
        ctx.edit(colon, ctx.node(type).end, '');
    }
    else {
        finding.editStart = colon;
        finding.editEnd = ctx.node(type).end;
    }
}
