import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number): void {
        let name = -1;
        for(const child of this.context.node(index).children) {
            if(this.context.node(child).kind === 'ExclamationToken') {
                if(name >= 0) { this.context.report(name, 'adamic/no-definite-assignment', 'definiteAssignment', message, '', '', ''); }
                return;
            }
            // Modifiers precede the name, and the assertion immediately follows it.
            name = child;
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
