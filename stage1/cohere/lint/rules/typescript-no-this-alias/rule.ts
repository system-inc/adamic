import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { assignmentMessage, destructureMessage } from './messages.ts';
import type { Suggestion } from './repairs.ts';

const assignments = ['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken', 'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken', 'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'];

export class Rule {
    readonly context: RuleContext;
    readonly allowed: string[] = [];
    readonly reportDestructuring: boolean = false;
    constructor(context: RuleContext) {
        this.context = context;
        // Captured Go options are already decoded; also accept the public spellings.
        this.allowed = context.settings.list('allowednames', context.settings.list('allownames', []));
        this.reportDestructuring = context.settings.read('reportdestructuring', context.settings.read('allowdestructuring', 'true') === 'false' ? 'true' : 'false') === 'true';
    }
    unwrap(index: number): number {
        let cursor = index;
        while(true) {
            const node = this.context.node(cursor);
            if(node.kind === 'TypeAssertionExpression') {
                cursor = node.children[1] ?? panic('type assertion without operand');
            }
            else if(['ParenthesizedExpression', 'AsExpression', 'SatisfiesExpression', 'NonNullExpression'].includes(node.kind)) {
                cursor = node.children[0] ?? panic('wrapper without operand');
            }
            else { return cursor; }
        }
    }
    suggestions(index: number): Suggestion[] { return []; }
    visit(index: number, parent: number): void {
        const path = this.context.parser.path;
        if(!['.ts', '.tsx', '.mts', '.cts'].some(suffix => path.endsWith(suffix))) { return; }
        const node = this.context.node(index);
        let target = -1;
        let initializer = -1;
        let destructure = false;
        if(node.kind === 'VariableDeclaration') {
            target = node.children[0] ?? -1;
            initializer = node.children[node.children.length - 1] ?? -1;
            destructure = target >= 0 && this.context.node(target).kind !== 'Identifier';
        }
        else if(node.kind === 'BinaryExpression') {
            const token = node.children[1] ?? panic('binary without operator');
            if(!assignments.includes(this.context.node(token).kind)) { return; }
            target = node.children[0] ?? -1;
            initializer = node.children[2] ?? -1;
            destructure = target >= 0 && ['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(this.context.node(target).kind);
            if(!destructure && target >= 0) { target = this.unwrap(target); }
        }
        if(target < 0 || initializer < 0 || this.context.node(initializer).kind !== 'ThisKeyword') { return; }
        if(destructure) {
            if(this.reportDestructuring) {
                this.context.report(target, '@typescript-eslint/no-this-alias', 'thisDestructure', destructureMessage, '', '', '');
            }
        }
        else if(this.context.node(target).kind === 'Identifier' && !this.allowed.includes(this.context.node(target).text)) {
            this.context.report(target, '@typescript-eslint/no-this-alias', 'thisAssignment', assignmentMessage, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
