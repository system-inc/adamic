import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { description, suggestionDescription } from './messages.ts';
import { Edits } from './edits.ts';

const assignments = ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken', 'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken', 'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'];
const wrappers = ['ParenthesizedExpression', 'NonNullExpression', 'AsExpression', 'TypeAssertionExpression', 'SatisfiesExpression'];

export class Rule {
    readonly context: RuleContext;
    readonly edits: Edits;
    constructor(context: RuleContext) { this.context = context; this.edits = new Edits(context.source); }
    assignmentTarget(index: number): boolean {
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) { return false; }
        const node = this.context.node(parent);
        const first = node.children[0] ?? -1;
        if(node.kind === 'BinaryExpression') {
            const operator = node.children[1] ?? panic('binary without operator');
            return assignments.includes(this.context.node(operator).kind) && first === index;
        }
        if(node.kind === 'DeleteExpression') { return first === index; }
        if(node.kind === 'PrefixUnaryExpression' || node.kind === 'PostfixUnaryExpression') {
            return ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator) && first === index;
        }
        if(node.kind === 'ArrayLiteralExpression' || node.kind === 'SpreadElement') { return true; }
        if(node.kind === 'PropertyAssignment') {
            if(node.children[1] !== index) { return false; }
            const object = this.context.parents[parent] ?? -1;
            return object >= 0 && this.context.node(object).kind === 'ObjectLiteralExpression' && this.assignmentTarget(object);
        }
        if(wrappers.includes(node.kind)) { return this.assignmentTarget(parent); }
        return false;
    }
    suggestion(index: number, parent: number): string {
        if(parent < 0) { return ''; }
        const node = this.context.node(parent);
        if(!['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(node.kind) || node.children[0] !== index) { return ''; }
        if(node.kind !== 'CallExpression' && this.assignmentTarget(parent)) { return ''; }
        const end = this.context.node(index).end;
        const optional = node.children.some(child => this.context.node(child).kind === 'QuestionDotToken');
        if(optional) { return this.edits.range(end - 1, end, ''); }
        if(node.kind !== 'PropertyAccessExpression') { return this.edits.range(end - 1, end, '?.'); }
        // Go deliberately scans raw bytes for the first dot, including one inside a comment.
        const dot = this.context.source.indexOf('.', end);
        if(dot < 0 || dot >= node.end) { return ''; }
        return `${this.edits.range(end - 1, end, '')}|${this.edits.range(dot, dot + 1, '?.')}`;
    }
    visit(index: number, parent: number): void {
        const suggestion = this.suggestion(index, parent);
        this.context.report(index, '@typescript-eslint/no-non-null-assertion', 'noNonNull', description,
            suggestion === '' ? '' : 'suggestion-edits:suggestOptionalChain', suggestion, suggestion === '' ? '' : suggestionDescription);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
