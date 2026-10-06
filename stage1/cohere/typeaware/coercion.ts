import type { Rules } from './rules.ts';
import type { TypeFact } from './type_fact.ts';
import type { Types } from './types.ts';
import { types } from './facts.ts';
import { Diagnostic } from './diagnostic.ts';
import { Frames, header } from './frames.ts';
import { Suggestion } from './suggestion.ts';
import { Repair } from './repair.ts';
import { booleanLike, numberLike, bigintLike, stringLike, intersectionFlag } from './flags.ts';
import { Reassign } from './reassign.ts';

export class Coercion {
    readonly rules: Rules;
    constructor(rules: Rules) {
        this.rules = rules;
    }
    operand(index: number): string {
        const inner = this.rules.skip(index);
        const node = this.rules.parser.node(inner);
        const text = this.rules.text(node);
        return node.kind === 'BinaryExpression' &&
            this.rules.parser.node(node.children[1] ?? -1).kind === 'CommaToken' ? `(${text})` : text;
    }
    numeric(index: number): boolean {
        const node = this.rules.parser.node(index);
        const callee = node.kind === 'CallExpression' ? this.rules.parser.node(node.children[0] ?? -1) : node;
        return node.kind === 'NumericLiteral' || (node.kind === 'CallExpression' && callee.kind === 'Identifier' &&
            ['Number', 'parseInt', 'parseFloat'].includes(callee.text));
    }
    string(index: number): boolean {
        const node = this.rules.parser.node(index);
        const callee = node.kind === 'CallExpression' ? this.rules.parser.node(node.children[0] ?? -1) : node;
        return ['StringLiteral', 'NoSubstitutionTemplateLiteral', 'TemplateExpression'].includes(node.kind) ||
            (node.kind === 'CallExpression' && callee.kind === 'Identifier' && callee.text === 'String');
    }
    empty(index: number): boolean {
        const node = this.rules.parser.node(index);
        return ['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(node.kind) && node.text === '';
    }
    already(index: number, mask: number): boolean {
        const facts = types(this.rules.ask(index, 'type-shape'), 'type-shape');
        return facts.parts(facts.root()).every((id) => {
            const part = facts.type(id);
            return (part.flags & mask) !== 0 || ((part.flags & intersectionFlag) !== 0 &&
                part.parts.some((child) => (facts.type(child).flags & mask) !== 0));
        });
    }
    falsy(part: TypeFact, index: number): boolean {
        if((part.flags & 4096) !== 0) {
            return true;
        }
        if((part.flags & (1024 | 2048 | 8192)) === 0) {
            return false;
        }
        const frames = new Frames(this.rules.ask(index, `literal-value\n${part.id}`));
        header(frames, 'literal-value');
        const kind = frames.field();
        const value = frames.field();
        frames.end();
        return kind === 'missing' || (kind === 'boolean' && value === 'false') ||
            (kind === 'string' && value === '') || (kind === 'number' && ['0', '-0', 'NaN'].includes(value));
    }
    truthy(facts: Types, part: TypeFact, index: number): boolean {
        if((part.flags & intersectionFlag) !== 0) {
            return part.parts.every((id) => this.truthy(facts, facts.type(id), index));
        }
        return (part.flags & (1048576 | 131072 | 512 | 16384)) !== 0 ||
            ((part.flags & (1024 | 2048 | 8192)) !== 0 && !this.falsy(part, index));
    }
    reference(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const node = this.rules.parser.node(current);
            if(['Identifier', 'ThisKeyword', 'SuperKeyword'].includes(node.kind)) {
                return true;
            }
            if(!['PropertyAccessExpression', 'ElementAccessExpression', 'NonNullExpression',
                'ParenthesizedExpression'].includes(node.kind)) {
                return false;
            }
            current = node.children[0] ?? -1;
        }
        return false;
    }
    parentheses(index: number): boolean {
        const parentIndex = this.rules.parents[index] ?? -1;
        if(parentIndex < 0) {
            return true;
        }
        const parent = this.rules.parser.node(parentIndex);
        if(['VariableDeclaration', 'ParenthesizedExpression', 'IfStatement', 'WhileStatement',
            'DoStatement', 'ForStatement', 'ReturnStatement', 'ArrowFunction', 'CallExpression',
            'NewExpression', 'PropertyAssignment', 'JsxExpression', 'ConditionalExpression',
            'ExpressionStatement', 'ArrayLiteralExpression', 'TemplateSpan', 'ExportAssignment',
            'PropertyDeclaration', 'Parameter', 'BindingElement', 'ThrowStatement', 'CaseClause',
            'SwitchStatement'].includes(parent.kind)) {
            return false;
        }
        return parent.kind !== 'BinaryExpression' || !['AmpersandAmpersandToken', 'BarBarToken',
            'EqualsToken', 'CommaToken'].includes(this.rules.parser.node(parent.children[1] ?? -1).kind);
    }
    narrowing(index: number, operand: number): string {
        if(!this.reference(operand)) {
            return 'Boolean';
        }
        const facts = types(this.rules.ask(operand, 'raw-shape'), 'raw-shape');
        let nullable = false;
        let undefinable = false;
        let possiblyFalsy = false;
        for(const id of facts.parts(facts.root())) {
            const part = facts.type(id);
            if((part.flags & 1) !== 0) {
                return 'Boolean';
            }
            if((part.flags & (2 | 524288 | 33554432 | 67108864 | 16777216 | 2097152 | 4194304 | 8388608)) !== 0) {
                return '';
            }
            if((part.flags & 8) !== 0) {
                nullable = true;
            }
            else if((part.flags & (4 | 16)) !== 0) {
                undefinable = true;
            }
            else if(this.falsy(part, operand)) {
                return '';
            }
            else if(!this.truthy(facts, part, operand)) {
                possiblyFalsy = true;
            }
        }
        if(!nullable && !undefinable) {
            return 'Boolean';
        }
        if(possiblyFalsy) {
            return '';
        }
        const text = this.operand(operand);
        const comparison = nullable && undefinable ? `${text} !== null && ${text} !== undefined` :
            nullable ? `${text} !== null` : `${text} !== undefined`;
        return this.parentheses(index) ? `(${comparison})` : comparison;
    }
    global(index: number): boolean {
        const frames = new Frames(this.rules.ask(index, 'resolved-name\nBoolean'));
        header(frames, 'resolved-name');
        let global = false;
        if(frames.yes()) {
            frames.field();
            const count = frames.natural();
            global = count > 0;
            for(let at = 0; at < count; at++) {
                frames.field();
                global = frames.yes() && global;
                frames.yes();
            }
        }
        frames.end();
        return global;
    }
    identifier(code: number): boolean {
        return (code >= 65 && code <= 90) || (code >= 97 && code <= 122) ||
            (code >= 48 && code <= 57) || code === 95 || code === 36;
    }
    report(index: number, recommendation: string, suggest: boolean, fix: boolean): void {
        const node = this.rules.parser.node(index);
        const start = this.rules.start(node);
        const finding = new Diagnostic('no-implicit-coercion', 'implicitCoercion',
            `Unexpected implicit coercion encountered. Use \`${recommendation}\` instead. The shorthand converts a type as a side effect of an operator chosen for something else, so a reader has to know the coercion table to see what it produces. The explicit call says which type is wanted, and says it where the conversion happens.`,
            this.rules.byte(start), this.rules.byte(node.end), '');
        const replacement = start > 0 && this.identifier(this.rules.scanner.text.charCodeAt(start - 1)) &&
            this.identifier(recommendation.charCodeAt(0)) ? ` ${recommendation}` : recommendation;
        const repair = new Repair(finding.start, finding.end, replacement);
        if(fix) {
            finding.repairs.push(repair);
        }
        else if(suggest) {
            finding.suggestions.push(new Suggestion('useRecommendation', `Use \`${recommendation}\` instead.`, [repair]));
        }
        this.rules.findings.push(finding);
    }
    unary(index: number): void {
        const parser = this.rules.parser;
        const node = parser.node(index);
        const operand = this.rules.skip(node.children[0] ?? -1);
        const inner = parser.node(operand);
        if(node.operator === 'ExclamationToken' && inner.kind === 'PrefixUnaryExpression' &&
            inner.operator === 'ExclamationToken') {
            const subject = this.rules.skip(inner.children[0] ?? -1);
            if(!this.already(subject, booleanLike)) {
                const rewrite = this.narrowing(index, subject);
                const recommend = `Boolean(${this.operand(subject)})`;
                this.report(index, rewrite !== '' && rewrite !== 'Boolean' ? rewrite : recommend, true,
                    rewrite !== '' && (rewrite !== 'Boolean' || this.global(index)));
            }
        }
        if(node.operator === 'TildeToken' && inner.kind === 'CallExpression') {
            const raw = inner.children[0] ?? -1;
            const calleeIndex = this.rules.skip(raw);
            const callee = parser.node(calleeIndex);
            if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(callee.kind)) {
                const key = parser.node(callee.children[callee.children.length - 1] ?? -1);
                if((callee.kind === 'PropertyAccessExpression' || key.kind === 'StringLiteral') &&
                    ['indexOf', 'lastIndexOf'].includes(key.text)) {
                    this.report(index, `${this.rules.text(inner)} ${raw === calleeIndex && callee.optional ? '>= 0' : '!== -1'}`, false, false);
                }
            }
        }
        if(node.operator === 'PlusToken' && !this.numeric(operand) && !this.already(operand, numberLike)) {
            this.report(index, `Number(${this.operand(operand)})`, true, false);
        }
        if(node.operator === 'MinusToken' && inner.kind === 'PrefixUnaryExpression' && inner.operator === 'MinusToken') {
            const subject = this.rules.skip(inner.children[0] ?? -1);
            if(!this.numeric(subject) && !this.already(subject, numberLike | bigintLike)) {
                this.report(index, `Number(${this.operand(subject)})`, true, false);
            }
        }
    }
    binary(index: number): void {
        const parser = this.rules.parser;
        const node = parser.node(index);
        const left = node.children[0] ?? -1;
        const right = node.children[2] ?? -1;
        const operator = parser.node(node.children[1] ?? -1).kind;
        if(operator === 'PlusEqualsToken') {
            if(this.empty(right) && !this.already(left, stringLike)) {
                const text = this.rules.text(parser.node(left));
                this.report(index, `${text} = String(${text})`, true, false);
            }
            return;
        }
        if(operator === 'PlusToken' && ((this.empty(left) && !this.string(right)) ||
            (this.empty(right) && !this.string(left)))) {
            const operand = this.empty(left) ? right : left;
            if(!this.already(operand, stringLike)) {
                this.report(index, `String(${this.operand(operand)})`, true, false);
            }
        }
        if(operator === 'MinusToken' && parser.node(right).kind === 'NumericLiteral' &&
            parser.node(right).text === '0' && !this.numeric(left) && !this.already(left, numberLike)) {
            this.report(index, `Number(${this.operand(left)})`, true, false);
        }
        if(operator !== 'AsteriskToken' || ![left, right].some((part) =>
            parser.node(part).kind === 'NumericLiteral' && parser.node(part).text === '1')) {
            return;
        }
        const parentIndex = this.rules.parents[index] ?? -1;
        if(parser.node(right).kind === 'NumericLiteral' && parser.node(right).text === '1' && parentIndex >= 0) {
            const parent = parser.node(parentIndex);
            if(parent.kind === 'BinaryExpression' && parent.children[0] === index &&
                parser.node(parent.children[1] ?? -1).kind === 'SlashToken') {
                return;
            }
        }
        let operand = -1;
        for(const part of [right, left]) {
            const unwrapped = this.rules.skip(part);
            const shape = parser.node(unwrapped);
            const binary = shape.kind === 'BinaryExpression' && !['AmpersandAmpersandToken', 'BarBarToken',
                'QuestionQuestionToken', 'CommaToken'].includes(parser.node(shape.children[1] ?? -1).kind) &&
                !new Reassign(this.rules).assignment(parser.node(shape.children[1] ?? -1).kind);
            if(!binary && !this.numeric(unwrapped)) {
                operand = unwrapped;
                break;
            }
        }
        if(operand >= 0 && !this.already(operand, numberLike)) {
            this.report(index, `Number(${this.operand(operand)})`, true, false);
        }
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            if((this.rules.parents[index] ?? -1) < 0) {
                continue;
            }
            const kind = this.rules.parser.node(index).kind;
            if(kind === 'PrefixUnaryExpression') {
                this.unary(index);
            }
            if(kind === 'BinaryExpression') {
                this.binary(index);
            }
        }
    }
}
