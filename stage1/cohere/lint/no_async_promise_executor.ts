// Port of pinned Go cohere's no-async-promise-executor.
import type { RuleContext } from './rule_context.ts';

const messageNoAsyncPromiseExecutor =
    "This Promise executor is `async`, which silently discards the errors it is there to catch. A rejection thrown inside an async executor becomes a rejection of the executor's own promise, which nothing holds, so the constructed Promise never settles and the error surfaces as an unhandled rejection instead of reaching the `reject` parameter. Needing `await` here usually means the `new Promise` wrapper is unnecessary.";

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-async-promise-executor')) return;
    const node = ctx.node(index);
    if(node.kind !== 'NewExpression' || node.list <= 0) return;
    const callee = ctx.unwrap(node.children[0] ?? -1);
    if(callee < 0 || ctx.node(callee).kind !== 'Identifier' || ctx.node(callee).text !== 'Promise') return;
    const argument = ctx.unwrap(node.children[node.children.length - node.list] ?? -1);
    if(argument < 0 || !['ArrowFunction', 'FunctionExpression'].includes(ctx.node(argument).kind)) return;
    for(const child of ctx.node(argument).children) {
        if(ctx.node(child).kind === 'AsyncKeyword')
            ctx.report(child, 'no-async-promise-executor', 'noAsyncPromiseExecutor', messageNoAsyncPromiseExecutor);
    }
}
