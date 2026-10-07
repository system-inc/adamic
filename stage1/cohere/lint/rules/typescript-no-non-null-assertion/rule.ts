import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { message, suggestionMessage } from './messages.ts';
import { Edit, Suggestion } from './repairs.ts';

const assignments = ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken', 'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken', 'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'];

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    assignmentTarget(index: number): boolean {
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) { return false; }
        const node = this.context.node(parent);
        if(node.kind === 'BinaryExpression') {
            const token = node.children[1] ?? panic('binary without operator');
            return assignments.includes(this.context.node(token).kind) && node.children[0] === index;
        }
        if(node.kind === 'DeleteExpression') { return node.children[0] === index; }
        if(node.kind === 'PrefixUnaryExpression' || node.kind === 'PostfixUnaryExpression') {
            return ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator) && node.children[0] === index;
        }
        if(node.kind === 'ArrayLiteralExpression' || node.kind === 'SpreadElement') { return true; }
        if(node.kind === 'PropertyAssignment') {
            if(node.children[1] !== index) { return false; }
            const object = this.context.parents[parent] ?? -1;
            return object >= 0 && this.context.node(object).kind === 'ObjectLiteralExpression' && this.assignmentTarget(object);
        }
        if(['ParenthesizedExpression', 'NonNullExpression', 'AsExpression', 'TypeAssertionExpression', 'SatisfiesExpression'].includes(node.kind)) {
            return this.assignmentTarget(parent);
        }
        return false;
    }
    suggestions(index: number): Suggestion[] {
        const assertion = this.context.node(index);
        if(assertion.kind !== 'NonNullExpression') { return []; }
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) { return []; }
        const node = this.context.node(parent);
        if(node.children[0] !== index) { return []; }
        const member = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind);
        if(!member && node.kind !== 'CallExpression') { return []; }
        if(member && this.assignmentTarget(parent)) { return []; }
        const end = assertion.end;
        for(const child of node.children) {
            if(this.context.node(child).kind === 'QuestionDotToken') {
                return [new Suggestion('suggestOptionalChain', suggestionMessage, [new Edit(end - 1, end, '')])];
            }
        }
        if(node.kind === 'ElementAccessExpression' || node.kind === 'CallExpression') {
            return [new Suggestion('suggestOptionalChain', suggestionMessage, [new Edit(end - 1, end, '?.')])];
        }
        // Match Go's positional search, including its treatment of comments between the tokens.
        const dot = this.context.source.indexOf('.', end);
        if(dot < 0 || dot >= node.end) { return []; }
        return [new Suggestion('suggestOptionalChain', suggestionMessage, [new Edit(end - 1, end, ''), new Edit(dot, dot + 1, '?.')])];
    }
    visit(index: number, parent: number): void {
        if(this.context.node(index).kind === 'NonNullExpression') {
            this.context.report(index, '@typescript-eslint/no-non-null-assertion', 'noNonNull', message, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
