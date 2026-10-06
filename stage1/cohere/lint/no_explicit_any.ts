import type { Batch4Context } from './batch4_context.ts';
const messageNoExplicitAny =
    'This annotation is `any`, which switches the type system off for every value that flows through it: no property is checked, no argument is checked, and no assignment is refused. The compiler stops answering questions about this value and reports nothing when it is used wrongly. Write the type it actually holds, or `unknown` when the shape is genuinely not known yet, which keeps the value opaque until it is narrowed rather than treating it as anything at all.';

export function visit(ctx: Batch4Context, index: number): void {
    if(!['.ts', '.tsx', '.mts', '.cts'].some((suffix) => ctx.parser.path.endsWith(suffix))) return;
    if(!ctx.enabled('@typescript-eslint/no-explicit-any') || ctx.node(index).kind !== 'AnyKeyword') return;
    if(ctx.settings.read('ignorerestargs', 'false') === 'true') {
        for(let parent = ctx.parent(index); parent >= 0; parent = ctx.parent(parent)) {
            if(ctx.node(parent).kind === 'Parameter' && ctx.hasKind(parent, 'DotDotDotToken')) return;
        }
    }
    const fix = ctx.settings.read('fixtounknown', 'false') === 'true';
    ctx.report(
        index,
        '@typescript-eslint/no-explicit-any',
        'unexpectedAny',
        messageNoExplicitAny,
        fix ? 'fix' : '',
        fix ? 'unknown' : '',
    );
}
