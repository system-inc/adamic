import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageTemplateCurly } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        let from = 0;
        for(;;) {
            const opening = node.text.indexOf('${', from);
            if(opening < 0) {
                return;
            }
            const body = opening + 2;
            if(node.text.indexOf('}', body) > body) {
                this.context.report(
                    index,
                    'no-template-curly-in-string',
                    'unexpectedTemplateExpression',
                    messageTemplateCurly,
                    '',
                    '',
                    '',
                );
                return;
            }
            from = body;
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
