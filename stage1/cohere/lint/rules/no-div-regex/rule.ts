import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { SuggestionEdit } from '../../suggestions.a';
import { messageDivRegex } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const start = this.context.start(index);
        if(node.end - start >= 2 && this.context.source[start + 1] === '=') {
            // The repair replaces the equals sign alone, inside the reported literal.
            this.context.reportRange(start, node.end, 'no-div-regex', 'unexpected', messageDivRegex, [
                new SuggestionEdit(start + 1, start + 2, '[=]'),
            ]);
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
