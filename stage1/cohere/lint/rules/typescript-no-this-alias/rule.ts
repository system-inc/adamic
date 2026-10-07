import { syntaxKinds as listenerKinds } from './listeners.a';
export const syntaxKinds: readonly number[] = listenerKinds;
import type { RuleContext } from '../../context.ts';
import { assignment, destructure } from './messages.a';
const name = '@typescript-eslint/no-this-alias';
const assignments: readonly string[] = ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken',
    'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken',
    'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken',
    'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'];
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    target(index: number): number {
        let cursor = index;
        while(['ParenthesizedExpression', 'AsExpression', 'SatisfiesExpression', 'TypeAssertionExpression', 'NonNullExpression'].includes(this.context.node(cursor).kind)) {
            const node = this.context.node(cursor);
            const child = node.children[node.kind === 'TypeAssertionExpression' ? 1 : 0] ?? -1;
            if(child < 0) return cursor;
            cursor = child;
        }
        return cursor;
    }
    visit(index: number): void {
        const context = this.context;
        if(!['.ts', '.tsx', '.mts', '.cts'].some((extension) => context.parser.path.endsWith(extension))) return;
        const node = context.node(index);
        let target = -1;
        if(node.kind === 'VariableDeclaration') {
            const initializer = node.children[node.children.length - 1] ?? -1;
            if(node.children.length < 2 || initializer < 0 || context.node(initializer).kind !== 'ThisKeyword') return;
            target = node.children[0] ?? -1;
        }
        else {
            const right = node.children[2] ?? -1;
            const operator = node.children[1] ?? -1;
            if(operator < 0 || !assignments.includes(context.node(operator).kind) || right < 0 || context.node(right).kind !== 'ThisKeyword') return;
            target = node.children[0] ?? -1;
        }
        if(target < 0) return;
        const kind = context.node(target).kind;
        const pattern = ['ObjectBindingPattern', 'ArrayBindingPattern', 'ObjectLiteralExpression', 'ArrayLiteralExpression'].includes(kind);
        const report = context.settings.read('reportdestructuring', context.settings.read('allowdestructuring', 'true') === 'false' ? 'true' : 'false') === 'true';
        if(pattern) {
            if(!report) return;
            context.report(target, name, 'thisDestructure', destructure, '', '', '');
            return;
        }
        target = this.target(target);
        if(context.node(target).kind !== 'Identifier') return;
        const allowed = context.settings.list('allowednames', context.settings.list('allownames', []));
        if(allowed.includes(context.node(target).text)) return;
        context.report(target, name, 'thisAssignment', assignment, '', '', '');
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
