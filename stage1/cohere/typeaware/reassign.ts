// Default production no-param-reassign. Name resolution is a checker fact;
// declaration anchoring and every write judgment run on Adamic's parsed tree.
import type { Rules } from './rules.ts';
import { Diagnostic } from './diagnostic.ts';
import { Frames, header } from './frames.ts';

export class Reassign {
    readonly linter: Rules;
    readonly declarations = new Map<string, number>();
    constructor(linter: Rules) {
        this.linter = linter;
    }
    assignment(operator: string): boolean {
        return ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken',
            'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken',
            'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken',
            'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken',
            'BarEqualsToken', 'CaretEqualsToken', 'AmpersandAmpersandEqualsToken',
            'BarBarEqualsToken', 'QuestionQuestionEqualsToken'].includes(operator);
    }
    writes(index: number): boolean {
        let current = index;
        while((this.linter.parents[current] ?? -1) >= 0) {
            const parentIndex = this.linter.parents[current] ?? -1;
            const parent = this.linter.parser.node(parentIndex);
            if(parent.kind === 'PropertyAccessExpression' && parent.children[parent.children.length - 1] === current) {
                return false;
            }
            if(['SpreadElement', 'SpreadAssignment', 'ArrayLiteralExpression',
                'ObjectLiteralExpression', 'ParenthesizedExpression'].includes(parent.kind)) {
                current = parentIndex;
                continue;
            }
            if(parent.kind === 'PropertyAssignment') {
                if(parent.children[parent.children.length - 1] !== current) {
                    return false;
                }
                current = parentIndex;
                continue;
            }
            if(parent.kind === 'ShorthandPropertyAssignment') {
                if(parent.children[0] !== current) {
                    return false;
                }
                current = parentIndex;
                continue;
            }
            if(parent.kind === 'BinaryExpression') {
                return parent.children[0] === current &&
                    this.assignment(this.linter.parser.node(parent.children[1] ?? -1).kind);
            }
            if(parent.kind === 'PrefixUnaryExpression' || parent.kind === 'PostfixUnaryExpression') {
                return parent.operator === 'PlusPlusToken' || parent.operator === 'MinusMinusToken';
            }
            return (parent.kind === 'ForInStatement' || parent.kind === 'ForOfStatement') &&
                parent.children[0] === current;
        }
        return false;
    }
    parameter(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const kind = this.linter.parser.node(current).kind;
            if(kind === 'Parameter') {
                return true;
            }
            if(!['BindingElement', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(kind)) {
                return false;
            }
            current = this.linter.parents[current] ?? -1;
        }
        return false;
    }
    run(): void {
        const parser = this.linter.parser;
        for(let index = 0; index < parser.nodes.length; index++) {
            const node = parser.node(index);
            if(this.parameter(index)) {
                this.declarations.set(`${this.linter.byte(node.pos)}:${this.linter.byte(node.end)}:${node.kind}`, index);
            }
        }
        for(let index = 0; index < parser.nodes.length; index++) {
            const node = parser.node(index);
            if(node.kind !== 'Identifier' || !this.writes(index)) {
                continue;
            }
            const facts = new Frames(this.linter.ask(index, 'binding-declarations'));
            header(facts, 'binding-declarations');
            let anchored = false;
            if(facts.yes()) {
                facts.natural();
                const length = facts.natural();
                for(let at = 0; at < length; at++) {
                    const path = facts.field();
                    const kind = facts.field();
                    const start = facts.natural();
                    const end = facts.natural();
                    if(path === this.linter.path && this.declarations.has(`${start}:${end}:${kind}`)) {
                        anchored = true;
                    }
                }
            }
            facts.end();
            if(anchored) {
                const finding = new Diagnostic('no-param-reassign', 'assignmentToFunctionParam',
                    `Assignment to function parameter '${node.text}'.`, this.linter.byte(this.linter.start(node)),
                    this.linter.byte(node.end), '');
                this.linter.findings.push(finding);
            }
        }
    }
}
