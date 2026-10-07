import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageWith } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        // Upstream underlines the `with` keyword alone, not the statement up to node.end.
        const start = this.context.start(index);
        this.context.reportRange(start, start + 4, 'no-with', 'noWith', messageWith);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
