// Port of pinned Go cohere's nexus/consistency-no-stuttering-name.
import type { RuleContext } from './rule_context.ts';
import { child, last, kind, text } from './batch4_helpers.ts';
const defaultNames = ['outcome', 'result', 'data', 'value', 'output', 'response', 'state', 'item', 'thing'];
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('nexus/consistency-no-stuttering-name') || kind(ctx, index) !== 'PropertyAccessExpression') return;
    const receiver = child(ctx, index, 0);
    const property = last(ctx, index);
    if(
        kind(ctx, receiver) !== 'Identifier' ||
        kind(ctx, property) !== 'Identifier' ||
        ctx.node(receiver).text !== ctx.node(property).text
    )
        return;
    const configured = ctx.settings.list('genericnames', defaultNames);
    const names = configured.length === 0 ? defaultNames : configured;
    if(!names.includes(ctx.node(property).text)) return;
    ctx.report(
        property,
        'nexus/consistency-no-stuttering-name',
        'stutteringName',
        text('nexus_consistency_no_stuttering_name:stutteringName').replaceAll('{{name}}', ctx.node(property).text),
    );
}
