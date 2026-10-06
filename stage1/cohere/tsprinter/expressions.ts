// Expression paths through cohere's print.go, print_binaryish.go, print_array.go and parentheses.go.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { ParseNode } from '../../typescript/parser/nodes.ts';
import { Documents, type SettingsOptions } from './doc.ts';
import { numberText, stringText } from './literals.ts';
import { isKeyName } from './keys.ts';
import { stringWidth } from './width.ts';
export type ResultType =
    { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'NotYet'; readonly reason: string };
const assignmentOperators: readonly string[] = [
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
        if(
            this.node(index).kind === 'ParenthesizedExpression' &&
            this.node(this.unwrapped(index)).kind !== 'ObjectLiteralExpression' &&
            this.hasOptional(index)
        )
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
                    const reason = this.unsupported(node.children[offset] ?? panic('missing conditional operand'));
                    if(reason !== '') return reason;
                }
                return '';
            }
            case 'BinaryExpression': {
                if(this.operator(id) === '') return this.node(node.children[1] ?? panic('missing binary token')).kind;
                if(
                    this.isAssignment(id) &&
                    ![
                        'Identifier',
                        'PropertyAccessExpression',
                        'ElementAccessExpression',
                        'NonNullExpression',
                    ].includes(this.node(this.child(id, 0)).kind)
                )
                    return 'assignment-pattern';
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
                const calleeReason = this.unsupported(node.children[0] ?? panic('missing callee'));
                if(calleeReason !== '') return calleeReason;
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
        const node = this.node(index);
        if(node.kind === 'ObjectLiteralExpression') return this.leftmost(this.ancestors[0] ?? index) === index;
        if(parent < 0) return false;
        const outer = this.node(parent);
        if(outer.kind === 'NewExpression' && role === 'callee') {
            let current = index;
            while(
                [
                    'PropertyAccessExpression',
                    'ElementAccessExpression',
                    'NonNullExpression',
                    'TaggedTemplateExpression',
                ].includes(this.node(current).kind)
            )
                current = this.child(current, 0);
            if(this.node(current).kind === 'CallExpression') return true;
        }
        if(
            node.kind === 'Identifier' &&
            node.text === 'let' &&
            role === 'object' &&
            outer.kind === 'ElementAccessExpression' &&
            !outer.optional &&
            this.leftmost(this.ancestors[0] ?? index) === index
        )
            return true;
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
        if(node.kind === 'ConditionalExpression')
            return (
                role === 'object' ||
                role === 'callee' ||
                (outer.kind === 'ConditionalExpression' && role === 'test') ||
                (outer.kind === 'BinaryExpression' && !this.isAssignment(parent) && this.operator(parent) !== ',') ||
                [
                    'PrefixUnaryExpression',
                    'PostfixUnaryExpression',
                    'DeleteExpression',
                    'VoidExpression',
                    'TypeOfExpression',
                    'NonNullExpression',
                    'SpreadElement',
                    'SpreadAssignment',
                ].includes(outer.kind)
            );
        if(node.kind === 'BinaryExpression') {
            if(this.isAssignment(index)) return !this.isAssignment(parent);
            if(this.operator(index) === ',') return true;
            if(this.isAssignment(parent)) return false;
            if(outer.kind === 'ConditionalExpression') return this.operator(index) === '??';
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
                    'SpreadAssignment',
                ].includes(outer.kind)
            );
        }
        return false;
    }
    leftmost(index: number): number {
        const node = this.node(index);
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
            return this.leftmost(this.child(index, 0));
        return index;
    }
    objectDoc(index: number): number {
        const node = this.node(index);
        if(node.children.length === 0) return this.docs.group(this.docs.text('{}'));
        const properties: number[] = [];
        for(const child of node.children) properties.push(this.print(child, index, 'properties'));
        const content = this.docs.concat([
            this.docs.text('{'),
            this.docs.indent(
                this.docs.concat([
                    this.docs.line(),
                    this.docs.join(this.docs.concat([this.docs.text(','), this.docs.line()]), properties),
                ]),
            ),
            this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
            this.docs.line(),
            this.docs.text('}'),
        ]);
        const first = this.node(node.children[0] ?? panic('missing object property'));
        const opening = this.source.indexOf('{', node.pos);
        const prefix = this.source.slice(opening + 1, first.pos);
        // The parser's first property position includes leading whitespace.
        const firstRaw = this.source.slice(first.pos, first.end);
        const start = firstRaw.length - firstRaw.trimStart().length;
        const leading = prefix + firstRaw.slice(0, start);
        return this.docs.group(content, '', leading.includes('\n'));
    }
    propertyKey(index: number): number {
        const keyIndex = this.child(index, 0);
        const key = this.node(keyIndex);
        if(key.kind === 'ComputedPropertyName')
            return this.docs.concat([
                this.docs.text('['),
                this.print(this.child(keyIndex, 0), keyIndex, 'expression'),
                this.docs.text(']'),
            ]);
        if(key.kind === 'StringLiteral') {
            const printed = stringText(this.source.slice(key.pos, key.end).trim(), false);
            if(printed.slice(1, -1) === key.text && isKeyName(key.text)) return this.docs.text(key.text);
        }
        return this.print(keyIndex, index, 'key');
    }
    conditionalDoc(index: number, parent: number, role: string): number {
        const outer = parent < 0 ? '' : this.node(parent).kind;
        const nested = outer === 'ConditionalExpression';
        const parentTest = nested && role === 'test';
        const forceNoIndent = nested && !parentTest;
        let ancestor = this.ancestors.length - 2;
        let child = index;
        while(ancestor >= 0) {
            const current = this.ancestors[ancestor] ?? panic('missing conditional ancestor');
            if(this.node(current).kind !== 'ConditionalExpression' || this.child(current, 0) === child) break;
            child = current;
            ancestor--;
        }
        const firstNonConditional = this.ancestors[ancestor] ?? -1;
        const consequent = this.child(index, 2);
        const alternate = this.child(index, 4);
        const consequentConditional = this.node(consequent).kind === 'ConditionalExpression';
        const alignedConsequent = this.docs.settings.useTabs
            ? this.docs.indent(this.print(consequent, index, 'consequent'))
            : this.docs.add('alignWidth', [this.print(consequent, index, 'consequent')], '', 2);
        const alignedAlternate = this.docs.settings.useTabs
            ? this.docs.indent(this.print(alternate, index, 'alternate'))
            : this.docs.add('alignWidth', [this.print(alternate, index, 'alternate')], '', 2);
        let parts = this.docs.concat([
            this.docs.line(),
            this.docs.text('? '),
            consequentConditional ? this.docs.ifBreak(this.docs.text(''), this.docs.text('(')) : this.docs.text(''),
            alignedConsequent,
            consequentConditional ? this.docs.ifBreak(this.docs.text(''), this.docs.text(')')) : this.docs.text(''),
            this.docs.line(),
            this.docs.text(': '),
            alignedAlternate,
        ]);
        if(nested && role === 'consequent')
            parts = this.docs.settings.useTabs
                ? this.docs.add('alignWidth', [this.docs.indent(parts)], '', -1)
                : this.docs.add('alignWidth', [parts], '', Math.max(0, this.docs.settings.tabWidth - 2));
        let test = this.print(this.child(index, 0), index, 'test');
        if(nested && role === 'alternate') test = this.docs.add('alignWidth', [test], '', 2);
        let chainChild = index;
        let chainAncestor = this.ancestors.length - 2;
        while(chainAncestor >= 0) {
            const current = this.ancestors[chainAncestor] ?? panic('missing conditional chain ancestor');
            if(
                ![
                    'NonNullExpression',
                    'PropertyAccessExpression',
                    'ElementAccessExpression',
                    'CallExpression',
                ].includes(this.node(current).kind) ||
                this.child(current, 0) !== chainChild
            )
                break;
            chainChild = current;
            chainAncestor--;
        }
        const enclosing = this.ancestors[chainAncestor] ?? -1;
        const extra =
            chainChild !== index &&
            enclosing >= 0 &&
            ((this.isAssignment(enclosing) && this.child(enclosing, 2) === chainChild) ||
                (['PrefixUnaryExpression', 'DeleteExpression', 'VoidExpression', 'TypeOfExpression'].includes(
                    this.node(enclosing).kind,
                ) &&
                    !['++', '--'].includes(this.operator(enclosing))));
        let result = this.docs.concat([
            test,
            forceNoIndent ? parts : this.docs.indent(parts),
            outer === 'PropertyAccessExpression' && !extra ? this.docs.softline() : this.docs.text(''),
        ]);
        if(parent === firstNonConditional) result = this.docs.group(result);
        if(parentTest || extra)
            result = this.docs.group(
                this.docs.concat([
                    this.docs.indent(this.docs.concat([this.docs.softline(), result])),
                    this.docs.softline(),
                ]),
            );
        return result;
    }
    isAssignment(index: number): boolean {
        if(index < 0 || this.node(index).kind !== 'BinaryExpression') return false;
        return assignmentOperators.includes(this.operator(index));
    }
    inlineLogical(index: number): boolean {
        return (
            this.node(index).kind === 'BinaryExpression' &&
            ['&&', '||', '??'].includes(this.operator(index)) &&
            ['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(this.node(this.child(index, 2)).kind) &&
            this.node(this.child(index, 2)).children.length > 0
        );
    }
    shortArgument(index: number, preferSingle = true): boolean {
        const node = this.node(index);
        const threshold = this.docs.settings.printWidth * 0.25;
        const raw = this.source.slice(node.pos, node.end).trim();
        switch(node.kind) {
            case 'ThisKeyword':
                return true;
            case 'Identifier':
                return node.text.length <= threshold;
            case 'NumericLiteral':
            case 'BigIntLiteral':
            case 'NullKeyword':
            case 'TrueKeyword':
            case 'FalseKeyword':
                return true;
            case 'RegularExpressionLiteral': {
                const pattern = raw.slice(1, raw.lastIndexOf('/'));
                return pattern === '' || pattern.length <= threshold;
            }
            case 'StringLiteral':
                return stringText(raw, false, preferSingle).length <= threshold;
            case 'NoSubstitutionTemplateLiteral':
                return raw.slice(1, -1).length <= threshold && !raw.includes('\n');
            case 'PrefixUnaryExpression':
                if(['++', '--'].includes(this.operator(index))) return false;
                if(
                    ['+', '-'].includes(this.operator(index)) &&
                    this.node(this.child(index, 0)).kind === 'NumericLiteral'
                )
                    return true;
                return this.shortArgument(this.child(index, 0), false);
            case 'DeleteExpression':
            case 'VoidExpression':
            case 'TypeOfExpression':
                return this.shortArgument(this.child(index, 0), false);
            case 'CallExpression':
                return (
                    !node.optional &&
                    node.list === 0 &&
                    this.node(this.child(index, 0)).kind === 'Identifier' &&
                    this.node(this.child(index, 0)).text.length <= threshold - 2
                );
            default:
                return false;
        }
    }
    poorlyBreakable(index: number, deep = false): boolean {
        const node = this.node(index);
        if(node.kind === 'NonNullExpression') return this.poorlyBreakable(this.child(index, 0), deep);
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression')
            return this.poorlyBreakable(this.child(index, 0), true);
        if(node.kind === 'CallExpression') {
            if(this.memberish(this.child(index, 0))) {
                const printed = this.print(index, this.ancestors[this.ancestors.length - 1] ?? -1, 'right');
                if(this.docs.get(printed).kind === 'label') return false;
            }
            if(node.list !== 0 && !(node.list === 1 && this.shortArgument(this.child(index, node.children.length - 1))))
                return false;
            return this.poorlyBreakable(this.child(index, 0), true);
        }
        return deep && (node.kind === 'Identifier' || node.kind === 'ThisKeyword');
    }
    breakAfterOperator(index: number, shortKey = false): boolean {
        const node = this.node(index);
        if(node.kind === 'BinaryExpression' && !this.isAssignment(index) && !this.inlineLogical(index)) return true;
        if(node.kind === 'ConditionalExpression') {
            const test = this.child(index, 0);
            if(
                this.node(test).kind === 'BinaryExpression' &&
                rank(this.operator(test)) >= 0 &&
                !this.inlineLogical(test)
            )
                return true;
        }
        if(shortKey) return false;
        let current = index;
        while(
            [
                'PrefixUnaryExpression',
                'DeleteExpression',
                'VoidExpression',
                'TypeOfExpression',
                'NonNullExpression',
            ].includes(this.node(current).kind) &&
            !['++', '--'].includes(this.operator(current))
        )
            current = this.child(current, 0);
        return this.node(current).kind === 'StringLiteral' || this.poorlyBreakable(current);
    }
    assignmentDoc(index: number, parent: number, left: number, right: number): number {
        return this.valueDoc(
            index,
            parent,
            left,
            right,
            this.child(index, 2),
            this.docs.text(` ${this.operator(index)}`),
        );
    }
    valueDoc(
        index: number,
        parent: number,
        left: number,
        right: number,
        rightIndex: number,
        operator: number,
        shortKey = false,
    ): number {
        const tail = !this.isAssignment(rightIndex);
        const ancestor = this.ancestors[this.ancestors.length - 3] ?? -1;
        const chain = this.isAssignment(parent) && (!tail || ancestor >= 0);
        if(chain)
            return this.docs.concat([
                this.docs.group(left),
                operator,
                tail
                    ? this.docs.indent(this.docs.concat([this.docs.line(), right]))
                    : this.docs.concat([this.docs.line(), right]),
            ]);
        const rightNode = this.node(rightIndex);
        const requireCall =
            rightNode.kind === 'CallExpression' && this.node(this.child(rightIndex, 0)).text === 'require';
        if(requireCall)
            return this.docs.group(this.docs.concat([this.docs.group(left), operator, this.docs.text(' '), right]));
        if((!tail && this.isAssignment(this.child(rightIndex, 2))) || this.breakAfterOperator(rightIndex, shortKey))
            return this.docs.group(
                this.docs.concat([
                    this.docs.group(left),
                    operator,
                    this.docs.group(this.docs.indent(this.docs.concat([this.docs.line(), right]))),
                ]),
            );
        if(
            !this.canBreak(left) &&
            (shortKey ||
                ['TrueKeyword', 'FalseKeyword', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(
                    rightNode.kind,
                ))
        )
            return this.docs.group(this.docs.concat([this.docs.group(left), operator, this.docs.text(' '), right]));
        const key = `assignment-${index}`;
        return this.docs.group(
            this.docs.concat([
                this.docs.group(left),
                operator,
                this.docs.group(this.docs.indent(this.docs.line()), key),
                this.docs.add('lineSuffixBoundary', []),
                this.docs.add('indentIfBreak', [right], '', 0, key),
            ]),
        );
    }
    ownOptional(index: number): boolean {
        for(const child of this.node(index).children) if(this.node(child).kind === 'QuestionDotToken') return true;
        return false;
    }
    memberish(index: number): boolean {
        return ['PropertyAccessExpression', 'ElementAccessExpression'].includes(this.node(index).kind);
    }
    simpleArgument(index: number, depth = 2): boolean {
        if(depth <= 0) return false;
        const node = this.node(index);
        if(node.kind === 'NonNullExpression') return this.simpleArgument(this.child(index, 0), depth);
        if(node.kind === 'RegularExpressionLiteral') {
            const raw = this.source.slice(node.pos, node.end).trim();
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
        if(node.kind === 'NoSubstitutionTemplateLiteral') return !this.source.slice(node.pos, node.end).includes('\n');
        if(node.kind === 'ObjectLiteralExpression') {
            for(const child of node.children) {
                const property = this.node(child);
                if(property.kind === 'ShorthandPropertyAssignment') continue;
                if(
                    property.kind !== 'PropertyAssignment' ||
                    this.node(this.child(child, 0)).kind === 'ComputedPropertyName' ||
                    !this.simpleArgument(this.child(child, 1), depth - 1)
                )
                    return false;
            }
            return true;
        }
        if(node.kind === 'ArrayLiteralExpression') {
            for(const child of node.children)
                if(
                    this.node(child).kind !== 'OmittedExpression' &&
                    !this.simpleArgument(this.unwrapped(child), depth - 1)
                )
                    return false;
            return true;
        }
        if(['CallExpression', 'NewExpression'].includes(node.kind)) {
            if(!this.simpleArgument(this.child(index, 0), depth) || node.list > depth) return false;
            for(
                let position = node.children.length - Math.max(node.list, 0);
                position < node.children.length;
                position++
            )
                if(!this.simpleArgument(this.child(index, position), depth - 1)) return false;
            return true;
        }
        if(this.memberish(index))
            return (
                this.simpleArgument(this.child(index, 0), depth) &&
                this.simpleArgument(this.child(index, node.children.length - 1), depth)
            );
        if(
            node.kind === 'PostfixUnaryExpression' ||
            (node.kind === 'PrefixUnaryExpression' && ['!', '-', '+', '~', '++', '--'].includes(this.operator(index)))
        )
            return this.simpleArgument(this.child(index, 0), depth);
        return false;
    }
    factory(name: string): boolean {
        if(name.charCodeAt(0) >= 65 && name.charCodeAt(0) <= 90) return true;
        if(name === '') return false;
        for(let index = 0; index < name.length; index++)
            if(!['_', '$'].includes(name.slice(index, index + 1))) return false;
        return true;
    }
    memberLookup(index: number): number {
        const node = this.node(index);
        const property = this.child(index, node.children.length - 1);
        if(node.kind === 'PropertyAccessExpression')
            return this.docs.concat([
                this.docs.text(this.ownOptional(index) ? '?.' : '.'),
                this.print(property, index, 'property'),
            ]);
        const printed = this.print(property, index, 'property');
        return this.docs.concat([
            this.docs.text(this.ownOptional(index) ? '?.[' : '['),
            this.node(property).kind === 'NumericLiteral'
                ? printed
                : this.docs.group(
                      this.docs.concat([
                          this.docs.indent(this.docs.concat([this.docs.softline(), printed])),
                          this.docs.softline(),
                      ]),
                  ),
            this.docs.text(']'),
        ]);
    }
    /**
     * @mutates nodes append reverse-order chain links for the caller to reverse once
     * @mutates docs append each corresponding doc in the same order as nodes
     */
    chainWalk(index: number, parent: number, role: string, nodes: number[], docs: number[]): void {
        const node = this.node(index);
        if(
            !this.parenthesize(index, parent, role) &&
            ((node.kind === 'CallExpression' &&
                (this.memberish(this.child(index, 0)) || this.node(this.child(index, 0)).kind === 'CallExpression')) ||
                this.memberish(index) ||
                node.kind === 'NonNullExpression')
        ) {
            this.ancestors.push(index);
            let printed: number;
            if(node.kind === 'CallExpression')
                printed = this.docs.concat([
                    this.docs.text(this.ownOptional(index) ? '?.' : ''),
                    this.argumentsDoc(index, parent),
                ]);
            else if(node.kind === 'NonNullExpression') printed = this.docs.text('!');
            else printed = this.memberLookup(index);
            nodes.push(index);
            docs.push(printed);
            this.chainWalk(
                this.child(index, 0),
                index,
                node.kind === 'CallExpression' ? 'callee' : node.kind === 'NonNullExpression' ? 'argument' : 'object',
                nodes,
                docs,
            );
            this.ancestors.pop();
        }
        else {
            nodes.push(index);
            docs.push(this.print(index, parent, role));
        }
    }
    memberChainDoc(index: number, parent: number): number {
        const reverseNodes = [index];
        const reverseDocs = [
            this.docs.concat([this.docs.text(this.ownOptional(index) ? '?.' : ''), this.argumentsDoc(index, parent)]),
        ];
        this.chainWalk(this.child(index, 0), index, 'callee', reverseNodes, reverseDocs);
        const nodes: number[] = [];
        const docs: number[] = [];
        for(let position = reverseNodes.length - 1; position >= 0; position--) {
            nodes.push(reverseNodes[position] ?? panic('missing chain node'));
            docs.push(reverseDocs[position] ?? panic('missing chain doc'));
        }
        const groups: number[][] = [];
        let current = [0];
        let position = 1;
        for(; position < nodes.length; position++) {
            const item = nodes[position] ?? panic('missing first-group node');
            const kind = this.node(item).kind;
            if(
                kind === 'NonNullExpression' ||
                kind === 'CallExpression' ||
                (kind === 'ElementAccessExpression' &&
                    this.node(this.child(item, this.node(item).children.length - 1)).kind === 'NumericLiteral')
            )
                current.push(position);
            else break;
        }
        if(this.node(nodes[0] ?? panic('missing chain base')).kind !== 'CallExpression') {
            for(; position + 1 < nodes.length; position++) {
                if(
                    this.memberish(nodes[position] ?? panic('missing chain member')) &&
                    this.memberish(nodes[position + 1] ?? panic('missing next member'))
                )
                    current.push(position);
                else break;
            }
        }
        groups.push(current);
        current = [];
        let seenCall = false;
        for(; position < nodes.length; position++) {
            const item = nodes[position] ?? panic('missing chain item');
            if(seenCall && this.memberish(item)) {
                if(
                    this.node(item).kind === 'ElementAccessExpression' &&
                    this.node(this.child(item, this.node(item).children.length - 1)).kind === 'NumericLiteral'
                ) {
                    current.push(position);
                    continue;
                }
                groups.push(current);
                current = [];
                seenCall = false;
            }
            if(this.node(item).kind === 'CallExpression') seenCall = true;
            current.push(position);
        }
        if(current.length > 0) groups.push(current);
        let merge = false;
        if(groups.length >= 2) {
            const firstGroup = groups[0] ?? panic('missing first chain group');
            const secondGroup = groups[1] ?? panic('missing second chain group');
            const computed =
                this.node(
                    nodes[secondGroup[0] ?? panic('missing second-group start')] ?? panic('missing second-group node'),
                ).kind === 'ElementAccessExpression';
            if(firstGroup.length === 1) {
                const base = this.node(nodes[firstGroup[0] ?? panic('missing base position')] ?? panic('missing base'));
                merge =
                    base.kind === 'ThisKeyword' ||
                    (base.kind === 'Identifier' &&
                        (this.factory(base.text) ||
                            (parent < 0 && base.text.length <= this.docs.settings.tabWidth) ||
                            computed));
            }
            else {
                const last =
                    nodes[firstGroup[firstGroup.length - 1] ?? panic('missing group end')] ??
                    panic('missing group-end node');
                if(this.memberish(last)) {
                    const property = this.node(this.child(last, this.node(last).children.length - 1));
                    merge = property.kind === 'Identifier' && (this.factory(property.text) || computed);
                }
            }
        }
        const printedGroups: number[] = [];
        for(const group of groups) {
            const parts: number[] = [];
            for(const item of group) parts.push(docs[item] ?? panic('missing grouped doc'));
            printedGroups.push(this.docs.concat(parts));
        }
        const oneLine = this.docs.concat(printedGroups);
        const cutoff = merge ? 3 : 2;
        const curried =
            parent >= 0 &&
            this.node(parent).kind === 'CallExpression' &&
            this.child(parent, 0) === index &&
            this.node(parent).list > 0 &&
            this.node(index).list > this.node(parent).list;
        if(groups.length <= cutoff) return curried ? oneLine : this.docs.group(oneLine);
        const prefix = this.docs.concat(printedGroups.slice(0, merge ? 2 : 1));
        const suffix = this.docs.indent(
            this.docs.concat([
                this.docs.hardline(),
                this.docs.join(this.docs.hardline(), printedGroups.slice(merge ? 2 : 1)),
            ]),
        );
        const expanded = this.docs.concat([prefix, suffix]);
        let calls = 0;
        let complex = false;
        for(const item of nodes)
            if(this.node(item).kind === 'CallExpression') {
                calls++;
                const call = this.node(item);
                for(let offset = call.children.length - Math.max(0, call.list); offset < call.children.length; offset++)
                    if(!this.simpleArgument(this.child(item, offset))) complex = true;
            }
        let brokenHead = false;
        for(let offset = 0; offset < printedGroups.length - 1; offset++)
            if(this.willBreak(printedGroups[offset] ?? panic('missing head group'))) brokenHead = true;
        const result =
            (calls > 2 && complex) || brokenHead
                ? this.docs.group(expanded)
                : this.docs.concat([
                      this.willBreak(oneLine) ? this.docs.add('breakParent', []) : this.docs.text(''),
                      this.docs.add('conditionalGroup', [oneLine, expanded]),
                  ]);
        return this.docs.add('label', [result], 'member-chain');
    }
    willBreak(index: number): boolean {
        const doc = this.docs.get(index);
        if(
            (['group', 'conditionalGroup'].includes(doc.kind) && doc.broken) ||
            ['breakParent', 'hardlineWithoutBreakParent', 'literallineWithoutBreakParent'].includes(doc.kind)
        )
            return true;
        const count = doc.kind === 'conditionalGroup' ? Math.min(1, doc.parts.length) : doc.parts.length;
        for(let position = 0; position < count; position++)
            if(this.willBreak(doc.parts[position] ?? panic('missing break child'))) return true;
        return false;
    }
    numericArray(index: number): boolean {
        const node = this.node(index);
        if(node.kind !== 'ArrayLiteralExpression' || node.children.length === 0) return false;
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
                return false;
        }
        return true;
    }
    argumentsDoc(index: number, parent: number): number {
        const node = this.node(index);
        const count = Math.max(node.list, 0);
        if(count === 0) return this.docs.group(this.docs.text('()'));
        const first = node.children.length - count;
        const args: number[] = [];
        for(let position = first; position < node.children.length; position++)
            args.push(this.print(this.child(index, position), index, 'arguments'));
        const callee = this.node(this.child(index, 0));
        const name = callee.kind === 'Identifier' ? callee.text : '';
        const firstArg = this.node(this.child(index, first));
        if(
            node.kind === 'CallExpression' &&
            !node.optional &&
            ((name === 'require' && (count > 1 || firstArg.kind === 'StringLiteral')) ||
                (name === 'define' &&
                    parent < 0 &&
                    (count === 1 ||
                        (count === 2 && firstArg.kind === 'ArrayLiteralExpression') ||
                        (count === 3 &&
                            firstArg.kind === 'StringLiteral' &&
                            this.node(this.child(index, first + 1)).kind === 'ArrayLiteralExpression'))))
        )
            return this.docs.concat([
                this.docs.text('('),
                this.docs.join(this.docs.text(', '), args),
                this.docs.text(')'),
            ]);
        const printed: number[] = [];
        for(let position = 0; position < args.length; position++)
            printed.push(
                position === args.length - 1
                    ? (args[position] ?? panic('missing last arg'))
                    : this.docs.concat([args[position] ?? panic('missing arg'), this.docs.text(','), this.docs.line()]),
            );
        const trailing = this.docs.ifBreak(this.docs.text(','), this.docs.text(''));
        const lastIndex = this.child(index, node.children.length - 1);
        const last = this.node(lastIndex);
        const previousKind = count > 1 ? this.node(this.child(index, node.children.length - 2)).kind : '';
        const expanded =
            ['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(last.kind) &&
            last.children.length > 0 &&
            previousKind !== last.kind &&
            !(count > 1 && this.numericArray(lastIndex));
        let anyBroken = false;
        let headBroken = false;
        for(let position = 0; position < printed.length; position++)
            if(this.willBreak(printed[position] ?? panic('missing printed arg'))) {
                anyBroken = true;
                if(position !== printed.length - 1) headBroken = true;
            }
        const allBroken = this.docs.group(
            this.docs.concat([
                this.docs.text('('),
                this.docs.indent(this.docs.concat([this.docs.line(), this.docs.concat(printed)])),
                trailing,
                this.docs.line(),
                this.docs.text(')'),
            ]),
            '',
            true,
        );
        if(expanded) {
            if(headBroken) return allBroken;
            const head = this.docs.concat(printed.slice(0, -1));
            const lastDoc = args[args.length - 1] ?? panic('missing expanded arg');
            const expandedDoc = this.docs.concat([
                this.docs.text('('),
                head,
                this.docs.group(lastDoc, '', true),
                this.docs.text(')'),
            ]);
            const states: number[] = [];
            if(!this.willBreak(lastDoc))
                states.push(this.docs.concat([this.docs.text('('), head, lastDoc, this.docs.text(')')]));
            states.push(expandedDoc);
            states.push(allBroken);
            const conditional = this.docs.add('conditionalGroup', states);
            return this.willBreak(lastDoc)
                ? this.docs.concat([this.docs.add('breakParent', []), conditional])
                : conditional;
        }
        const contents = this.docs.concat([
            this.docs.text('('),
            this.docs.indent(this.docs.concat([this.docs.softline(), this.docs.concat(printed)])),
            trailing,
            this.docs.softline(),
            this.docs.text(')'),
        ]);
        const curried =
            parent >= 0 &&
            this.node(parent).kind === 'CallExpression' &&
            this.child(parent, 0) === index &&
            this.node(parent).list > 0 &&
            node.list > this.node(parent).list;
        return curried ? contents : this.docs.group(contents, '', anyBroken);
    }
    canBreak(index: number): boolean {
        const doc = this.docs.get(index);
        if(['line', 'softline', 'hardlineWithoutBreakParent', 'literallineWithoutBreakParent'].includes(doc.kind))
            return true;
        for(const child of doc.parts) if(this.canBreak(child)) return true;
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
            this.ancestors.push(left);
            for(const part of this.binaryParts(left, index)) parts.push(part);
            this.ancestors.pop();
        }
        else parts.push(this.docs.group(this.print(left, index, 'left')));
        const inline =
            ['&&', '||', '??'].includes(operator) &&
            ['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(this.node(right).kind) &&
            this.node(right).children.length > 0;
        let rightDoc = this.docs.concat([
            this.docs.text(operator),
            inline ? this.docs.text(' ') : this.docs.line(),
            this.print(right, index, 'right'),
            this.docs.text(''),
        ]);
        const logical = ['&&', '||', '??'].includes(operator);
        const binaryType = (id: number): boolean =>
            this.node(id).kind === 'BinaryExpression' &&
            rank(this.operator(id)) >= 0 &&
            ['&&', '||', '??'].includes(this.operator(id)) === logical;
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
            case 'ObjectLiteralExpression':
                result = this.objectDoc(id);
                break;
            case 'PropertyAssignment': {
                const left = this.propertyKey(id);
                const value = this.child(id, 1);
                const keyDoc = this.docs.get(left);
                const shortKey = keyDoc.kind === 'text' && stringWidth(keyDoc.text) < this.docs.settings.tabWidth + 3;
                result = this.valueDoc(
                    id,
                    parent,
                    left,
                    this.print(value, id, 'value'),
                    value,
                    this.docs.text(':'),
                    shortKey,
                );
                break;
            }
            case 'ShorthandPropertyAssignment':
                result = this.print(this.child(id, 0), id, 'value');
                break;
            case 'SpreadAssignment':
                result = this.docs.concat([this.docs.text('...'), this.print(this.child(id, 0), id, 'argument')]);
                break;
            case 'ConditionalExpression':
                result = this.conditionalDoc(id, parent, role);
                break;
            case 'BinaryExpression': {
                if(this.isAssignment(id)) {
                    result = this.assignmentDoc(
                        id,
                        parent,
                        this.print(this.child(id, 0), id, 'left'),
                        this.print(this.child(id, 2), id, 'right'),
                    );
                    break;
                }
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
                        ['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(
                            this.node(this.child(id, 2)).kind,
                        ) &&
                        this.node(this.child(id, 2)).children.length > 0;
                    const coercion =
                        role === 'arguments' &&
                        outer === 'CallExpression' &&
                        this.node(parent).list === 1 &&
                        this.node(parent).children.length === 2 &&
                        this.node(this.child(parent, 0)).text === 'Boolean';
                    result =
                        inline ||
                        coercion ||
                        this.isAssignment(parent) ||
                        outer === 'PropertyAssignment' ||
                        (outer === 'ConditionalExpression' &&
                            (this.ancestors.length < 3 ||
                                !['CallExpression', 'NewExpression', 'ReturnStatement', 'ThrowStatement'].includes(
                                    this.node(this.ancestors[this.ancestors.length - 3] ?? panic('missing grandparent'))
                                        .kind,
                                )))
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
                    if(
                        !['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(item.kind) ||
                        item.children.length <= 1 ||
                        item.kind !== this.node(this.child(id, 0)).kind
                    )
                        nested = false;
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
                    let firstNonMember = this.ancestors.length - 2;
                    while(
                        firstNonMember >= 0 &&
                        ['PropertyAccessExpression', 'ElementAccessExpression', 'NonNullExpression'].includes(
                            this.node(this.ancestors[firstNonMember] ?? panic('missing ancestor')).kind,
                        )
                    )
                        firstNonMember--;
                    const enclosing = this.ancestors[firstNonMember] ?? -1;
                    const assignmentInline =
                        (this.isAssignment(enclosing) && this.node(this.child(enclosing, 0)).kind !== 'Identifier') ||
                        (this.isAssignment(this.ancestors[ancestor] ?? -1) &&
                            this.node(object).kind === 'CallExpression' &&
                            this.node(object).list > 0);
                    const inline =
                        assignmentInline ||
                        (this.node(object).kind === 'Identifier' &&
                            this.node(property).kind === 'Identifier' &&
                            !memberParent);
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
                const callee = this.child(id, 0);
                if(
                    node.kind === 'CallExpression' &&
                    this.memberish(callee) &&
                    !this.parenthesize(callee, id, 'callee')
                ) {
                    result = this.memberChainDoc(id, parent);
                    break;
                }
                const count = Math.max(node.list, 0);
                const optional = node.children.length - count === 2 ? '?.' : '';
                const argumentDoc = this.argumentsDoc(id, parent);
                result = this.docs.concat([
                    this.docs.text(node.kind === 'NewExpression' ? 'new ' : ''),
                    this.print(this.child(id, 0), id, 'callee'),
                    this.docs.text(optional),
                    argumentDoc,
                ]);
                if(this.node(callee).kind === 'CallExpression') result = this.docs.group(result);
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
