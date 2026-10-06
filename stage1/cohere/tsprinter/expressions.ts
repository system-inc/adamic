// Expression paths through cohere's print.go, print_binaryish.go, print_array.go and parentheses.go.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { ParseNode } from '../../typescript/parser/nodes.ts';
import { Documents, type SettingsOptions } from './doc.ts';
import { numberText, stringText } from './literals.ts';
export type ResultType =
    { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'NotYet'; readonly reason: string };
const operators = new Map<string, string>([
    ['CommaToken', ','],
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
function rank(operator: string): number {
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
function flatten(parent: string, child: string): boolean {
    if(rank(parent) !== rank(child) || parent === '**') return false;
    if(['==', '!=', '===', '!=='].includes(parent) && ['==', '!=', '===', '!=='].includes(child)) return false;
    if(['*', '/', '%'].includes(parent) && ['*', '/', '%'].includes(child) && (parent !== child || parent === '%'))
        return false;
    if(['<<', '>>', '>>>'].includes(parent) && ['<<', '>>', '>>>'].includes(child)) return false;
    return true;
}
function hasBlankLine(source: string): boolean {
    let previous = source.indexOf('\n');
    while(previous >= 0) {
        const next = source.indexOf('\n', previous + 1);
        if(next < 0) return false;
        if(source.slice(previous + 1, next).trim() === '') return true;
        previous = next;
    }
    return false;
}
export class Expressions {
    readonly parser: Parser;
    readonly source: string;
    readonly docs: Documents;
    directive = true;
    readonly ancestors: number[] = [];
    readonly sequenceBoundaries = new Set<number>();
    constructor(parser: Parser, source: string, docs: Documents) {
        this.parser = parser;
        this.source = source;
        this.docs = docs;
    }
    node(index: number): ParseNode {
        return this.parser.nodes[index] ?? panic('missing expression');
    }
    unwrapped(index: number): number {
        const node = this.node(index);
        return node.kind === 'ParenthesizedExpression'
            ? this.unwrapped(node.children[0] ?? panic('missing parenthesized child'))
            : index;
    }
    child(index: number, offset: number): number {
        return this.unwrapped(this.node(index).children[offset] ?? panic('missing expression child'));
    }
    operator(index: number): string {
        const node = this.node(index);
        const kind =
            node.kind === 'BinaryExpression'
                ? this.node(node.children[1] ?? panic('missing operator')).kind
                : node.operator === ''
                  ? node.kind
                  : node.operator;
        return operators.get(kind) ?? '';
    }
    hasOptional(index: number): boolean {
        const node = this.node(index);
        if(node.kind === 'QuestionDotToken') return true;
        for(const child of node.children) if(this.hasOptional(child)) return true;
        return false;
    }
    unsupported(index: number): string {
        if(this.node(index).kind === 'ParenthesizedExpression' && this.hasOptional(index))
            return 'optional-chain-parentheses';
        const id = this.unwrapped(index);
        const node = this.node(id);
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
            case 'StringLiteral':
            case 'NoSubstitutionTemplateLiteral':
                if(this.source.slice(node.pos, node.end).trim().includes('\n'))
                    return node.kind === 'StringLiteral' ? 'string-literal-layout' : 'template-literal-layout';
                return '';
            case 'PrefixUnaryExpression':
            case 'PostfixUnaryExpression':
            case 'DeleteExpression':
            case 'VoidExpression':
            case 'TypeOfExpression':
            case 'ArrayLiteralExpression':
            case 'SpreadElement':
            case 'NonNullExpression':
                break;
            case 'BinaryExpression': {
                if(this.operator(id) === '') return this.node(node.children[1] ?? panic('missing binary token')).kind;
                const left = this.unsupported(node.children[0] ?? panic('missing left'));
                return left !== '' ? left : this.unsupported(node.children[2] ?? panic('missing right'));
            }
            case 'PropertyAccessExpression':
            case 'ElementAccessExpression': {
                const left = this.unsupported(node.children[0] ?? panic('missing object'));
                return left !== ''
                    ? left
                    : this.unsupported(node.children[node.children.length - 1] ?? panic('missing property'));
            }
            case 'CallExpression':
            case 'NewExpression': {
                const callee = this.child(id, 0);
                if(this.node(callee).kind !== 'Identifier') return 'call-callee-layout';
                const count = Math.max(node.list, 0);
                const first = node.children.length - count;
                if(
                    first !== 1 &&
                    !(
                        first === 2 &&
                        this.node(node.children[1] ?? panic('missing optional token')).kind === 'QuestionDotToken'
                    )
                )
                    return 'type-arguments';
                for(let position = first; position < node.children.length; position++) {
                    const arg = this.node(this.child(id, position));
                    if(arg.kind === 'ArrayLiteralExpression') return 'expanded-call-argument';
                    const reason = this.unsupported(node.children[position] ?? panic('missing argument'));
                    if(reason !== '') return reason;
                }
                return '';
            }
            default:
                return node.kind;
        }
        for(const child of node.children) {
            const reason = this.unsupported(child);
            if(reason !== '') return reason;
        }
        return '';
    }
    normalize(index: number): number {
        const id = this.unwrapped(index);
        if(
            this.node(index).kind === 'ParenthesizedExpression' &&
            this.node(id).kind === 'BinaryExpression' &&
            this.operator(id) === ','
        )
            this.sequenceBoundaries.add(id);
        const node = this.node(id);
        for(let position = 0; position < node.children.length; position++)
            node.children[position] = this.normalize(node.children[position] ?? panic('missing normalization child'));
        if(node.kind !== 'BinaryExpression' || !['&&', '||', '??'].includes(this.operator(id))) return id;
        const right = this.child(id, 2);
        if(this.node(right).kind !== 'BinaryExpression' || this.operator(right) !== this.operator(id)) return id;
        // Match estree/postprocess.go: preserve evaluation order while rotating
        // associative logical expressions into the printer's left-leaning tree.
        const left = new ParseNode(
            'BinaryExpression',
            this.node(this.child(id, 0)).pos,
            this.node(this.child(right, 0)).end,
            [this.child(id, 0), node.children[1] ?? panic('missing normalization operator'), this.child(right, 0)],
        );
        const leftIndex = this.parser.nodes.length;
        this.parser.nodes.push(left);
        node.children[0] = this.normalize(leftIndex);
        node.children[2] = this.child(right, 2);
        return this.normalize(id);
    }
    parenthesize(index: number, parent: number, role: string): boolean {
        if(parent < 0) return false;
        const node = this.node(index);
        const outer = this.node(parent);
        if(node.kind === 'NumericLiteral')
            return role === 'object' && ['PropertyAccessExpression', 'ElementAccessExpression'].includes(outer.kind);
        const unary = [
            'PrefixUnaryExpression',
            'PostfixUnaryExpression',
            'DeleteExpression',
            'VoidExpression',
            'TypeOfExpression',
        ].includes(node.kind);
        if(unary) {
            const operator = this.operator(index);
            const update = node.kind === 'PostfixUnaryExpression' || ['++', '--'].includes(operator);
            if(outer.kind === 'PrefixUnaryExpression' && !['++', '--'].includes(this.operator(parent))) {
                if(update)
                    return (
                        node.kind !== 'PostfixUnaryExpression' &&
                        ((operator === '++' && this.operator(parent) === '+') ||
                            (operator === '--' && this.operator(parent) === '-'))
                    );
                return operator === this.operator(parent) && ['+', '-'].includes(operator);
            }
            if(outer.kind === 'BinaryExpression')
                return (
                    role === 'left' &&
                    (this.operator(parent) === '**' ||
                        (!update && ['in', 'instanceof'].includes(this.operator(parent))))
                );
            return role === 'object' || role === 'callee' || outer.kind === 'NonNullExpression';
        }
        if(node.kind === 'BinaryExpression') {
            if(this.operator(index) === ',') return true;
            if(outer.kind === 'BinaryExpression') {
                const operator = this.operator(index);
                const other = this.operator(parent);
                if(['??', '||', '&&'].includes(operator) && ['??', '||', '&&'].includes(other))
                    return operator !== other;
                if(rank(other) > rank(operator)) return true;
                if(rank(other) === rank(operator) && (role === 'right' || !flatten(other, operator))) return true;
                if(operator === '%' && (other === '+' || other === '-')) return true;
                return ['|', '^', '&', '<<', '>>', '>>>'].includes(other);
            }
            return (
                role === 'object' ||
                role === 'callee' ||
                [
                    'PrefixUnaryExpression',
                    'PostfixUnaryExpression',
                    'DeleteExpression',
                    'VoidExpression',
                    'TypeOfExpression',
                    'NonNullExpression',
                    'SpreadElement',
                ].includes(outer.kind)
            );
        }
        return false;
    }
    // ESTree flattens the left comma spine, stopping at explicit parentheses.
    sequenceParts(index: number): number[] {
        if(this.node(index).kind !== 'BinaryExpression' || this.operator(index) !== ',') return [index];
        const left = this.child(index, 0);
        return (this.sequenceBoundaries.has(left) ? [left] : this.sequenceParts(left)).concat([this.child(index, 2)]);
    }
    binaryParts(index: number, parent: number): number[] {
        const parts: number[] = [];
        const operator = this.operator(index);
        const left = this.child(index, 0);
        const right = this.child(index, 2);
        if(this.node(left).kind === 'BinaryExpression' && flatten(operator, this.operator(left))) {
            for(const part of this.binaryParts(left, index)) parts.push(part);
        }
        else parts.push(this.docs.group(this.print(left, index, 'left')));
        const inline =
            ['&&', '||', '??'].includes(operator) &&
            this.node(right).kind === 'ArrayLiteralExpression' &&
            this.node(right).children.length > 0;
        let rightDoc = this.docs.concat([
            this.docs.text(operator),
            inline ? this.docs.text(' ') : this.docs.line(),
            this.print(right, index, 'right'),
            this.docs.text(''),
        ]);
        const logical = ['&&', '||', '??'].includes(operator);
        const binaryType = (id: number): boolean =>
            this.node(id).kind === 'BinaryExpression' && ['&&', '||', '??'].includes(this.operator(id)) === logical;
        if((parent < 0 || !binaryType(parent)) && !binaryType(left) && !binaryType(right))
            rightDoc = this.docs.group(rightDoc);
        parts.push(this.docs.text(' '));
        parts.push(rightDoc);
        return parts;
    }
    print(index: number, parent = -1, role = ''): number {
        const id = this.unwrapped(index);
        const node = this.node(id);
        let result = -1;
        this.ancestors.push(id);
        const raw = this.source.slice(node.pos, node.end).trim();
        switch(node.kind) {
            case 'Identifier':
            case 'PrivateIdentifier':
                result = this.docs.text(node.text);
                break;
            case 'NumericLiteral':
                result = this.docs.text(numberText(raw));
                break;
            case 'BigIntLiteral':
                result = this.docs.text(raw.toLowerCase());
                break;
            case 'StringLiteral':
                result = this.docs.text(stringText(raw, parent < 0 && this.directive));
                break;
            case 'RegularExpressionLiteral': {
                const slash = raw.lastIndexOf('/');
                const flags = raw.slice(slash + 1).split('');
                flags.sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));
                result = this.docs.text(raw.slice(0, slash + 1) + flags.join(''));
                break;
            }
            case 'NoSubstitutionTemplateLiteral':
                result = this.docs.text(raw);
                break;
            case 'ThisKeyword':
                result = this.docs.text('this');
                break;
            case 'SuperKeyword':
                result = this.docs.text('super');
                break;
            case 'NullKeyword':
                result = this.docs.text('null');
                break;
            case 'TrueKeyword':
                result = this.docs.text('true');
                break;
            case 'FalseKeyword':
                result = this.docs.text('false');
                break;
            case 'OmittedExpression':
                result = this.docs.text('');
                break;
            case 'PrefixUnaryExpression':
            case 'PostfixUnaryExpression':
            case 'DeleteExpression':
            case 'VoidExpression':
            case 'TypeOfExpression': {
                const operator = this.operator(id);
                const operand = this.print(this.child(id, 0), id, 'argument');
                const prefix = this.docs.text(
                    operator +
                        (operator.length > 0 && operator.slice(-1) >= 'a' && operator.slice(-1) <= 'z' ? ' ' : ''),
                );
                result = this.docs.concat(
                    node.kind === 'PostfixUnaryExpression' ? [operand, prefix] : [prefix, operand],
                );
                break;
            }
            case 'NonNullExpression':
                result = this.docs.concat([this.print(this.child(id, 0), id, 'argument'), this.docs.text('!')]);
                break;
            case 'SpreadElement':
                result = this.docs.concat([this.docs.text('...'), this.print(this.child(id, 0), id, 'argument')]);
                break;
            case 'BinaryExpression': {
                if(this.operator(id) === ',') {
                    const sequence = this.sequenceParts(id);
                    const parts: number[] = [];
                    for(let position = 0; position < sequence.length; position++) {
                        const printed = this.print(
                            sequence[position] ?? panic('missing sequence item'),
                            id,
                            'expressions',
                        );
                        if(position === 0) parts.push(printed);
                        else {
                            parts.push(this.docs.text(','));
                            parts.push(
                                parent < 0
                                    ? this.docs.indent(this.docs.concat([this.docs.line(), printed]))
                                    : this.docs.concat([this.docs.line(), printed]),
                            );
                        }
                    }
                    result = this.docs.group(this.docs.concat(parts));
                    break;
                }
                const parts = this.binaryParts(id, parent);
                const outer = parent < 0 ? '' : this.node(parent).kind;
                if(
                    (role === 'object' && outer === 'PropertyAccessExpression') ||
                    role === 'callee' ||
                    ['PrefixUnaryExpression', 'DeleteExpression', 'VoidExpression', 'TypeOfExpression'].includes(outer)
                )
                    result = this.docs.group(
                        this.docs.concat([
                            this.docs.indent(this.docs.concat([this.docs.softline(), this.docs.concat(parts)])),
                            this.docs.softline(),
                        ]),
                    );
                else {
                    const inline =
                        ['&&', '||', '??'].includes(this.operator(id)) &&
                        this.node(this.child(id, 2)).kind === 'ArrayLiteralExpression' &&
                        this.node(this.child(id, 2)).children.length > 0;
                    const coercion =
                        role === 'arguments' &&
                        outer === 'CallExpression' &&
                        this.node(parent).list === 1 &&
                        this.node(parent).children.length === 2 &&
                        this.node(this.child(parent, 0)).text === 'Boolean';
                    result =
                        inline || coercion
                            ? this.docs.group(this.docs.concat(parts))
                            : this.docs.group(
                                  this.docs.concat([
                                      parts[0] ?? panic('missing binary head'),
                                      this.docs.indent(this.docs.concat(parts.slice(1))),
                                  ]),
                              );
                }
                break;
            }
            case 'ArrayLiteralExpression': {
                const printed: number[] = [];
                let numeric = true;
                let nested = node.children.length > 1;
                for(const child of node.children) {
                    const item = this.node(this.unwrapped(child));
                    if(
                        item.kind !== 'NumericLiteral' &&
                        !(
                            item.kind === 'PrefixUnaryExpression' &&
                            ['+', '-'].includes(this.operator(child)) &&
                            this.node(this.child(child, 0)).kind === 'NumericLiteral'
                        )
                    )
                        numeric = false;
                    if(item.kind !== 'ArrayLiteralExpression' || item.children.length <= 1) nested = false;
                    printed.push(this.print(child, id, 'elements'));
                }
                if(printed.length === 0) {
                    result = this.docs.text('[]');
                    break;
                }
                const key = `array-${id}`;
                const last = this.node(
                    this.unwrapped(node.children[node.children.length - 1] ?? panic('missing last element')),
                );
                const comma =
                    last.kind === 'OmittedExpression'
                        ? this.docs.text(',')
                        : this.docs.ifBreak(this.docs.text(','), this.docs.text(''), numeric ? key : '');
                let content: number;
                if(numeric) {
                    const parts: number[] = [];
                    for(let position = 0; position < printed.length; position++) {
                        parts.push(
                            this.docs.concat([
                                printed[position] ?? panic('missing element doc'),
                                position === printed.length - 1 ? comma : this.docs.text(','),
                            ]),
                        );
                        if(position < printed.length - 1) parts.push(this.docs.line());
                    }
                    content = this.docs.fill(parts);
                }
                else
                    content = this.docs.concat([
                        this.docs.join(this.docs.concat([this.docs.text(','), this.docs.line()]), printed),
                        comma,
                    ]);
                result = this.docs.group(
                    this.docs.concat([
                        this.docs.text('['),
                        this.docs.indent(this.docs.concat([this.docs.softline(), content])),
                        this.docs.softline(),
                        this.docs.text(']'),
                    ]),
                    key,
                    nested,
                );
                break;
            }
            case 'PropertyAccessExpression':
            case 'ElementAccessExpression': {
                const object = this.child(id, 0);
                const property = this.child(id, node.children.length - 1);
                const optional =
                    node.children.length > 2 &&
                    this.node(node.children[1] ?? panic('missing optional token')).kind === 'QuestionDotToken'
                        ? '?'
                        : '';
                const value = this.print(property, id, 'property');
                const before = this.print(object, id, 'object');
                if(node.kind === 'ElementAccessExpression') {
                    const contents =
                        this.node(property).kind === 'NumericLiteral'
                            ? value
                            : this.docs.group(
                                  this.docs.concat([
                                      this.docs.indent(this.docs.concat([this.docs.softline(), value])),
                                      this.docs.softline(),
                                  ]),
                              );
                    result = this.docs.concat([
                        before,
                        this.docs.text(optional === '?' ? '?.[' : '['),
                        contents,
                        this.docs.text(']'),
                    ]);
                }
                else {
                    const lookup = this.docs.concat([this.docs.text(`${optional}.`), value]);
                    let ancestor = this.ancestors.length - 2;
                    while(
                        ancestor >= 0 &&
                        this.node(this.ancestors[ancestor] ?? panic('missing member ancestor')).kind ===
                            'NonNullExpression'
                    )
                        ancestor--;
                    const memberParent =
                        ancestor >= 0 &&
                        ['PropertyAccessExpression', 'ElementAccessExpression'].includes(
                            this.node(this.ancestors[ancestor] ?? panic('missing member ancestor')).kind,
                        );
                    const inline =
                        this.node(object).kind === 'Identifier' &&
                        this.node(property).kind === 'Identifier' &&
                        !memberParent;
                    result = this.docs.concat([
                        before,
                        inline
                            ? lookup
                            : this.docs.group(this.docs.indent(this.docs.concat([this.docs.softline(), lookup]))),
                    ]);
                }
                break;
            }
            case 'CallExpression':
            case 'NewExpression': {
                const count = Math.max(node.list, 0);
                const args: number[] = [];
                for(let position = node.children.length - count; position < node.children.length; position++)
                    args.push(this.print(this.child(id, position), id, 'arguments'));
                const optional = node.children.length - count === 2 ? '?.' : '';
                const argumentDoc =
                    args.length === 0
                        ? this.docs.text('()')
                        : this.docs.group(
                              this.docs.concat([
                                  this.docs.text('('),
                                  this.docs.indent(
                                      this.docs.concat([
                                          this.docs.softline(),
                                          this.docs.join(
                                              this.docs.concat([this.docs.text(','), this.docs.line()]),
                                              args,
                                          ),
                                      ]),
                                  ),
                                  this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
                                  this.docs.softline(),
                                  this.docs.text(')'),
                              ]),
                          );
                result = this.docs.concat([
                    this.docs.text(node.kind === 'NewExpression' ? 'new ' : ''),
                    this.print(this.child(id, 0), id, 'callee'),
                    this.docs.text(optional),
                    argumentDoc,
                ]);
                break;
            }
            default:
                panic(`unclassified expression ${node.kind}`);
        }
        this.ancestors.pop();
        return this.parenthesize(id, parent, role)
            ? this.docs.concat([this.docs.text('('), result, this.docs.text(')')])
            : result;
    }
}
export function formatExpression(source: string, settings: SettingsOptions): ResultType {
    if(source.includes('/*') || source.includes('//')) return { kind: 'NotYet', reason: 'comment-attachment' };
    if(hasBlankLine(source) || source.includes('\r')) return { kind: 'NotYet', reason: 'source-trivia' };
    const parser = new Parser(source, 'expression.ts');
    const root = parser.expression();
    if(parser.kind() === 'SemicolonToken') parser.next();
    if(parser.kind() !== 'EndOfFile') return { kind: 'NotYet', reason: 'expression-file' };
    const docs = new Documents(settings);
    const printer = new Expressions(parser, source, docs);
    const reason = printer.unsupported(root);
    if(reason !== '') return { kind: 'NotYet', reason };
    const parenthesizedString =
        printer.node(root).kind === 'ParenthesizedExpression' &&
        printer.node(printer.unwrapped(root)).kind === 'StringLiteral';
    printer.directive = !parenthesizedString;
    let printed = printer.print(printer.normalize(root));
    if(
        parenthesizedString ||
        (printer.node(printer.unwrapped(root)).kind === 'BinaryExpression' &&
            printer.operator(printer.unwrapped(root)) === ',')
    )
        printed = docs.concat([docs.text('('), printed, docs.text(')')]);
    return { kind: 'Ok', text: docs.print(docs.concat([printed, docs.text(';'), docs.hardline()])) };
}
