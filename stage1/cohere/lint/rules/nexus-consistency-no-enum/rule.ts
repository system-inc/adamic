import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageNoEnum } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // The finding is the enum's name, or the whole declaration when it has none.
    visit(node: ParseNode, index: number): void {
        const name = this.context.name(index);
        this.context.report(name < 0 ? index : name, 'nexus/consistency-no-enum', 'noEnum', messageNoEnum, '', '', '');
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
