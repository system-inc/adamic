import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageNoNegatedCondition } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // A conditional, or an if with an else that is not an else-if, whose test is negated.
    visit(node: ParseNode, index: number): void {
        const test = node.children[0] ?? panic('condition');
        const alternate = node.children[2] ?? -1;
        if(
            (node.kind === 'ConditionalExpression' ||
                (alternate >= 0 && this.context.node(alternate).kind !== 'IfStatement')) &&
            this.negated(test)
        ) {
            this.context.report(
                index,
                'no-negated-condition',
                'unexpectedNegated',
                messageNoNegatedCondition,
                '',
                '',
                '',
            );
        }
    }
    negated(index: number): boolean {
        const node = this.context.node(this.context.unwrap(index));
        return node.kind === 'PrefixUnaryExpression'
            ? node.operator === 'ExclamationToken'
            : node.kind === 'BinaryExpression' &&
                  ['ExclamationEqualsToken', 'ExclamationEqualsEqualsToken'].includes(
                      this.context.node(node.children[1] ?? panic('negated operator')).kind,
                  );
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
