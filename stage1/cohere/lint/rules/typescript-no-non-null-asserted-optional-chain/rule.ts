import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { message, suggestionMessage } from './messages.ts';
import { Edit, Suggestion } from './repairs.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    reports(index: number): boolean {
        const node = this.context.node(index);
        if(node.kind !== 'NonNullExpression') { return false; }
        if(node.optional) {
            const parent = this.context.parents[index] ?? -1;
            if(parent < 0) { return true; }
            const outer = this.context.node(parent);
            const continues = ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(outer.kind);
            return !(continues && outer.optional && outer.children[0] === index);
        }
        let inner = node.children[0] ?? panic('assertion without operand');
        while(this.context.node(inner).kind === 'ParenthesizedExpression') {
            inner = this.context.node(inner).children[0] ?? panic('parenthesis without operand');
        }
        return this.context.node(inner).optional;
    }
    suggestions(index: number): Suggestion[] {
        if(!this.reports(index)) { return []; }
        const end = this.context.node(index).end;
        return [new Suggestion('suggestRemovingNonNull', suggestionMessage, [new Edit(end - 1, end, '')])];
    }
    visit(index: number, parent: number): void {
        if(this.reports(index)) {
            // Shared Finding cannot yet express a repair span separate from its diagnostic span.
            this.context.report(index, '@typescript-eslint/no-non-null-asserted-optional-chain', 'noNonNullOptionalChain', message, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
