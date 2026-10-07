import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageContinue } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // Every `continue` reports, so the node needs no test beyond the kind the driver dispatched on.
    visit(node: ParseNode, index: number): void {
        this.context.reportRange(this.context.start(index), node.end, 'no-continue', 'unexpected', messageContinue);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
