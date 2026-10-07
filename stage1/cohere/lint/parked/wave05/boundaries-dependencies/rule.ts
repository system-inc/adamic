import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Engine } from './decision.a';
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number, parent: number): void {
        if(this.context.settings.text === '' || this.context.settings.text === 'null') { return; }
        panic('NotYet boundaries/dependencies: RuleContext has no Program/module-resolution facts or config/project root; decoded element/policy engine is available in this rule directory');
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
