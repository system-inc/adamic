import type { RuleContext } from '../../context.ts';
import { messageWith } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        const node = this.context.node(index);
        if(node.kind === 'WithStatement') {
            // Upstream underlines the `with` keyword alone, not the statement.
            const start = this.context.start(index);
            this.context.reportRange(start, start + 4, 'no-with', 'noWith', messageWith);
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
