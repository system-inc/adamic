// Port of pinned Go cohere's @typescript-eslint/ban-tslint-comment.
import type { RuleContext } from './rule_context.ts';
import { text, kind, range, trim, comments } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/ban-tslint-comment')) return;
    const rule = '@typescript-eslint/ban-tslint-comment';
    if(kind(ctx, index) !== 'SourceFile') return;
    for(const comment of comments(ctx, index)) {
        if(comment.block && comment.end - comment.start < 4) continue;
        const interior = ctx.source.slice(comment.start + 2, comment.end - (comment.block ? 2 : 0));
        let cursor = 0;
        while([' ', '\t', '\r', '\n', '\f'].includes(interior[cursor] ?? '')) cursor++;
        const directive = interior.slice(cursor);
        let length = -1;
        for(const flag of ['enable', 'disable'])
            for(const suffix of ['-next-line', '-line', '']) {
                const prefix = `tslint:${flag}${suffix}`;
                if(
                    directive.startsWith(prefix) &&
                    ['', ':', ' ', '\t', '\r', '\n', '\f'].includes(directive[prefix.length] ?? '')
                )
                    length = prefix.length;
            }
        if(length < 0) continue;
        const rendered = comment.block ? `/* ${trim(interior)} */` : `// ${trim(interior)}`;
        let start = comment.start;
        let end = comment.end;
        const blankCharacters = [' ', '\t', '\r', '\n', '\f', '\v'];
        if(
            start > 0 &&
            !['\n', '\r', '\u2028', '\u2029'].includes(ctx.source[start - 1] ?? '') &&
            blankCharacters.includes(ctx.source[start - 1] ?? '')
        )
            start--;
        if(end < ctx.source.length && blankCharacters.includes(ctx.source[end] ?? '')) end++;
        range(
            ctx,
            rule,
            'commentDetected',
            text('ts_ban_tslint_comment').replace('%s', rendered),
            comment.start,
            comment.end,
            'fix',
            '',
            start,
            end,
        );
    }
}
