import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Edit, Suggestion, record } from './diagnostic.ts';
import { message, suggestion } from './messages.ts';
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind !== 'NonNullExpression') { return; }
        if(node.optional) {
            // Go marks assertions reached through by a later chain link.
            if(parent >= 0) {
                const outer = this.context.node(parent);
                let root = false;
                for(const child of outer.children) {
                    if(this.context.node(child).kind === 'QuestionDotToken') { root = true; }
                }
                if(outer.optional && !root && outer.children[0] === index && ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(outer.kind)) { return; }
            }
        }
        else {
            const operand = node.children[0] ?? panic('assertion without expression');
            const inner = this.context.node(this.context.unwrap(operand));
            if(!inner.optional || !['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression', 'NonNullExpression'].includes(inner.kind)) { return; }
        }
        record(this.context, index, '@typescript-eslint/no-non-null-asserted-optional-chain', 'noNonNullOptionalChain', message,
            [new Suggestion('suggestRemovingNonNull', suggestion, [new Edit(node.end - 1, node.end, '')])]);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
