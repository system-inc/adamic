// Port of pinned Go cohere's max-nested-callbacks.
import type { RuleContext } from './rule_context.ts';
import { integer, text, child, kind, range } from './batch4_helpers.ts';
function callback(ctx: RuleContext, index: number): boolean {
    if(!['FunctionExpression', 'ArrowFunction'].includes(kind(ctx, index))) return false;
    const parent = ctx.parent(index);
    return (
        parent >= 0 &&
        child(ctx, parent, 0) !== index &&
        (kind(ctx, parent) === 'CallExpression' ||
            (kind(ctx, parent) === 'NewExpression' &&
                ctx.settings.read('checkconstructorcallcallbacks', 'false') === 'true'))
    );
}
function checkNode(ctx: RuleContext, index: number): void {
    const rule = 'max-nested-callbacks';
    if(!callback(ctx, index)) return;
    let depth = 0;
    let current = index;
    while(current >= 0) {
        if(callback(ctx, current)) depth++;
        current = ctx.parent(current);
    }
    const maximum = integer(ctx.settings.read('maximum', '10'));
    if(depth <= maximum) return;
    let start = ctx.start(index);
    let end = ctx.node(index).end;
    if(kind(ctx, index) === 'ArrowFunction') {
        for(const part of ctx.node(index).children)
            if(kind(ctx, part) === 'EqualsGreaterThanToken') {
                start = ctx.start(part);
                end = ctx.node(part).end;
            }
    }
    else {
        let cursor = start;
        for(const part of ctx.node(index).children) {
            if(kind(ctx, part) === 'TypeParameter') cursor = ctx.node(part).end;
            if(kind(ctx, part) === 'Parameter') {
                end = ctx.node(part).pos - 1;
                break;
            }
        }
        let token = ctx.scanAt(cursor);
        while(token !== 'EndOfFile' && ctx.scanStart() < end) {
            if(token === 'OpenParenToken') {
                end = ctx.scanStart();
                break;
            }
            token = ctx.scanner.scan();
        }
    }
    range(
        ctx,
        rule,
        'exceed',
        text('max_nested_callbacks').replace('%d', `${depth}`).replace('%d', `${maximum}`),
        start,
        end,
    );
}
function walk(ctx: RuleContext, index: number): void {
    checkNode(ctx, index);
    for(const part of ctx.node(index).children) walk(ctx, part);
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('max-nested-callbacks') || kind(ctx, index) !== 'SourceFile') return;
    walk(ctx, index);
}
