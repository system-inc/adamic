// Port of pinned Go cohere's no-empty-pattern.
import type { RuleContext } from './rule_context.ts';

const messageEmptyObjectPattern =
    'This destructuring pattern is empty, so it binds nothing. It reads like an object literal but is a pattern, which is why the mistake survives review: `const {} = foo` declares no variables at all and only asserts that `foo` is not null. Name the properties being pulled out, or drop the destructuring.';
const messageEmptyArrayPattern =
    'This destructuring pattern is empty, so it binds nothing. `const [] = foo` declares no variables and only forces `foo` to be iterated. Name the elements being pulled out, or drop the destructuring.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-empty-pattern')) return;
    const node = ctx.node(index);
    if((node.kind !== 'ObjectBindingPattern' && node.kind !== 'ArrayBindingPattern') || node.children.length !== 0)
        return;
    if(
        node.kind === 'ObjectBindingPattern' &&
        ctx.settings.read('allowobjectpatternsasparameters', 'false') === 'true'
    ) {
        const parent = ctx.parent(index);
        if(parent >= 0 && ctx.node(parent).kind === 'Parameter') {
            const value = ctx.initializer(parent);
            if(
                value < 0 ||
                (ctx.node(value).kind === 'ObjectLiteralExpression' && ctx.node(value).children.length === 0)
            )
                return;
        }
    }
    ctx.report(
        index,
        'no-empty-pattern',
        node.kind === 'ObjectBindingPattern' ? 'unexpectedObject' : 'unexpectedArray',
        node.kind === 'ObjectBindingPattern' ? messageEmptyObjectPattern : messageEmptyArrayPattern,
    );
}
