// Port of pinned Go cohere's no-iterator.
import type { RuleContext } from './rule_context.ts';

const messageNoIterator =
    "This uses the reserved name `__iterator__`. That was SpiderMonkey's pre-standard iteration protocol, removed from Firefox in 2015 and never implemented anywhere else, so the property does nothing in any engine running today. Use `[Symbol.iterator]`, which is the standard protocol that replaced it.";
const messageUseSymbolIterator = 'Replace `__iterator__` with `[Symbol.iterator]`.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-iterator')) return;
    if(ctx.accessName(index) !== '__iterator__') return;
    const object = ctx.node(index).children[0] ?? -1;
    const finding = ctx.report(
        index,
        'no-iterator',
        'noIterator',
        messageNoIterator,
        'suggestion',
        '[Symbol.iterator]',
        messageUseSymbolIterator,
    );
    finding.editStart = ctx.node(object).end;
}
