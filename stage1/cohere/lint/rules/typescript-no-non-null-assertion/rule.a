import type { RuleContext } from '../../context.ts';
import { Fix, Suggestion, SuggestedFinding, requireDetailedReporting } from './suggestions.a';
import { message, suggestion } from './messages.a';
const name = '@typescript-eslint/no-non-null-assertion';
const assignments: readonly string[] = ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken',
    'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken',
    'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken',
    'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'];
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    assignmentTarget(index: number): boolean {
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) return false;
        const node = this.context.node(parent);
        const first = node.children[0] ?? -1;
        if(node.kind === 'BinaryExpression') {
            const operator = node.children[1] ?? -1;
            return first === index && operator >= 0 && assignments.includes(this.context.node(operator).kind);
        }
        if(node.kind === 'DeleteExpression') return first === index;
        if(['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(node.kind)) {
            return first === index && ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator);
        }
        if(['ArrayLiteralExpression', 'SpreadElement'].includes(node.kind)) return true;
        if(node.kind === 'PropertyAssignment') {
            if(node.children[1] !== index) return false;
            const object = this.context.parents[parent] ?? -1;
            return object >= 0 && this.context.node(object).kind === 'ObjectLiteralExpression' && this.assignmentTarget(object);
        }
        if(['ParenthesizedExpression', 'NonNullExpression', 'AsExpression', 'TypeAssertionExpression', 'SatisfiesExpression'].includes(node.kind)) {
            return this.assignmentTarget(parent);
        }
        return false;
    }
    visit(index: number): void {
        const context = this.context;
        const node = context.node(index);
        const parent = context.parents[index] ?? -1;
        const fixes: Fix[] = [];
        if(parent >= 0) {
            const outer = context.node(parent);
            const member = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(outer.kind);
            if((member || outer.kind === 'CallExpression') && outer.children[0] === index && (!member || !this.assignmentTarget(parent))) {
                const operator = node.end - 1;
                const question = outer.children.some((child) => context.node(child).kind === 'QuestionDotToken');
                if(question) fixes.push(new Fix(operator, node.end, ''));
                else if(outer.kind !== 'PropertyAccessExpression') fixes.push(new Fix(operator, node.end, '?.'));
                else {
                    const dot = context.source.indexOf('.', node.end);
                    if(dot >= node.end && dot < outer.end) {
                        fixes.push(new Fix(operator, node.end, ''));
                        fixes.push(new Fix(dot, dot + 1, '?.'));
                    }
                }
            }
        }
        const suggestions: Suggestion[] = [];
        if(fixes.length > 0) suggestions.push(new Suggestion('suggestOptionalChain', suggestion, fixes));
        context.findings.push(new SuggestedFinding(name, 'noNonNull', message, context.start(index), node.end, suggestions));
    }
}
export function create(context: RuleContext): Rule {
    requireDetailedReporting(context, name);
    return new Rule(context);
}
