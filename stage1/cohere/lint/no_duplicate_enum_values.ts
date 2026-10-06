import type { Batch5Context } from './batch5_context.ts';
import { initializer } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-duplicate-enum-values';
    if(context.node(index).kind !== 'EnumDeclaration' || !context.enabled(rule)) {
        return;
    }
    const numbers = new Map<string, number>();
    const strings = new Map<string, number>();
    for(const member of context.node(index).children) {
        if(context.node(member).kind !== 'EnumMember') {
            continue;
        }
        const value = initializer(context, member);
        if(value < 0) {
            continue;
        }
        const node = context.node(value);
        if(!['NumericLiteral', 'StringLiteral'].includes(node.kind)) {
            continue;
        }
        const seen = node.kind === 'NumericLiteral' ? numbers : strings;
        const previous = seen.get(node.text);
        if(previous !== undefined) {
            context.report(previous, rule, 'noDuplicateEnumValues', ruleMessage('noDuplicateEnumValues'), [], []);
        }
        if(previous === undefined || node.kind === 'StringLiteral') {
            seen.set(node.text, value);
        }
    }
}
