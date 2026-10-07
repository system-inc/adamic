import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { duplicateMessage } from './messages.ts';

function duplicateCases(context: RuleContext, index: number): void {
    const seen = new Set<string>();
    for(const clause of context.node(index).children) {
        if(context.node(clause).kind !== 'CaseClause') {
            continue;
        }
        const expression = context.node(clause).children[0] ?? panic('case without expression');
        const signature = context.signature(expression);
        if(seen.has(signature)) {
            context.report(expression, 'no-duplicate-case', 'unexpected', duplicateMessage, '', '', '');
        }
        seen.add(signature);
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        for(const child of this.context.node(index).children) {
            if(this.context.node(child).kind === 'CaseBlock') {
                duplicateCases(this.context, child);
            }
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
