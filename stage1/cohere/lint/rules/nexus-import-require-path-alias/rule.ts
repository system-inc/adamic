import type { RuleContext } from '../../context.ts';
import { Rule as Implementation } from './implementation.ts';
export class Rule {
    readonly implementation: Implementation;
    constructor(context: RuleContext) { this.implementation = new Implementation(context, context.parser.path, undefined); }
    visit(index: number): void { this.implementation.visit(index); }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
