import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageConsistencyNoConsole } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const receiver = this.context.unwrap(node.children[0] ?? panic('receiver'));
        const key = node.children[node.children.length - 1] ?? panic('property');
        if(this.context.node(receiver).kind === 'Identifier' && this.context.node(receiver).text === 'console') {
            this.context.report(
                index,
                'base/consistency-no-console',
                'consistencyNoConsole',
                messageConsistencyNoConsole(
                    this.context.node(key).kind === 'Identifier' ? this.context.node(key).text : 'log',
                ),
                '',
                '',
                '',
            );
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
