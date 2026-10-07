import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { description, suggestionDescription } from './messages.ts';
import { Edits } from './edits.ts';

export class Rule {
    readonly context: RuleContext;
    readonly edits: Edits;
    constructor(context: RuleContext) { this.context = context; this.edits = new Edits(context.source); }
    optional(index: number): boolean {
        const node = this.context.node(index);
        return node.optional && ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(node.kind);
    }
    root(index: number): boolean {
        const node = this.context.node(index);
        return this.optional(index) && node.kind !== 'NonNullExpression' && node.children.some(child => this.context.node(child).kind === 'QuestionDotToken');
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        let reports = false;
        if(this.optional(index)) {
            reports = parent < 0 || !this.optional(parent) || this.root(parent) || this.context.node(parent).children[0] !== index;
        }
        else {
            const operand = node.children[0] ?? panic('non-null without operand');
            reports = this.optional(this.context.unwrap(operand));
        }
        if(!reports) { return; }
        this.context.report(index, '@typescript-eslint/no-non-null-asserted-optional-chain', 'noNonNullOptionalChain', description,
            'suggestion-edits:suggestRemovingNonNull', this.edits.range(node.end - 1, node.end, ''), suggestionDescription);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
