// Port of pinned Go cohere's max-lines.
import type { RuleContext } from './rule_context.ts';
import { integer, text, kind, range, blank, comments } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('max-lines')) return;
    const rule = 'max-lines';
    if(kind(ctx, index) !== 'SourceFile') return;
    const maximum = integer(ctx.settings.read('maximum', '300'));
    const skipBlank = ctx.settings.read('skipblanklines', 'false') === 'true';
    const skipComments = ctx.settings.read('skipcomments', 'false') === 'true';
    const mask: boolean[] = skipComments ? new Array<boolean>(ctx.source.length).fill(false) : [];
    if(skipComments)
        for(const comment of comments(ctx, index)) for(let i = comment.start; i < comment.end; i++) mask[i] = true;
    let counted = 0;
    let excess = 0;
    let start = 0;
    while(start <= ctx.source.length) {
        let end = start;
        while(end < ctx.source.length && !['\r', '\n', '\u2028', '\u2029'].includes(ctx.source[end] ?? '')) end++;
        let outside = '';
        let touched = false;
        if(skipComments)
            for(let i = start; i < end; i++) {
                if(mask[i] === true) touched = true;
                else outside += ctx.source[i] ?? '';
            }
        if(!(skipBlank && blank(ctx.source.slice(start, end))) && !(skipComments && touched && blank(outside))) {
            if(counted === maximum) excess = start;
            counted++;
        }
        if(end === ctx.source.length) break;
        start = end + (ctx.source.slice(end, end + 2) === '\r\n' ? 2 : 1);
        if(start === ctx.source.length) break;
    }
    if(counted > maximum)
        range(
            ctx,
            rule,
            'exceed',
            text('max_lines').replace('%d', `${counted}`).replace('%d', `${maximum}`),
            excess,
            ctx.source.length,
        );
}
