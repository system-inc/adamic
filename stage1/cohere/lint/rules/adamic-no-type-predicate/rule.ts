import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageTypePredicate } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // Every written type predicate reports; the driver only dispatches TypePredicate nodes here.
    visit(node: ParseNode, index: number): void {
        this.context.report(index, 'adamic/no-type-predicate', 'typePredicate', messageTypePredicate, '', '', '');
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
