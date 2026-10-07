import type { RuleContext } from '../../context.ts';
import { Edit, Suggestion, record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { message, suggestion } from './messages.ts';
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    assignmentTarget(index: number): boolean {
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) { return false; }
        const node = this.context.node(parent);
        switch(node.kind) {
            case 'BinaryExpression':
                return ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken', 'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken', 'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'].includes(this.context.node(node.children[1] ?? -1).kind) && node.children[0] === index;
            case 'DeleteExpression':
                return node.children[0] === index;
            case 'PrefixUnaryExpression':
            case 'PostfixUnaryExpression':
                return ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator) && node.children[0] === index;
            case 'ArrayLiteralExpression':
            case 'SpreadElement':
                return true;
            case 'PropertyAssignment': {
                if(node.children[node.children.length - 1] !== index) { return false; }
                const object = this.context.parents[parent] ?? -1;
                return object >= 0 && this.context.node(object).kind === 'ObjectLiteralExpression' && this.assignmentTarget(object);
            }
            case 'ParenthesizedExpression':
            case 'NonNullExpression':
            case 'AsExpression':
            case 'TypeAssertionExpression':
            case 'SatisfiesExpression':
                return this.assignmentTarget(parent);
            default:
                return false;
        }
    }
    edits(index: number, parent: number): Edit[] {
        if(parent < 0) { return []; }
        const node = this.context.node(index);
        const outer = this.context.node(parent);
        const member = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(outer.kind);
        if(!member && outer.kind !== 'CallExpression') { return []; }
        if(outer.children[0] !== index) { return []; }
        if(member && this.assignmentTarget(parent)) { return []; }
        // optional is chain membership; a QuestionDotToken is this link's token.
        for(const child of outer.children) {
            if(this.context.node(child).kind === 'QuestionDotToken') { return [new Edit(node.end - 1, node.end, '')]; }
        }
        if(outer.kind !== 'PropertyAccessExpression') { return [new Edit(node.end - 1, node.end, '?.')]; }
        // Match Go's actual first-dot scan, including any intervening comments.
        const dot = this.context.source.indexOf('.', node.end);
        if(dot < 0 || dot >= outer.end) { return []; }
        return [new Edit(node.end - 1, node.end, ''), new Edit(dot, dot + 1, '?.')];
    }
    visit(index: number, parent: number): void {
        if(this.context.node(index).kind !== 'NonNullExpression') { return; }
        const edits = this.edits(index, parent);
        const suggestions: Suggestion[] = [];
        if(edits.length > 0) { suggestions.push(new Suggestion('suggestOptionalChain', suggestion, edits)); }
        record(this.context, index, '@typescript-eslint/no-non-null-assertion', 'noNonNull', message, suggestions);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
