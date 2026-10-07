import { syntaxKinds as listenerKinds } from './listeners.a';
export const syntaxKinds: readonly number[] = listenerKinds;
import type { RuleContext } from '../../context.ts';
import { Fix, Suggestion, SuggestedFinding, requireDetailedReporting } from '../typescript-no-non-null-assertion/suggestions.a';
import { message, suggestion } from './messages.a';
const name = '@typescript-eslint/no-non-null-asserted-optional-chain';
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    asserts(index: number): boolean {
        const context = this.context;
        const node = context.node(index);
        if(node.optional) {
            const parent = context.parents[index] ?? -1;
            if(parent < 0) return true;
            const outer = context.node(parent);
            const root = outer.optional && outer.kind !== 'NonNullExpression' && outer.children.some((child) => context.node(child).kind === 'QuestionDotToken');
            return !outer.optional || root || outer.children[0] !== index;
        }
        const expression = node.children[0] ?? -1;
        return expression >= 0 && context.node(context.unwrap(expression)).optional;
    }
    visit(index: number): void {
        if(!this.asserts(index)) return;
        const context = this.context;
        const node = context.node(index);
        const fixes = [new Fix(node.end - 1, node.end, '')];
        const suggestions = [new Suggestion('suggestRemovingNonNull', suggestion, fixes)];
        context.findings.push(new SuggestedFinding(name, 'noNonNullOptionalChain', message, context.start(index), node.end, suggestions));
    }
}
export function create(context: RuleContext): Rule {
    requireDetailedReporting(context, name);
    return new Rule(context);
}
