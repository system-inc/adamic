// Port of pinned Go cohere's no-unsafe-finally.
import type { RuleContext } from './rule_context.ts';

const messageUnsafeFinallyReturn =
    'A `return` inside `finally` overrides whatever the `try` or `catch` was about to do. JavaScript suspends a pending return or throw until the finally block completes, so this statement replaces the value that was on its way out and discards a pending exception entirely, with nothing at the call site to show an error was swallowed. Set a variable here and return it after the try statement instead.';
const messageUnsafeFinallyThrow =
    'A `throw` inside `finally` replaces whatever the `try` or `catch` was about to do. A pending return never happens and a pending exception is lost, so the error that actually caused the failure is replaced by this one and the original stack is gone. Throw after the try statement, or attach the original as a cause.';
const messageUnsafeFinallyBreak =
    'A `break` inside `finally` leaves the try statement before the pending return or throw can complete, so the value is discarded and the exception never surfaces. Execution continues after the loop as though nothing had failed.';
const messageUnsafeFinallyContinue =
    'A `continue` inside `finally` leaves the try statement before the pending return or throw can complete, so the value is discarded and the exception never surfaces. The loop advances to its next iteration as though nothing had failed.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-unsafe-finally')) return;
    const node = ctx.node(index);
    if(!['ReturnStatement', 'ThrowStatement', 'BreakStatement', 'ContinueStatement'].includes(node.kind)) return;
    const label = node.kind === 'BreakStatement' || node.kind === 'ContinueStatement' ? (node.children[0] ?? -1) : -1;
    for(let parent = ctx.parent(index); parent >= 0; parent = ctx.parent(parent)) {
        const owner = ctx.node(parent);
        if(ctx.functionLike(parent) || owner.kind === 'ClassDeclaration' || owner.kind === 'ClassExpression') return;
        if((node.kind === 'ContinueStatement' || (node.kind === 'BreakStatement' && label < 0)) && ctx.loop(parent))
            return;
        if(node.kind === 'BreakStatement' && label < 0 && owner.kind === 'SwitchStatement') return;
        if(
            label >= 0 &&
            owner.kind === 'LabeledStatement' &&
            ctx.node(owner.children[0] ?? -1).text === ctx.node(label).text
        )
            return;
        const grandparent = ctx.parent(parent);
        if(grandparent >= 0 && ctx.node(grandparent).kind === 'TryStatement' && owner.kind === 'Block') {
            const children = ctx.node(grandparent).children;
            if(children.length > 1 && children[children.length - 1] === parent && children[0] !== parent) {
                const id =
                    node.kind === 'ReturnStatement'
                        ? 'unsafeReturn'
                        : node.kind === 'ThrowStatement'
                          ? 'unsafeThrow'
                          : node.kind === 'BreakStatement'
                            ? 'unsafeBreak'
                            : 'unsafeContinue';
                const message =
                    node.kind === 'ReturnStatement'
                        ? messageUnsafeFinallyReturn
                        : node.kind === 'ThrowStatement'
                          ? messageUnsafeFinallyThrow
                          : node.kind === 'BreakStatement'
                            ? messageUnsafeFinallyBreak
                            : messageUnsafeFinallyContinue;
                ctx.report(index, 'no-unsafe-finally', id, message);
                return;
            }
        }
    }
}
