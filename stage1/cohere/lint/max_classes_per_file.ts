// Port of pinned Go cohere's max-classes-per-file.
import type { RuleContext } from './rule_context.ts';
import { integer, text, kind, range } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('max-classes-per-file')) return;
    const rule = 'max-classes-per-file';
    if(kind(ctx, index) !== 'SourceFile') return;
    const maximum = integer(ctx.settings.read('maximum', '1'));
    const ignore = ctx.settings.read('ignoreexpressions', 'false') === 'true';
    let count = 0;
    for(let i = 0; i < ctx.parser.nodes.length; i++) {
        if(kind(ctx, i) === 'ClassExpression' && !ignore) count++;
        if(kind(ctx, i) === 'ClassDeclaration') {
            let parent = ctx.parent(i);
            let inside = false;
            while(parent >= 0) {
                if(ctx.functionLike(parent)) inside = true;
                parent = ctx.parent(parent);
            }
            if(!inside) count++;
        }
    }
    if(count <= maximum) return;
    const statements = ctx.node(index).children.filter((part) => kind(ctx, part) !== 'EndOfFile');
    const first = statements[0] ?? -1;
    const end = statements[statements.length - 1] ?? -1;
    if(first < 0 || end < 0) return;
    range(
        ctx,
        rule,
        'maximumExceeded',
        text('max_classes_per_file').replace('%d', `${count}`).replace('%d', `${maximum}`),
        ctx.start(first),
        ctx.node(end).end,
    );
}
