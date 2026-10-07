import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { assignment, destructure } from './messages.ts';
export class Rule {
    readonly context: RuleContext;
    readonly reportDestructuring: boolean;
    readonly allowedNames: string[];
    constructor(context: RuleContext) {
        this.context = context;
        // Captured options are Go's decoded struct. Wire configurations are
        // accepted too, with the same inversion and correct-spelling preference.
        this.reportDestructuring = context.settings.read('reportdestructuring', context.settings.read('allowdestructuring', 'true') === 'false' ? 'true' : 'false') === 'true';
        this.allowedNames = context.settings.list('allowednames', context.settings.list('allownames', []));
    }
    identifier(index: number): void {
        if(this.allowedNames.includes(this.context.node(index).text)) { return; }
        record(this.context, index, '@typescript-eslint/no-this-alias', 'thisAssignment', assignment, []);
    }
    visit(index: number, parent: number): void {
        const path = this.context.parser.path;
        if(!['.ts', '.tsx', '.mts', '.cts'].some((suffix) => path.endsWith(suffix))) { return; }
        const node = this.context.node(index);
        if(node.kind === 'VariableDeclaration') {
            const initializer = node.children[node.children.length - 1] ?? -1;
            if(initializer < 0 || this.context.node(initializer).kind !== 'ThisKeyword') { return; }
            const target = node.children[0] ?? panic('missing declaration name');
            this.context.scanner.pos = this.context.node(target).end;
            let initialized = false;
            while(this.context.scanner.scan() !== 'EndOfFile' && this.context.scanner.start < this.context.node(initializer).pos) {
                if(this.context.scanner.kind === 'EqualsToken') { initialized = true; }
            }
            if(!initialized) { return; }
            if(this.context.node(target).kind === 'Identifier') { this.identifier(target); }
            else if(this.reportDestructuring) { record(this.context, target, '@typescript-eslint/no-this-alias', 'thisDestructure', destructure, []); }
            return;
        }
        if(node.kind !== 'BinaryExpression' || !['EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken', 'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken', 'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken', 'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken', 'BarEqualsToken', 'CaretEqualsToken', 'BarBarEqualsToken', 'AmpersandAmpersandEqualsToken', 'QuestionQuestionEqualsToken'].includes(this.context.node(node.children[1] ?? -1).kind)) { return; }
        const rhs = node.children[node.children.length - 1] ?? panic('missing assignment value');
        if(this.context.node(rhs).kind !== 'ThisKeyword') { return; }
        let target = node.children[0] ?? panic('missing assignment target');
        if(['ArrayLiteralExpression', 'ObjectLiteralExpression'].includes(this.context.node(target).kind)) {
            if(this.reportDestructuring) { record(this.context, target, '@typescript-eslint/no-this-alias', 'thisDestructure', destructure, []); }
            return;
        }
        while(['ParenthesizedExpression', 'AsExpression', 'SatisfiesExpression', 'TypeAssertionExpression', 'NonNullExpression'].includes(this.context.node(target).kind)) {
            const current = this.context.node(target);
            target = current.children[current.kind === 'TypeAssertionExpression' ? current.children.length - 1 : 0] ?? panic('missing wrapped target');
        }
        if(this.context.node(target).kind === 'Identifier') { this.identifier(target); }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
