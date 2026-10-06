// Operator precedence and argument-shape classification, independent of document layout.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { stringWidth } from './width.ts';
export const assignmentOperators: readonly string[] = [
    '=',
    '+=',
    '-=',
    '*=',
    '/=',
    '%=',
    '**=',
    '<<=',
    '>>=',
    '>>>=',
    '&=',
    '|=',
    '^=',
    '&&=',
    '||=',
    '??=',
];
const operators = new Map<string, string>([
    ['CommaToken', ','],
    ['EqualsToken', '='],
    ['PlusEqualsToken', '+='],
    ['MinusEqualsToken', '-='],
    ['AsteriskEqualsToken', '*='],
    ['SlashEqualsToken', '/='],
    ['PercentEqualsToken', '%='],
    ['AsteriskAsteriskEqualsToken', '**='],
    ['LessThanLessThanEqualsToken', '<<='],
    ['GreaterThanGreaterThanEqualsToken', '>>='],
    ['GreaterThanGreaterThanGreaterThanEqualsToken', '>>>='],
    ['AmpersandEqualsToken', '&='],
    ['BarEqualsToken', '|='],
    ['CaretEqualsToken', '^='],
    ['AmpersandAmpersandEqualsToken', '&&='],
    ['BarBarEqualsToken', '||='],
    ['QuestionQuestionEqualsToken', '??='],
    ['PlusToken', '+'],
    ['MinusToken', '-'],
    ['AsteriskToken', '*'],
    ['SlashToken', '/'],
    ['PercentToken', '%'],
    ['AsteriskAsteriskToken', '**'],
    ['LessThanToken', '<'],
    ['GreaterThanToken', '>'],
    ['LessThanEqualsToken', '<='],
    ['GreaterThanEqualsToken', '>='],
    ['EqualsEqualsToken', '=='],
    ['ExclamationEqualsToken', '!='],
    ['EqualsEqualsEqualsToken', '==='],
    ['ExclamationEqualsEqualsToken', '!=='],
    ['BarToken', '|'],
    ['CaretToken', '^'],
    ['AmpersandToken', '&'],
    ['BarBarToken', '||'],
    ['AmpersandAmpersandToken', '&&'],
    ['QuestionQuestionToken', '??'],
    ['LessThanLessThanToken', '<<'],
    ['GreaterThanGreaterThanToken', '>>'],
    ['GreaterThanGreaterThanGreaterThanToken', '>>>'],
    ['InKeyword', 'in'],
    ['InstanceOfKeyword', 'instanceof'],
    ['ExclamationToken', '!'],
    ['TildeToken', '~'],
    ['PlusPlusToken', '++'],
    ['MinusMinusToken', '--'],
    ['DeleteExpression', 'delete'],
    ['VoidExpression', 'void'],
    ['TypeOfExpression', 'typeof'],
]);
export function rank(operator: string): number {
    const levels: readonly (readonly string[])[] = [
        ['??'],
        ['||'],
        ['&&'],
        ['|'],
        ['^'],
        ['&'],
        ['==', '!=', '===', '!=='],
        ['<', '>', '<=', '>=', 'in', 'instanceof'],
        ['<<', '>>', '>>>'],
        ['+', '-'],
        ['*', '/', '%'],
        ['**'],
    ];
    for(let index = 0; index < levels.length; index++)
        if((levels[index] ?? panic('missing precedence level')).includes(operator)) return index;
    return -1;
}
export function flatten(parent: string, child: string): boolean {
    if(rank(parent) !== rank(child) || parent === '**') return false;
    if(['==', '!=', '===', '!=='].includes(parent) && ['==', '!=', '===', '!=='].includes(child)) return false;
    if(['*', '/', '%'].includes(parent) && ['*', '/', '%'].includes(child) && (parent !== child || parent === '%'))
        return false;
    if(['<<', '>>', '>>>'].includes(parent) && ['<<', '>>', '>>>'].includes(child)) return false;
    return true;
}
function syntaxNode(parser: Parser, index: number): ParseNode {
    return parser.nodes[index] ?? panic('missing syntax node');
}
function unwrapped(parser: Parser, index: number): number {
    const item = syntaxNode(parser, index);
    return item.kind === 'ParenthesizedExpression'
        ? unwrapped(parser, item.children[0] ?? panic('missing parenthesized syntax'))
        : index;
}
function syntaxChild(parser: Parser, index: number, offset: number): number {
    return unwrapped(parser, syntaxNode(parser, index).children[offset] ?? panic('missing syntax child'));
}
export function syntaxOperator(parser: Parser, index: number): string {
    const item = syntaxNode(parser, index);
    const kind =
        item.kind === 'BinaryExpression'
            ? syntaxNode(parser, item.children[1] ?? panic('missing syntax operator')).kind
            : item.operator === ''
              ? item.kind
              : item.operator;
    return operators.get(kind) ?? '';
}
export function simpleArgument(parser: Parser, source: string, index: number, depth = 2): boolean {
    if(depth <= 0) return false;
    const node = syntaxNode(parser, index);
    if(node.kind === 'NonNullExpression') return simpleArgument(parser, source, syntaxChild(parser, index, 0), depth);
    if(node.kind === 'RegularExpressionLiteral') {
        const raw = source.slice(node.pos, node.end).trim();
        return stringWidth(raw.slice(1, raw.lastIndexOf('/'))) <= 5;
    }
    if(
        [
            'Identifier',
            'PrivateIdentifier',
            'ThisKeyword',
            'SuperKeyword',
            'NumericLiteral',
            'BigIntLiteral',
            'StringLiteral',
            'NullKeyword',
            'TrueKeyword',
            'FalseKeyword',
        ].includes(node.kind)
    )
        return true;
    if(node.kind === 'NoSubstitutionTemplateLiteral') return !source.slice(node.pos, node.end).includes('\n');
    if(node.kind === 'TemplateExpression') {
        if(syntaxNode(parser, node.children[0] ?? panic('missing template head')).raw.includes('\n')) return false;
        for(let position = 1; position < node.children.length; position++) {
            const span = syntaxNode(parser, node.children[position] ?? panic('missing template span'));
            if(
                syntaxNode(parser, span.children[1] ?? panic('missing quasi')).raw.includes('\n') ||
                !simpleArgument(
                    parser,
                    source,
                    unwrapped(parser, span.children[0] ?? panic('missing expression')),
                    depth - 1,
                )
            )
                return false;
        }
        return true;
    }
    if(node.kind === 'ObjectLiteralExpression') {
        for(const child of node.children) {
            const property = syntaxNode(parser, child);
            if(property.kind === 'ShorthandPropertyAssignment') continue;
            if(
                property.kind !== 'PropertyAssignment' ||
                syntaxNode(parser, syntaxChild(parser, child, 0)).kind === 'ComputedPropertyName' ||
                !simpleArgument(parser, source, syntaxChild(parser, child, 1), depth - 1)
            )
                return false;
        }
        return true;
    }
    if(node.kind === 'ArrayLiteralExpression') {
        for(const child of node.children)
            if(
                syntaxNode(parser, child).kind !== 'OmittedExpression' &&
                !simpleArgument(parser, source, unwrapped(parser, child), depth - 1)
            )
                return false;
        return true;
    }
    if(['CallExpression', 'NewExpression'].includes(node.kind)) {
        if(!simpleArgument(parser, source, syntaxChild(parser, index, 0), depth) || node.list > depth) return false;
        for(let position = node.children.length - Math.max(node.list, 0); position < node.children.length; position++)
            if(!simpleArgument(parser, source, syntaxChild(parser, index, position), depth - 1)) return false;
        return true;
    }
    if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind))
        return (
            simpleArgument(parser, source, syntaxChild(parser, index, 0), depth) &&
            simpleArgument(parser, source, syntaxChild(parser, index, node.children.length - 1), depth)
        );
    if(
        node.kind === 'PostfixUnaryExpression' ||
        (node.kind === 'PrefixUnaryExpression' &&
            ['!', '-', '+', '~', '++', '--'].includes(syntaxOperator(parser, index)))
    )
        return simpleArgument(parser, source, syntaxChild(parser, index, 0), depth);
    return false;
}
export function factory(name: string): boolean {
    if(name.charCodeAt(0) >= 65 && name.charCodeAt(0) <= 90) return true;
    if(name === '') return false;
    for(let index = 0; index < name.length; index++)
        if(!['_', '$'].includes(name.slice(index, index + 1))) return false;
    return true;
}
export function couldExpandArgument(parser: Parser, index: number, chain = false): boolean {
    const node = syntaxNode(parser, index);
    if(['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(node.kind)) return node.children.length > 0;
    if(node.kind === 'FunctionExpression') return true;
    if(node.kind !== 'ArrowFunction') return false;
    const body = unwrapped(parser, node.children[node.children.length - 1] ?? panic('missing arrow body'));
    const kind = syntaxNode(parser, body).kind;
    return (
        ['Block', 'ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(kind) ||
        (kind === 'ArrowFunction' && couldExpandArgument(parser, body, true)) ||
        (!chain && ['ConditionalExpression', 'CallExpression'].includes(kind))
    );
}
export function functionArgument(parser: Parser, index: number): boolean {
    return ['ArrowFunction', 'FunctionExpression'].includes(syntaxNode(parser, index).kind);
}
export function functionComposition(parser: Parser, index: number): boolean {
    const node = syntaxNode(parser, index);
    if(node.list <= 1) return false;
    let count = 0;
    for(let position = node.children.length - node.list; position < node.children.length; position++) {
        const child = syntaxChild(parser, index, position);
        if(functionArgument(parser, child)) {
            count++;
            if(count > 1) return true;
        }
        else if(syntaxNode(parser, child).kind === 'CallExpression') {
            const call = syntaxNode(parser, child);
            for(let offset = call.children.length - Math.max(call.list, 0); offset < call.children.length; offset++)
                if(functionArgument(parser, syntaxChild(parser, child, offset))) return true;
        }
    }
    return false;
}
export function hookArguments(parser: Parser, index: number): boolean {
    const node = syntaxNode(parser, index);
    if(![2, 3].includes(node.list)) return false;
    const first = node.children.length - node.list;
    if(node.list === 3 && syntaxNode(parser, syntaxChild(parser, index, first)).kind !== 'Identifier') return false;
    const callback = syntaxNode(parser, syntaxChild(parser, index, node.children.length - 2));
    if(
        callback.kind !== 'ArrowFunction' ||
        syntaxNode(parser, callback.children[callback.children.length - 1] ?? panic('missing callback body')).kind !==
            'Block'
    )
        return false;
    for(const child of callback.children) if(syntaxNode(parser, child).kind === 'Parameter') return false;
    return syntaxNode(parser, syntaxChild(parser, index, node.children.length - 1)).kind === 'ArrayLiteralExpression';
}
export function numericArray(parser: Parser, index: number): boolean {
    const node = syntaxNode(parser, index);
    if(node.kind !== 'ArrayLiteralExpression' || node.children.length === 0) return false;
    for(const child of node.children) {
        const item = syntaxNode(parser, unwrapped(parser, child));
        if(
            item.kind !== 'NumericLiteral' &&
            !(
                item.kind === 'PrefixUnaryExpression' &&
                ['+', '-'].includes(syntaxOperator(parser, child)) &&
                syntaxNode(parser, syntaxChild(parser, child, 0)).kind === 'NumericLiteral'
            )
        )
            return false;
    }
    return true;
}

export function hasBlankLine(source: string): boolean {
    let previous = source.indexOf('\n');
    while(previous >= 0) {
        const next = source.indexOf('\n', previous + 1);
        if(next < 0) return false;
        if(source.slice(previous + 1, next).trim() === '') return true;
        previous = next;
    }
    return false;
}
