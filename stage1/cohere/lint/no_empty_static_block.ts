// Port of pinned Go cohere's no-empty-static-block.
import type { RuleContext } from './rule_context.ts';

const messageUnexpectedEmptyStaticBlock =
    'This static initialization block is empty, so it runs nothing at class definition time and exists only as syntax. Unlike an empty catch, there is no reading where that is deliberate: a static block has no parameter to bind and no error to swallow, so an empty one is a body that was deleted or never written. Remove it, or say in a comment why the class needs an initialization step that does nothing.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-empty-static-block')) return;
    if(ctx.node(index).kind !== 'ClassStaticBlockDeclaration') return;
    const body = ctx.node(index).children[0] ?? -1;
    if(body < 0 || ctx.node(body).kind !== 'Block' || ctx.node(body).children.length !== 0 || ctx.comment(body)) return;
    ctx.report(index, 'no-empty-static-block', 'unexpectedEmptyStaticBlock', messageUnexpectedEmptyStaticBlock);
}
