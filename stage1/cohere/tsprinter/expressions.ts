// Expression paths through cohere's print.go, print_binaryish.go, print_array.go and parentheses.go.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { ParseNode } from '../../typescript/parser/nodes.ts';
import { Documents, type SettingsOptions } from './doc.ts';
import { numberText, stringText } from './literals.ts';
import { isKeyName } from './keys.ts';
import { parenthesize } from './parentheses.ts';
import { stringWidth } from './width.ts';
import {
    assignmentOperators,
    unsupported,
    statementUnsupported,
    functionUnsupported,
    hasBlankLine,
    syntaxOperator,
    rank,
    flatten,
    simpleArgument,
    factory,
    couldExpandArgument,
    functionArgument,
    functionComposition,
    hookArguments,
    numericArray,
} from './syntax.ts';
export type ResultType =
    { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'NotYet'; readonly reason: string };
export class Expressions {
    readonly parser: Parser;
    readonly source: string;
    readonly docs: Documents;
    readonly state = {
        directive: true,
        expandedArrow: -1,
        expandedFirstArrow: -1,
        statementExpressionRoot: -1,
    };
    readonly ancestors: number[] = [];
    readonly sequenceBoundaries = new Set<number>();
    readonly optionalBoundaries = new Set<number>();
    readonly protectedStrings = new Set<number>();
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
        return syntaxOperator(this.parser, index);
    }
    activeOptional(index: number): boolean {
        const node = this.node(index);
        if(node.kind === 'ParenthesizedExpression' || this.optionalBoundaries.has(index)) return false;
        if(node.kind === 'NonNullExpression' || this.memberish(index) || node.kind === 'CallExpression')
            return this.ownOptional(index) || this.activeOptional(node.children[0] ?? panic('missing chain receiver'));
        return false;
    }
    hasOptional(index: number): boolean {
        const node = this.node(index);
        if(node.kind === 'QuestionDotToken') return true;
        for(const child of node.children) if(this.hasOptional(child)) return true;
        return false;
    }
    unsupported(index: number): string {
        return unsupported(this.parser, this.source, index);
    }
    normalize(index: number): number {
        const id = this.unwrapped(index);
        if(this.node(index).kind === 'ParenthesizedExpression' && this.activeOptional(id))
            this.optionalBoundaries.add(id);
        if(
            this.node(index).kind === 'ParenthesizedExpression' &&
            this.node(id).kind === 'BinaryExpression' &&
            this.operator(id) === ','
        )
            this.sequenceBoundaries.add(id);
        if(this.node(index).kind === 'ParenthesizedExpression' && this.node(id).kind === 'StringLiteral')
            this.protectedStrings.add(id);
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
        return parenthesize(
            this.parser,
            this.state.statementExpressionRoot,
            this.ancestors,
            this.optionalBoundaries,
            index,
            parent,
            role,
        );
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
    propertyKey(index: number, offset = 0): number {
        const keyIndex = this.child(index, offset);
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
                (this.node(enclosing).kind === 'VariableDeclaration' && this.child(enclosing, 1) === chainChild) ||
                (['ReturnStatement', 'ThrowStatement', 'AwaitExpression'].includes(this.node(enclosing).kind) &&
                    this.child(enclosing, 0) === chainChild) ||
                (this.node(enclosing).kind === 'YieldExpression' &&
                    this.child(enclosing, this.node(enclosing).children.length - 1) === chainChild) ||
                (['PrefixUnaryExpression', 'DeleteExpression', 'VoidExpression', 'TypeOfExpression'].includes(
                    this.node(enclosing).kind,
                ) &&
                    !['++', '--'].includes(this.operator(enclosing))));
        let result = this.docs.concat([
            test,
            forceNoIndent ? parts : this.docs.indent(parts),
            outer === 'PropertyAccessExpression' && !extra ? this.docs.softline() : this.docs.text(''),
        ]);
        if(parent < 0 || parent === firstNonConditional) result = this.docs.group(result);
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
                'AwaitExpression',
                'YieldExpression',
            ].includes(this.node(current).kind) &&
            !(this.node(current).kind === 'YieldExpression' && this.node(current).children.length === 0) &&
            !['++', '--'].includes(this.operator(current))
        )
            current = this.child(
                current,
                this.node(current).kind === 'YieldExpression' ? this.node(current).children.length - 1 : 0,
            );
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
        const assignedParent =
            this.isAssignment(parent) || (parent >= 0 && this.node(parent).kind === 'VariableDeclaration');
        const chain =
            assignedParent &&
            (!tail ||
                (ancestor >= 0 &&
                    !['ExpressionStatement', 'VariableDeclarationList', 'VariableStatement'].includes(
                        this.node(ancestor).kind,
                    )));
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
            !this.docs.canBreak(left) &&
            (shortKey ||
                [
                    'TrueKeyword',
                    'FalseKeyword',
                    'NumericLiteral',
                    'NoSubstitutionTemplateLiteral',
                    'TemplateExpression',
                    'TaggedTemplateExpression',
                ].includes(rightNode.kind))
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
    parameterDoc(index: number, parent: number): number {
        const children = this.node(index).children;
        const rest = this.node(children[0] ?? panic('missing parameter')).kind === 'DotDotDotToken';
        const name = children[rest ? 1 : 0] ?? panic('missing parameter name');
        const left = this.docs.text(`${rest ? '...' : ''}${this.node(name).text}`);
        if(children.length === (rest ? 2 : 1)) return left;
        const initializer = children[rest ? 2 : 1] ?? panic('missing parameter initializer');
        return this.valueDoc(
            index,
            parent,
            left,
            this.print(initializer, index, 'right'),
            initializer,
            this.docs.text(' ='),
            false,
        );
    }
    parametersDoc(index: number): number {
        const printed: number[] = [];
        let rest = false;
        for(const child of this.node(index).children) {
            if(this.node(child).kind !== 'Parameter') continue;
            if(printed.length > 0) printed.push(this.docs.concat([this.docs.text(','), this.docs.line()]));
            printed.push(this.parameterDoc(child, index));
            rest = this.node(this.node(child).children[0] ?? panic('missing parameter')).kind === 'DotDotDotToken';
        }
        if(printed.length === 0) return this.docs.text('()');
        if(
            index === this.state.expandedArrow ||
            (index === this.state.expandedFirstArrow && this.node(index).kind === 'ArrowFunction')
        )
            return this.docs.group(
                this.docs.concat([
                    this.docs.text('('),
                    this.docs.removeLines(this.docs.concat(printed)),
                    this.docs.text(')'),
                ]),
            );
        return this.docs.group(
            this.docs.concat([
                this.docs.text('('),
                this.docs.indent(this.docs.concat([this.docs.softline(), ...printed])),
                rest ? this.docs.text('') : this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
                this.docs.softline(),
                this.docs.text(')'),
            ]),
        );
    }
    arrowDoc(index: number, parent: number, role: string): number {
        const signatures: number[] = [];
        let current = index;
        let forceBreak = false;
        let body: number;
        while(true) {
            const node = this.node(current);
            let async = false;
            for(const child of node.children) if(this.node(child).kind === 'AsyncKeyword') async = true;
            signatures.push(this.docs.concat([this.docs.text(async ? 'async ' : ''), this.parametersDoc(current)]));
            for(const child of node.children) {
                if(this.node(child).kind === 'Parameter' && this.node(child).children.length !== 1) forceBreak = true;
            }
            body = this.unwrapped(node.children[node.children.length - 1] ?? panic('missing arrow body'));
            if(this.node(body).kind !== 'ArrowFunction' || index === this.state.expandedArrow) break;
            current = body;
            this.ancestors.push(current);
        }
        const chain = signatures.length > 1;
        forceBreak = chain && forceBreak;
        const callee =
            role === 'callee' && parent >= 0 && ['CallExpression', 'NewExpression'].includes(this.node(parent).kind);
        const assigned =
            parent >= 0 &&
            (this.isAssignment(parent) ||
                this.node(parent).kind === 'PropertyAssignment' ||
                this.node(parent).kind === 'VariableDeclaration' ||
                this.node(parent).kind === 'Parameter');
        const kind = this.node(body).kind;
        const conditional =
            kind === 'ConditionalExpression' && this.node(this.leftmost(body)).kind !== 'ObjectLiteralExpression';
        const sameLine =
            ['ArrayLiteralExpression', 'ObjectLiteralExpression', 'Block', 'ArrowFunction'].includes(kind) ||
            (kind === 'BinaryExpression' && this.operator(body) === ',') ||
            this.templateOwnLine(body) ||
            (!forceBreak && conditional);
        let printedBody = this.print(body, current, 'body');
        for(let position = 1; position < signatures.length; position++) this.ancestors.pop();
        if(sameLine && conditional)
            printedBody = this.docs.group(
                this.docs.concat([
                    this.docs.ifBreak(this.docs.text(''), this.docs.text('(')),
                    this.docs.indent(this.docs.concat([this.docs.softline(), printedBody])),
                    this.docs.ifBreak(this.docs.text(''), this.docs.text(')')),
                    index === this.state.expandedArrow
                        ? this.docs.ifBreak(this.docs.text(','), this.docs.text(''))
                        : this.docs.text(''),
                    index === this.state.expandedArrow ? this.docs.softline() : this.docs.text(''),
                ]),
            );
        printedBody = sameLine
            ? this.docs.concat([this.docs.text(' '), printedBody])
            : this.docs.indent(this.docs.concat([this.docs.line(), printedBody]));
        if(index === this.state.expandedArrow && !sameLine)
            printedBody = this.docs.concat([
                printedBody,
                this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
                this.docs.softline(),
            ]);
        const parts: number[] = [];
        for(let position = 0; position < signatures.length; position++) {
            if(position > 0) parts.push(this.docs.concat([this.docs.text(' =>'), this.docs.line()]));
            parts.push(signatures[position] ?? panic('missing arrow signature'));
        }
        let printedSignatures = this.docs.concat(parts);
        if(
            chain &&
            parent >= 0 &&
            !callee &&
            (['CallExpression', 'NewExpression'].includes(this.node(parent).kind) ||
                (this.node(parent).kind === 'BinaryExpression' &&
                    !this.isAssignment(parent) &&
                    this.operator(parent) !== ','))
        ) {
            printedSignatures = this.docs.concat([
                signatures[0] ?? panic('missing arrow signature'),
                this.docs.text(' =>'),
                this.docs.indent(this.docs.concat([this.docs.line(), this.docs.concat(parts.slice(2))])),
            ]);
        }
        else if(chain && !callee && !assigned) printedSignatures = this.docs.indent(printedSignatures);
        if(chain) printedSignatures = this.docs.group(printedSignatures, '', forceBreak);
        if(chain && (callee || assigned))
            printedSignatures = this.docs.indent(this.docs.concat([this.docs.softline(), printedSignatures]));
        const key = `arrow-${index}`;
        return this.docs.group(
            this.docs.concat([
                this.docs.group(printedSignatures, key, chain && callee && !sameLine),
                this.docs.text(' =>'),
                chain ? this.docs.add('indentIfBreak', [printedBody], '', 0, key) : this.docs.group(printedBody),
                chain && callee ? this.docs.ifBreak(this.docs.softline(), this.docs.text(''), key) : this.docs.text(''),
            ]),
        );
    }
    importSpecifierDoc(index: number): number {
        const node = this.node(index);
        if(node.kind === 'NamespaceImport')
            return this.docs.concat([this.docs.text('* as '), this.print(this.child(index, 0), index)]);
        const parts: number[] = [];
        if(node.semantic === '1') parts.push(this.docs.text('type '));
        parts.push(this.print(this.child(index, 0), index));
        // Explicit `x as x` has distinct source locations and retains its alias.
        if(node.children.length === 2) {
            parts.push(this.docs.text(' as '));
            parts.push(this.print(this.child(index, 1), index));
        }
        return this.docs.concat(parts);
    }
    importClauseDoc(index: number): number {
        const node = this.node(index);
        const standalone: number[] = [];
        const grouped: number[] = [];
        for(const child of node.children) {
            const part = this.node(child);
            if(part.kind === 'NamedImports') {
                for(const specifier of part.children) grouped.push(this.importSpecifierDoc(specifier));
            }
            else if(part.kind === 'NamespaceImport') standalone.push(this.importSpecifierDoc(child));
            else standalone.push(this.print(child, index));
        }
        const parts: number[] = [this.docs.text(' '), this.docs.join(this.docs.text(', '), standalone)];
        if(grouped.length > 0) {
            if(standalone.length > 0) parts.push(this.docs.text(', '));
            if(grouped.length > 1 || standalone.length > 0) {
                parts.push(this.docs.group(this.docs.concat([
                    this.docs.text('{'),
                    this.docs.indent(this.docs.concat([
                        this.docs.line(),
                        this.docs.join(this.docs.concat([this.docs.text(','), this.docs.line()]), grouped),
                    ])),
                    this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
                    this.docs.line(),
                    this.docs.text('}'),
                ])));
            }
            else parts.push(this.docs.concat([this.docs.text('{ '), grouped[0] ?? panic('missing import specifier'), this.docs.text(' }')]));
        }
        else if(standalone.length === 0) parts.push(this.docs.text('{}'));
        return this.docs.concat(parts);
    }
    importAttributesDoc(index: number): number {
        const node = this.node(index);
        if(node.children.length === 0) return this.docs.text('{}');
        const properties: number[] = [];
        for(const child of node.children) properties.push(this.docs.concat([
            this.propertyKey(child),
            this.docs.text(': '),
            this.print(this.child(child, 1), child),
        ]));
        const content = this.docs.group(this.docs.concat([
            this.docs.text('{'),
            this.docs.indent(this.docs.concat([
                this.docs.line(),
                this.docs.join(this.docs.concat([this.docs.text(','), this.docs.line()]), properties),
            ])),
            this.docs.ifBreak(this.docs.text(','), this.docs.text('')),
            this.docs.line(),
            this.docs.text('}'),
        ]), '', node.multiLine);
        const first = this.child(index, 0);
        const key = this.node(this.child(first, 0));
        // Go removes lines for exactly one string-valued `type` attribute,
        // even when the original braces contain newlines or the value is long.
        return node.children.length === 1 && key.text === 'type'
            ? this.docs.removeLines(content)
            : content;
    }
    importDoc(index: number): number {
        const parts: number[] = [this.docs.text('import')];
        let offset = 0;
        const first = this.child(index, 0);
        if(this.node(first).kind === 'ImportClause') {
            const phase = this.node(first).semantic;
            if(phase === 'TypeKeyword') parts.push(this.docs.text(' type'));
            if(phase === 'DeferKeyword') parts.push(this.docs.text(' defer'));
            parts.push(this.importClauseDoc(first));
            parts.push(this.docs.text(' from'));
            offset = 1;
        }
        parts.push(this.docs.text(' '));
        parts.push(this.print(this.child(index, offset), index));
        if(this.node(index).children.length > offset + 1) {
            const attributes = this.child(index, offset + 1);
            parts.push(this.docs.text(this.node(attributes).operator === 'AssertKeyword' ? ' assert ' : ' with '));
            parts.push(this.importAttributesDoc(attributes));
        }
        parts.push(this.docs.text(';'));
        return this.docs.concat(parts);
    }
    statementUnsupported(index: number): string {
        return statementUnsupported(this.parser, this.source, index);
    }
    statementDoc(index: number, parent: number): number {
        const pushed = this.ancestors[this.ancestors.length - 1] !== index;
        if(pushed) this.ancestors.push(index);
        const result = this.statementBodyDoc(index, parent);
        if(pushed) this.ancestors.pop();
        return result;
    }
    statementBodyDoc(index: number, parent: number): number {
        const node = this.node(index);
        const parts: number[] = [];
        switch(node.kind) {
            case 'SourceFile': {
                let directive = true;
                for(const child of node.children) {
                    const statement = this.node(child);
                    if(statement.kind === 'EndOfFile') continue;
                    const expression =
                        statement.kind === 'ExpressionStatement'
                            ? this.unwrapped(statement.children[0] ?? panic('missing program directive'))
                            : -1;
                    if(
                        expression < 0 ||
                        this.node(expression).kind !== 'StringLiteral' ||
                        this.protectedStrings.has(expression)
                    )
                        directive = false;
                    if(statement.kind === 'EmptyStatement') continue;
                    if(parts.length > 0) parts.push(this.docs.hardline());
                    const previous = this.state.directive;
                    this.state.directive = directive;
                    parts.push(this.statementDoc(child, index));
                    this.state.directive = previous;
                }
                return this.docs.concat(parts);
            }
            case 'ImportDeclaration':
                return this.importDoc(index);
            case 'FunctionDeclaration':
                return this.functionDoc(index);
            case 'VariableStatement':
                return this.variableDoc(this.child(index, 0));
            case 'Block': {
                let directive = true;
                for(const child of node.children) {
                    const statement = this.node(child);
                    const expression =
                        statement.kind === 'ExpressionStatement'
                            ? this.unwrapped(statement.children[0] ?? panic('missing directive'))
                            : -1;
                    if(
                        expression < 0 ||
                        this.node(expression).kind !== 'StringLiteral' ||
                        this.protectedStrings.has(expression)
                    )
                        directive = false;
                    if(statement.kind === 'EmptyStatement') continue;
                    if(parts.length > 0) parts.push(this.docs.hardline());
                    const previousDirective = this.state.directive;
                    this.state.directive = directive;
                    parts.push(this.statementDoc(child, index));
                    this.state.directive = previousDirective;
                }
                if(parts.length === 0) {
                    const compact =
                        parent >= 0 &&
                        [
                            'ArrowFunction',
                            'FunctionExpression',
                            'FunctionDeclaration',
                            'MethodDeclaration',
                            'GetAccessor',
                            'SetAccessor',
                            'ForStatement',
                            'WhileStatement',
                            'DoStatement',
                        ].includes(this.node(parent).kind);
                    return this.docs.concat([
                        this.docs.text('{'),
                        compact ? this.docs.text('') : this.docs.hardline(),
                        this.docs.text('}'),
                    ]);
                }
                return this.docs.concat([
                    this.docs.text('{'),
                    this.docs.indent(this.docs.concat([this.docs.hardline(), this.docs.concat(parts)])),
                    this.docs.hardline(),
                    this.docs.text('}'),
                ]);
            }
            case 'ExpressionStatement': {
                const expression = this.unwrapped(node.children[0] ?? panic('missing statement expression'));
                const previousRoot = this.state.statementExpressionRoot;
                this.state.statementExpressionRoot = expression;
                let printed = this.print(expression, -1, '');
                this.state.statementExpressionRoot = previousRoot;
                if(
                    (this.node(expression).kind === 'StringLiteral' &&
                        (!this.state.directive || this.protectedStrings.has(expression))) ||
                    (this.node(expression).kind === 'BinaryExpression' && this.operator(expression) === ',')
                )
                    printed = this.docs.concat([this.docs.text('('), printed, this.docs.text(')')]);
                return this.docs.concat([printed, this.docs.text(';')]);
            }
            case 'ReturnStatement':
            case 'ThrowStatement': {
                parts.push(this.docs.text(node.kind === 'ReturnStatement' ? 'return' : 'throw'));
                if(node.children.length > 0) {
                    const expression = this.unwrapped(node.children[0] ?? panic('missing return argument'));
                    let printed = this.print(expression, index, 'argument');
                    if(
                        this.node(expression).kind === 'BinaryExpression' &&
                        !this.isAssignment(expression) &&
                        this.operator(expression) !== ','
                    )
                        printed = this.docs.group(
                            this.docs.concat([
                                this.docs.ifBreak(this.docs.text('('), this.docs.text('')),
                                this.docs.indent(this.docs.concat([this.docs.softline(), printed])),
                                this.docs.softline(),
                                this.docs.ifBreak(this.docs.text(')'), this.docs.text('')),
                            ]),
                        );
                    parts.push(this.docs.concat([this.docs.text(' '), printed]));
                }
                parts.push(this.docs.text(';'));
                return this.docs.concat(parts);
            }
            case 'BreakStatement':
            case 'ContinueStatement':
                return this.docs.concat([
                    this.docs.text(node.kind === 'BreakStatement' ? 'break' : 'continue'),
                    this.docs.text(
                        node.children.length === 0
                            ? ''
                            : ` ${this.node(node.children[0] ?? panic('missing label')).text}`,
                    ),
                    this.docs.text(';'),
                ]);
            case 'DebuggerStatement':
                return this.docs.text('debugger;');
            case 'EmptyStatement':
                return this.docs.text(';');
            default:
                panic(`unclassified statement ${node.kind}`);
        }
    }
    functionUnsupported(index: number): string {
        return functionUnsupported(this.parser, this.source, index);
    }
    variableDoc(index: number): number {
        const list = this.node(index);
        const printed: number[] = [];
        let initialized = false;
        this.ancestors.push(index);
        for(const declaration of list.children) {
            const item = this.node(declaration);
            const name = this.child(declaration, 0);
            this.ancestors.push(declaration);
            const left = this.print(name, declaration, 'id');
            if(item.children.length === 1) printed.push(left);
            else {
                initialized = true;
                const value = this.child(declaration, 1);
                const right = this.print(value, declaration, 'init');
                printed.push(this.valueDoc(declaration, index, left, right, value, this.docs.text(' =')));
            }
            this.ancestors.pop();
        }
        this.ancestors.pop();
        const kind =
            list.semantic === '2'
                ? 'const'
                : list.semantic === '1'
                  ? 'let'
                  : list.semantic === '4'
                    ? 'using'
                    : list.semantic === '6'
                      ? 'await using'
                      : 'var';
        const first = printed[0] ?? panic('missing variable declaration');
        const rest: number[] = [];
        for(let position = 1; position < printed.length; position++)
            rest.push(
                this.docs.concat([
                    this.docs.text(','),
                    initialized ? this.docs.hardline() : this.docs.line(),
                    printed[position] ?? panic('missing variable doc'),
                ]),
            );
        return this.docs.group(
            this.docs.concat([
                this.docs.text(kind),
                this.docs.text(' '),
                printed.length === 1 ? first : this.docs.indent(first),
                this.docs.indent(this.docs.concat(rest)),
                this.docs.text(';'),
            ]),
        );
    }
    functionDoc(index: number): number {
        const node = this.node(index);
        let async = '';
        let generator = '';
        let name = '';
        for(const child of node.children) {
            const item = this.node(child);
            if(item.kind === 'AsyncKeyword') async = 'async ';
            if(item.kind === 'AsteriskToken') generator = '*';
            if(item.kind === 'Identifier') name = ` ${item.text}`;
        }
        const signature = this.docs.concat([
            this.docs.text(`${async}function${generator}${name}`),
            this.parametersDoc(index),
        ]);
        const body = node.children[node.children.length - 1] ?? panic('missing function child');
        if(node.kind === 'FunctionDeclaration' && this.node(body).kind !== 'Block')
            return this.docs.concat([signature, this.docs.text(';')]);
        return this.docs.concat([signature, this.docs.text(' '), this.statementDoc(body, index)]);
    }
    methodDoc(index: number): number {
        const node = this.node(index);
        let offset = 0;
        let prefix = node.kind === 'GetAccessor' ? 'get ' : node.kind === 'SetAccessor' ? 'set ' : '';
        while(offset < node.children.length) {
            const kind = this.node(node.children[offset] ?? panic('missing method prefix')).kind;
            if(kind === 'AsyncKeyword') prefix += 'async ';
            else if(kind === 'AsteriskToken') prefix += '*';
            else break;
            offset++;
        }
        return this.docs.concat([
            this.docs.text(prefix),
            this.propertyKey(index, offset),
            this.parametersDoc(index),
            this.docs.text(' '),
            this.statementDoc(node.children[node.children.length - 1] ?? panic('missing method body'), index),
        ]);
    }
    templateHasLines(index: number): boolean {
        const node = this.node(index);
        if(node.kind === 'TaggedTemplateExpression')
            return this.templateHasLines(this.child(index, node.children.length - 1));
        if(node.kind === 'NoSubstitutionTemplateLiteral') return this.source.slice(node.pos, node.end).includes('\n');
        if(node.kind !== 'TemplateExpression') return false;
        if(this.node(node.children[0] ?? panic('missing quasi head')).raw.includes('\n')) return true;
        for(let position = 1; position < node.children.length; position++) {
            const span = this.node(node.children[position] ?? panic('missing quasi span'));
            if(this.node(span.children[1] ?? panic('missing quasi tail')).raw.includes('\n')) return true;
        }
        return false;
    }
    templateOwnLine(index: number): boolean {
        if(!this.templateHasLines(index)) return false;
        const node = this.node(index);
        const raw = this.source.slice(node.pos, node.end);
        let position = node.pos + raw.length - raw.trimStart().length - 1;
        while(position >= 0 && [9, 32].includes(this.source.charCodeAt(position))) position--;
        return ![10, 13, 8232, 8233].includes(this.source.charCodeAt(position));
    }
    templateRaw(raw: string): number {
        const parts: number[] = [];
        const lines = raw.split('\n');
        for(let index = 0; index < lines.length; index++) {
            if(index !== 0)
                parts.push(
                    this.docs.concat([
                        this.docs.add('literallineWithoutBreakParent', []),
                        this.docs.add('breakParent', []),
                    ]),
                );
            parts.push(this.docs.text(lines[index] ?? panic('missing template line')));
        }
        return this.docs.concat(parts);
    }
    templateDoc(index: number): number {
        const node = this.node(index);
        const head = this.node(node.children[0] ?? panic('missing template head'));
        const parts = [this.docs.add('lineSuffixBoundary', []), this.docs.text('`'), this.templateRaw(head.raw)];
        let previous = head.raw;
        let indentSize = 0;
        for(let position = 1; position < node.children.length; position++) {
            const span = this.node(node.children[position] ?? panic('missing template span'));
            const expression = this.unwrapped(span.children[0] ?? panic('missing interpolation'));
            const literal = this.node(span.children[1] ?? panic('missing template quasi'));
            let expressionDoc = this.print(expression, index, 'expressions');
            const tailRaw = this.source.slice(literal.pos, literal.end);
            const interpolationEnd = literal.pos + tailRaw.length - tailRaw.trimStart().length;
            let newline = this.source.slice(span.pos, interpolationEnd).includes('\n');
            if(!newline) {
                const flatDocs = new Documents({
                    printWidth: 1e15,
                    tabWidth: this.docs.settings.tabWidth,
                    useTabs: this.docs.settings.useTabs,
                });
                const flatPrinter = new Expressions(this.parser, this.source, flatDocs);
                flatPrinter.state.directive = false;
                for(const ancestor of this.ancestors) flatPrinter.ancestors.push(ancestor);
                for(const boundary of this.sequenceBoundaries) flatPrinter.sequenceBoundaries.add(boundary);
                for(const boundary of this.optionalBoundaries) flatPrinter.optionalBoundaries.add(boundary);
                const rendered = flatDocs.print(flatPrinter.print(expression, index, 'expressions'));
                if(rendered.includes('\n')) newline = true;
                else expressionDoc = this.docs.text(rendered);
            }
            const expressionKind = this.node(expression).kind;
            if(
                newline &&
                ([
                    'Identifier',
                    'PropertyAccessExpression',
                    'ElementAccessExpression',
                    'NonNullExpression',
                    'ConditionalExpression',
                ].includes(expressionKind) ||
                    (expressionKind === 'BinaryExpression' && !this.isAssignment(expression)))
            )
                expressionDoc = this.docs.concat([
                    this.docs.indent(this.docs.concat([this.docs.softline(), expressionDoc])),
                    this.docs.softline(),
                ]);
            const lastNewline = previous.lastIndexOf('\n');
            if(lastNewline >= 0) {
                indentSize = 0;
                const remainder = previous.slice(lastNewline + 1);
                for(let offset = 0; offset < remainder.length; offset++) {
                    const character = remainder.charCodeAt(offset);
                    if(character === 32) indentSize++;
                    else if(character === 9)
                        indentSize += this.docs.settings.tabWidth - (indentSize % this.docs.settings.tabWidth);
                    else break;
                }
            }
            if(indentSize > 0) {
                for(let level = 0; level < Math.floor(indentSize / this.docs.settings.tabWidth); level++)
                    expressionDoc = this.docs.indent(expressionDoc);
                expressionDoc = this.docs.add(
                    'alignWidth',
                    [expressionDoc],
                    '',
                    indentSize % this.docs.settings.tabWidth,
                );
                expressionDoc = this.docs.add('dedentToRoot', [expressionDoc]);
            }
            else if(previous.endsWith('\n')) expressionDoc = this.docs.add('dedentToRoot', [expressionDoc]);
            parts.push(
                this.docs.group(
                    this.docs.concat([
                        this.docs.text('${'),
                        expressionDoc,
                        this.docs.add('lineSuffixBoundary', []),
                        this.docs.text('}'),
                    ]),
                ),
            );
            parts.push(this.templateRaw(literal.raw));
            previous = literal.raw;
        }
        parts.push(this.docs.text('`'));
        return this.docs.concat(parts);
    }
    memberish(index: number): boolean {
        return ['PropertyAccessExpression', 'ElementAccessExpression'].includes(this.node(index).kind);
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
                !this.optionalBoundaries.has(this.child(index, 0)) &&
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
                        (factory(base.text) ||
                            (parent < 0 && base.text.length <= this.docs.settings.tabWidth) ||
                            computed));
            }
            else {
                const last =
                    nodes[firstGroup[firstGroup.length - 1] ?? panic('missing group end')] ??
                    panic('missing group-end node');
                if(this.memberish(last)) {
                    const property = this.node(this.child(last, this.node(last).children.length - 1));
                    merge = property.kind === 'Identifier' && (factory(property.text) || computed);
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
                    if(!simpleArgument(this.parser, this.source, this.child(item, offset))) complex = true;
            }
        let brokenHead = false;
        for(let offset = 0; offset < printedGroups.length - 1; offset++)
            if(this.docs.willBreak(printedGroups[offset] ?? panic('missing head group'))) brokenHead = true;
        const result =
            (calls > 2 && complex) || brokenHead
                ? this.docs.group(expanded)
                : this.docs.concat([
                      this.docs.willBreak(oneLine) ? this.docs.add('breakParent', []) : this.docs.text(''),
                      this.docs.add('conditionalGroup', [oneLine, expanded]),
                  ]);
        return this.docs.add('label', [result], 'member-chain');
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
        if(count === 1 && this.templateOwnLine(this.child(index, first)))
            return this.docs.concat([
                this.docs.text('('),
                args[0] ?? panic('missing template argument'),
                this.docs.text(')'),
            ]);
        if(hookArguments(this.parser, index))
            return this.docs.concat([
                this.docs.text('('),
                this.docs.join(this.docs.text(', '), args),
                this.docs.text(')'),
            ]);
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
            couldExpandArgument(this.parser, lastIndex) &&
            !(count === 2 && previousKind === 'ArrowFunction' && last.kind === 'ArrayLiteralExpression') &&
            previousKind !== last.kind &&
            !(count > 1 && numericArray(this.parser, lastIndex));
        let anyBroken = false;
        let headBroken = false;
        for(let position = 0; position < printed.length; position++)
            if(this.docs.willBreak(printed[position] ?? panic('missing printed arg'))) {
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
        if(functionComposition(this.parser, index)) return allBroken;
        const firstIndex = this.child(index, first);
        const firstBody = firstArg.children[firstArg.children.length - 1];
        const expandFirst =
            count === 2 &&
            functionArgument(this.parser, firstIndex) &&
            firstBody !== undefined &&
            this.node(firstBody).kind === 'Block' &&
            !functionArgument(this.parser, lastIndex) &&
            last.kind !== 'ConditionalExpression' &&
            (simpleArgument(this.parser, this.source, lastIndex, 2) ||
                (last.kind === 'BinaryExpression' &&
                    simpleArgument(this.parser, this.source, this.child(lastIndex, 0), 1) &&
                    simpleArgument(this.parser, this.source, this.child(lastIndex, 2), 1))) &&
            !(last.kind === 'CallExpression' && last.list > 1) &&
            !couldExpandArgument(this.parser, lastIndex);
        if(expandFirst) {
            if(this.docs.willBreak(args[1] ?? panic('missing tail argument'))) return allBroken;
            for(const child of firstArg.children)
                if(this.node(child).kind === 'Parameter' && this.docs.willBreak(this.parameterDoc(child, firstIndex)))
                    return allBroken;
            const previous = this.state.expandedFirstArrow;
            this.state.expandedFirstArrow = firstIndex;
            const firstDoc = this.print(firstIndex, index, 'arguments');
            this.state.expandedFirstArrow = previous;
            const tail = args[1] ?? panic('missing tail argument');
            const flat = this.docs.concat([
                this.docs.text('('),
                firstDoc,
                this.docs.text(', '),
                tail,
                this.docs.text(')'),
            ]);
            const broken = this.docs.concat([
                this.docs.text('('),
                this.docs.group(firstDoc, '', true),
                this.docs.text(', '),
                tail,
                this.docs.text(')'),
            ]);
            const states = this.docs.willBreak(firstDoc) ? [broken, allBroken] : [flat, broken, allBroken];
            const group = this.docs.add('conditionalGroup', states);
            return this.docs.willBreak(firstDoc) ? this.docs.concat([this.docs.add('breakParent', []), group]) : group;
        }
        if(expanded) {
            if(headBroken) return allBroken;
            const head = this.docs.concat(printed.slice(0, -1));
            let lastDoc = args[args.length - 1] ?? panic('missing expanded arg');
            let expandFunctionParameters = true;
            if(last.kind === 'FunctionExpression')
                for(const child of last.children)
                    if(count === 1 && this.node(child).kind === 'Parameter' && this.node(child).children.length !== 1)
                        expandFunctionParameters = false;
            if(last.kind === 'ArrowFunction' || (last.kind === 'FunctionExpression' && expandFunctionParameters)) {
                for(const child of last.children)
                    if(
                        this.node(child).kind === 'Parameter' &&
                        this.docs.willBreak(this.parameterDoc(child, lastIndex))
                    )
                        return allBroken;
                const previous = this.state.expandedArrow;
                this.state.expandedArrow = lastIndex;
                lastDoc = this.print(lastIndex, index, 'arguments');
                this.state.expandedArrow = previous;
            }
            const expandedDoc = this.docs.concat([
                this.docs.text('('),
                head,
                this.docs.group(lastDoc, '', true),
                this.docs.text(')'),
            ]);
            const states: number[] = [];
            if(!this.docs.willBreak(lastDoc))
                states.push(this.docs.concat([this.docs.text('('), head, lastDoc, this.docs.text(')')]));
            states.push(expandedDoc);
            states.push(allBroken);
            const conditional = this.docs.add('conditionalGroup', states);
            return this.docs.willBreak(lastDoc)
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
                result = this.docs.text(stringText(raw, parent < 0 && this.state.directive));
                break;
            case 'RegularExpressionLiteral': {
                const slash = raw.lastIndexOf('/');
                const flags = raw.slice(slash + 1).split('');
                flags.sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));
                result = this.docs.text(raw.slice(0, slash + 1) + flags.join(''));
                break;
            }
            case 'NoSubstitutionTemplateLiteral':
                result = this.docs.concat([this.docs.add('lineSuffixBoundary', []), this.templateRaw(raw)]);
                break;
            case 'MethodDeclaration':
            case 'GetAccessor':
            case 'SetAccessor':
                result = this.methodDoc(id);
                break;
            case 'FunctionExpression':
                result = this.functionDoc(id);
                break;
            case 'ArrowFunction':
                result = this.arrowDoc(id, parent, role);
                break;
            case 'Block':
                result = this.statementDoc(id, parent);
                break;
            case 'TemplateExpression':
                result = this.templateDoc(id);
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
                const keyText = this.docs.textContent(left);
                const shortKey = keyText !== undefined && stringWidth(keyText) < this.docs.settings.tabWidth + 3;
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
                    const content = this.docs.concat(parts);
                    result =
                        parent >= 0 &&
                        ((this.node(parent).kind === 'ArrowFunction' && role === 'body') ||
                            (['ReturnStatement', 'ThrowStatement'].includes(this.node(parent).kind) &&
                                role === 'argument'))
                            ? this.docs.group(
                                  this.docs.ifBreak(
                                      this.docs.concat([
                                          this.docs.indent(this.docs.concat([this.docs.softline(), content])),
                                          this.docs.softline(),
                                      ]),
                                      content,
                                  ),
                              )
                            : this.docs.group(content);
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
                        outer === 'VariableDeclaration' ||
                        outer === 'TemplateExpression' ||
                        outer === 'ReturnStatement' ||
                        outer === 'ThrowStatement' ||
                        (outer === 'ArrowFunction' && role === 'body') ||
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
                        ((this.isAssignment(this.ancestors[ancestor] ?? -1) ||
                            (ancestor >= 0 &&
                                this.node(this.ancestors[ancestor] ?? panic('missing assignment ancestor')).kind ===
                                    'VariableDeclaration')) &&
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
            case 'AwaitExpression': {
                result = this.docs.concat([this.docs.text('await '), this.print(this.child(id, 0), id, 'argument')]);
                if(role === 'callee' || role === 'object') {
                    result = this.docs.concat([
                        this.docs.indent(this.docs.concat([this.docs.softline(), result])),
                        this.docs.softline(),
                    ]);
                    let enclosing = -1;
                    for(let position = this.ancestors.length - 2; position >= 0; position--) {
                        const ancestor = this.ancestors[position] ?? panic('missing await ancestor');
                        if(['AwaitExpression', 'Block'].includes(this.node(ancestor).kind)) {
                            enclosing = ancestor;
                            break;
                        }
                    }
                    if(
                        enclosing < 0 ||
                        this.node(enclosing).kind !== 'AwaitExpression' ||
                        this.leftmost(this.child(enclosing, 0)) !== id
                    )
                        result = this.docs.group(result);
                }
                break;
            }
            case 'YieldExpression': {
                const delegated =
                    node.children.length > 0 &&
                    this.node(node.children[0] ?? panic('missing yield token')).kind === 'AsteriskToken';
                const offset = delegated ? 1 : 0;
                result = this.docs.text(delegated ? 'yield*' : 'yield');
                if(node.children.length > offset)
                    result = this.docs.concat([
                        result,
                        this.docs.text(' '),
                        this.print(this.child(id, offset), id, 'argument'),
                    ]);
                break;
            }
            case 'TaggedTemplateExpression':
                result = this.docs.concat([
                    this.print(this.child(id, 0), id, 'tag'),
                    this.docs.add('lineSuffixBoundary', []),
                    this.print(this.child(id, 1), id, 'quasi'),
                ]);
                break;
            case 'CallExpression':
            case 'NewExpression': {
                const callee = this.child(id, 0);
                if(
                    node.kind === 'CallExpression' &&
                    !this.optionalBoundaries.has(callee) &&
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
    if(source.startsWith('#!') || source.startsWith('\ufeff#!') || source.includes('/*') || source.includes('//'))
        return { kind: 'NotYet', reason: 'comment-attachment' };
    if(hasBlankLine(source) || source.includes('\r')) return { kind: 'NotYet', reason: 'source-trivia' };
    const parser = new Parser(source, 'expression.ts');
    // Go formats the standalone text as a program: a leading brace is a block,
    // even when the selector extracted it from an object receiver.
    if(parser.kind() === 'OpenBraceToken') return formatProgram(parser, source, settings);
    // Go parses the text as a file, so its expression is read inside the source
    // elements' list context. Error recovery in a speculative parse (an arrow's
    // parameters, say) stops where an outer context owns the token, so without
    // this context `({ m() {} })` recovers into an arrow head and refuses.
    parser.beginList('source');
    const root = parser.expression();
    parser.endList('source');
    if(parser.kind() === 'SemicolonToken') parser.next();
    if(parser.kind() !== 'EndOfFile') return { kind: 'NotYet', reason: 'expression-file' };
    const docs = new Documents(settings);
    const printer = new Expressions(parser, source, docs);
    const reason = printer.unsupported(root);
    if(reason !== '') return { kind: 'NotYet', reason };
    const parenthesizedString =
        printer.node(root).kind === 'ParenthesizedExpression' &&
        printer.node(printer.unwrapped(root)).kind === 'StringLiteral';
    printer.state.directive = !parenthesizedString;
    let printed = printer.print(printer.normalize(root));
    if(
        parenthesizedString ||
        (printer.node(printer.unwrapped(root)).kind === 'BinaryExpression' &&
            printer.operator(printer.unwrapped(root)) === ',')
    )
        printed = docs.concat([docs.text('('), printed, docs.text(')')]);
    return { kind: 'Ok', text: docs.print(docs.concat([printed, docs.text(';'), docs.hardline()])) };
}

export function formatProgram(parser: Parser, source: string, settings: SettingsOptions): ResultType {
    const root = parser.file();
    const docs = new Documents(settings);
    const printer = new Expressions(parser, source, docs);
    const reason = printer.statementUnsupported(root);
    if(reason !== '') return { kind: 'NotYet', reason };
    const printed = docs.print(printer.statementDoc(printer.normalize(root), -1));
    return { kind: 'Ok', text: printed === '' ? '' : `${printed}\n` };
}
