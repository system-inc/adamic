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
export function syntaxNode(parser: Parser, index: number): ParseNode {
    return parser.nodes[index] ?? panic('missing syntax node');
}
export function unwrapped(parser: Parser, index: number): number {
    const item = syntaxNode(parser, index);
    return item.kind === 'ParenthesizedExpression'
        ? unwrapped(parser, item.children[0] ?? panic('missing parenthesized syntax'))
        : index;
}
export function syntaxChild(parser: Parser, index: number, offset: number): number {
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

export function syntaxAssignment(parser: Parser, index: number): boolean {
    return (
        index >= 0 &&
        syntaxNode(parser, index).kind === 'BinaryExpression' &&
        assignmentOperators.includes(syntaxOperator(parser, index))
    );
}

export function jestTag(parser: Parser, index: number): boolean {
    const node = syntaxNode(parser, unwrapped(parser, index));
    if(!['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind)) return false;
    const property = syntaxNode(parser, syntaxChild(parser, index, node.children.length - 1));
    if(property.kind !== 'Identifier' || property.text !== 'each') return false;
    let object = syntaxChild(parser, index, 0);
    const receiver = syntaxNode(parser, object);
    if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(receiver.kind)) {
        const modifier = syntaxNode(parser, syntaxChild(parser, object, receiver.children.length - 1));
        if(modifier.kind !== 'Identifier' || !['only', 'skip'].includes(modifier.text)) return false;
        object = syntaxChild(parser, object, 0);
    }
    const root = syntaxNode(parser, object);
    return (
        root.kind === 'Identifier' &&
        ['describe', 'it', 'test', 'fdescribe', 'fit', 'ftest', 'xdescribe', 'xit', 'xtest'].includes(root.text)
    );
}

export function syntaxLeftmost(parser: Parser, index: number): number {
    const node = syntaxNode(parser, index);
    if(
        [
            'BinaryExpression',
            'ConditionalExpression',
            'PropertyAccessExpression',
            'ElementAccessExpression',
            'NonNullExpression',
            'PostfixUnaryExpression',
            'CallExpression',
            'TaggedTemplateExpression',
        ].includes(node.kind)
    )
        return syntaxLeftmost(parser, syntaxChild(parser, index, 0));
    return index;
}

export function syntaxMemberish(parser: Parser, index: number): boolean {
    return ['PropertyAccessExpression', 'ElementAccessExpression'].includes(syntaxNode(parser, index).kind);
}

export function syntaxOwnOptional(parser: Parser, index: number): boolean {
    for(const child of syntaxNode(parser, index).children)
        if(syntaxNode(parser, child).kind === 'QuestionDotToken') return true;
    return false;
}

export function unsupported(parser: Parser, source: string, index: number): string {
    const id = unwrapped(parser, index);
    const node = syntaxNode(parser, id);
    switch(node.kind) {
        case 'Identifier':
        case 'PrivateIdentifier':
        case 'NumericLiteral':
        case 'BigIntLiteral':
        case 'RegularExpressionLiteral':
        case 'ThisKeyword':
        case 'SuperKeyword':
        case 'NullKeyword':
        case 'TrueKeyword':
        case 'FalseKeyword':
        case 'OmittedExpression':
            return '';
        case 'NoSubstitutionTemplateLiteral':
            return '';
        case 'StringLiteral':
            if(source.slice(node.pos, node.end).trim().includes('\n')) return 'string-literal-layout';
            return '';
        case 'ArrowFunction':
        case 'FunctionExpression':
        case 'FunctionDeclaration':
        case 'MethodDeclaration':
        case 'GetAccessor':
        case 'SetAccessor':
            return functionUnsupported(parser, source, id);
        case 'AwaitExpression':
            break;
        case 'YieldExpression':
            for(const child of node.children) {
                if(syntaxNode(parser, child).kind === 'AsteriskToken') continue;
                const reason = unsupported(parser, source, child);
                if(reason !== '') return reason;
            }
            return '';
        case 'TaggedTemplateExpression':
            if(node.children.length !== 2) return 'tag-type-arguments';
            if(jestTag(parser, syntaxChild(parser, id, 0))) return 'jest-template-table';
            break;
        case 'TemplateExpression':
            for(let position = 1; position < node.children.length; position++) {
                const span = syntaxNode(parser, node.children[position] ?? panic('missing template span'));
                const reason = unsupported(parser, source, span.children[0] ?? panic('missing template expression'));
                if(reason !== '') return reason;
            }
            return '';
        case 'PrefixUnaryExpression':
        case 'PostfixUnaryExpression':
        case 'DeleteExpression':
        case 'VoidExpression':
        case 'TypeOfExpression':
        case 'ArrayLiteralExpression':
        case 'ObjectLiteralExpression':
        case 'PropertyAssignment':
        case 'SpreadAssignment':
        case 'ComputedPropertyName':
        case 'SpreadElement':
        case 'NonNullExpression':
            break;
        case 'ShorthandPropertyAssignment':
            if(node.children.length !== 1) return 'object-pattern';
            break;
        case 'ConditionalExpression': {
            for(const offset of [0, 2, 4]) {
                const reason = unsupported(
                    parser,
                    source,
                    node.children[offset] ?? panic('missing conditional operand'),
                );
                if(reason !== '') return reason;
            }
            return '';
        }
        case 'BinaryExpression': {
            if(syntaxOperator(parser, id) === '')
                return syntaxNode(parser, node.children[1] ?? panic('missing binary token')).kind;
            if(
                syntaxAssignment(parser, id) &&
                !['Identifier', 'PropertyAccessExpression', 'ElementAccessExpression', 'NonNullExpression'].includes(
                    syntaxNode(parser, syntaxChild(parser, id, 0)).kind,
                )
            )
                return 'assignment-pattern';
            const left = unsupported(parser, source, node.children[0] ?? panic('missing left'));
            return left !== '' ? left : unsupported(parser, source, node.children[2] ?? panic('missing right'));
        }
        case 'PropertyAccessExpression':
        case 'ElementAccessExpression': {
            const left = unsupported(parser, source, node.children[0] ?? panic('missing object'));
            return left !== ''
                ? left
                : unsupported(parser, source, node.children[node.children.length - 1] ?? panic('missing property'));
        }
        case 'CallExpression':
        case 'NewExpression': {
            const calleeReason = unsupported(parser, source, node.children[0] ?? panic('missing callee'));
            if(calleeReason !== '') return calleeReason;
            const count = Math.max(node.list, 0);
            const first = node.children.length - count;
            if(
                first !== 1 &&
                !(
                    first === 2 &&
                    syntaxNode(parser, node.children[1] ?? panic('missing optional token')).kind === 'QuestionDotToken'
                )
            )
                return 'type-arguments';
            for(let position = first; position < node.children.length; position++) {
                const reason = unsupported(parser, source, node.children[position] ?? panic('missing argument'));
                if(reason !== '') return reason;
            }
            return '';
        }
        default:
            return node.kind;
    }
    for(const child of node.children) {
        const reason = unsupported(parser, source, child);
        if(reason !== '') return reason;
    }
    return '';
}

export function statementUnsupported(parser: Parser, source: string, index: number): string {
    const node = syntaxNode(parser, index);
    switch(node.kind) {
        case 'SourceFile':
            for(const child of node.children) {
                if(syntaxNode(parser, child).kind === 'EndOfFile') continue;
                const reason = statementUnsupported(parser, source, child);
                if(reason !== '') return reason;
            }
            return '';
        case 'FunctionDeclaration':
            return functionUnsupported(parser, source, index);
        case 'VariableStatement': {
            if(
                node.children.length !== 1 ||
                syntaxNode(parser, syntaxChild(parser, index, 0)).kind !== 'VariableDeclarationList'
            )
                return 'variable-modifiers';
            const list = syntaxNode(parser, syntaxChild(parser, index, 0));
            for(const child of list.children) {
                const declaration = syntaxNode(parser, child);
                const name = syntaxNode(parser, declaration.children[0] ?? panic('missing variable name'));
                if(name.kind !== 'Identifier') return 'variable-pattern';
                if(declaration.children.length > 2) return 'variable-types';
                if(declaration.children.length === 2) {
                    const initializer = declaration.children[1] ?? panic('missing variable value');
                    if(!source.slice(name.end, syntaxNode(parser, initializer).pos).includes('='))
                        return 'variable-types';
                    const reason = unsupported(parser, source, initializer);
                    if(reason !== '') return reason;
                }
            }
            return '';
        }
        case 'Block':
            for(const child of node.children) {
                const reason = statementUnsupported(parser, source, child);
                if(reason !== '') return reason;
            }
            return '';
        case 'ExpressionStatement':
        case 'ReturnStatement':
        case 'ThrowStatement':
            for(const child of node.children) {
                const reason = unsupported(parser, source, child);
                if(reason !== '') return reason;
            }
            return '';
        case 'EmptyStatement':
        case 'DebuggerStatement':
            return '';
        case 'BreakStatement':
        case 'ContinueStatement':
            return node.children.length === 0 ||
                syntaxNode(parser, node.children[0] ?? panic('missing label')).kind === 'Identifier'
                ? ''
                : 'statement-label';
        default:
            return node.kind;
    }
}

export function functionUnsupported(parser: Parser, source: string, index: number): string {
    const node = syntaxNode(parser, index);
    const last = node.children[node.children.length - 1] ?? panic('missing function child');
    const bodylessDeclaration = node.kind === 'FunctionDeclaration' && syntaxNode(parser, last).kind !== 'Block';
    const signatureLength = node.children.length - (bodylessDeclaration ? 0 : 1);
    let named = false;
    for(let position = 0; position < signatureLength; position++) {
        const child = node.children[position] ?? panic('missing function child');
        const item = syntaxNode(parser, child);
        if(
            ['MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(node.kind) &&
            !named &&
            !['AsyncKeyword', 'AsteriskToken'].includes(item.kind)
        ) {
            if(!['Identifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName'].includes(item.kind))
                return 'method-key';
            const reason = unsupported(parser, source, child);
            if(reason !== '') return reason;
            named = true;
            continue;
        }
        if(item.kind === 'Identifier') {
            if(node.kind === 'ArrowFunction' || source.slice(node.pos, item.pos).includes('(')) return 'function-types';
            named = true;
        }
        if(['AsyncKeyword', 'AsteriskToken', 'EqualsGreaterThanToken', 'Identifier'].includes(item.kind)) continue;
        if(item.kind !== 'Parameter') return 'function-types';
        let offset =
            syntaxNode(parser, item.children[0] ?? panic('missing parameter')).kind === 'DotDotDotToken' ? 1 : 0;
        const name = syntaxNode(parser, item.children[offset] ?? panic('missing parameter name'));
        if(name.kind !== 'Identifier') return 'parameter-pattern';
        offset++;
        if(item.children.length > offset) {
            const initializer = item.children[offset] ?? panic('missing initializer');
            const preceding = source.slice(name.end, syntaxNode(parser, initializer).pos);
            if(!preceding.includes('=')) return 'function-types';
            const reason = unsupported(parser, source, initializer);
            if(reason !== '') return reason;
            offset++;
        }
        if(item.children.length > offset) return 'function-types';
    }
    if(node.kind === 'FunctionExpression' && !named) return 'FunctionExpression';
    if(node.kind === 'FunctionDeclaration' && !named) return 'FunctionDeclaration';
    if(bodylessDeclaration) return '';
    const body = node.children[node.children.length - 1] ?? panic('missing function body');
    return syntaxNode(parser, body).kind === 'Block'
        ? statementUnsupported(parser, source, body)
        : unsupported(parser, source, body);
}
