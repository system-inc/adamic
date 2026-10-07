import type { RuleContext } from '../../context.ts';
import { Rule as Implementation } from './implementation.ts';
export class Rule {
    readonly implementation: Implementation;
    constructor(context: RuleContext, rawOptions: string = context.settings.text) { this.implementation = new Implementation(context, context.parser.path, undefined, rawOptions); }
    visit(index: number): void { this.implementation.visit(index); }
}
export function create(context: RuleContext, rawOptions: string = context.settings.text): Rule { return new Rule(context, rawOptions); }
