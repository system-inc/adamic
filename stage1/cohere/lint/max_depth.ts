// Port of pinned Go cohere's max-depth.
import type { RuleContext } from './rule_context.ts';
import { integer, text, child, kind, range } from './batch4_helpers.ts';
function counted(ctx: RuleContext, index: number): boolean {
    if(kind(ctx, index) === 'IfStatement') {
        const parent = ctx.parent(index);
        return !(kind(ctx, parent) === 'IfStatement' && child(ctx, parent, 2) === index);
    }
    return [
        'SwitchStatement',
        'TryStatement',
        'DoStatement',
        'WhileStatement',
        'WithStatement',
        'ForStatement',
        'ForInStatement',
        'ForOfStatement',
    ].includes(kind(ctx, index));
}
function checkNode(ctx: RuleContext, index: number): void {
    const rule = 'max-depth';
    if(ctx.settings.read('disabled', 'false') === 'true' || !counted(ctx, index)) return;
    let depth = 0;
    let current = index;
    while(current >= 0) {
        if(ctx.functionLike(current) || kind(ctx, current) === 'ClassStaticBlockDeclaration') break;
        if(counted(ctx, current)) depth++;
        current = ctx.parent(current);
    }
    const maximum = integer(ctx.settings.read('maximum', '4'));
    if(depth <= maximum) return;
    const start = ctx.start(index);
    ctx.scanAt(start);
    range(
        ctx,
        rule,
        'tooDeeply',
        `Blocks are nested too deeply (${depth}). Maximum allowed is ${maximum}. ${text('max_depth')}`,
        start,
        ctx.scanEnd(),
    );
}
function walk(ctx: RuleContext, index: number): void {
    checkNode(ctx, index);
    for(const part of ctx.node(index).children) walk(ctx, part);
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('max-depth') || kind(ctx, index) !== 'SourceFile') return;
    walk(ctx, index);
}
