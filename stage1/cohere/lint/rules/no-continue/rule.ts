import type { RuleContext } from '../../context.ts';
import { messageContinue } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        const node = this.context.node(index);
        if(node.kind === 'ContinueStatement') {
            this.context.report(index, 'no-continue', 'unexpected', messageContinue, '', '', '');
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
