import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageNew } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const expression = node.children[0] ?? -1;
        // Our tree keeps the parentheses ESTree discards, so `(new Date());` is seen through them, as upstream does.
        if(expression >= 0 && this.context.node(this.context.unwrap(expression)).kind === 'NewExpression') {
            this.context.report(index, 'no-new', 'noNewStatement', messageNew, '', '', '');
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
