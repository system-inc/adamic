// Port of pinned Go cohere's no-multi-assign.
import type { RuleContext } from './rule_context.ts';

const messageNoMultiAssign =
    'This chains one assignment inside another, so a single statement writes to two names and only one of them is where a reader looks. In `let a = b = c` the `b` is not declared at all, so a chain inside a declaration silently creates or writes an outer binding. Write each assignment on its own line.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-multi-assign')) return;
    if(!ctx.assignment(index)) return;
    let outer = index;
    let parent = ctx.parent(outer);
    while(parent >= 0 && ctx.node(parent).kind === 'ParenthesizedExpression') {
        outer = parent;
        parent = ctx.parent(outer);
    }
    if(parent < 0) return;
    const kind = ctx.node(parent).kind;
    const chained =
        kind === 'VariableDeclaration' || kind === 'PropertyDeclaration'
            ? ctx.initializer(parent) === outer
            : ctx.settings.read('ignorenondeclaration', 'false') !== 'true' &&
              ctx.assignment(parent) &&
              ctx.node(parent).children[2] === outer;
    if(chained) ctx.report(index, 'no-multi-assign', 'unexpectedChain', messageNoMultiAssign);
}
