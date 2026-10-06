import { panic } from 'adamic';
import { space } from './comments.ts';
import type { Batch4Context } from './batch4_context.ts';

function blank(value: string): boolean {
    return value !== '' && ' \t\n\r\f\v'.includes(value);
}
function trimmed(text: string): string {
    let start = 0;
    let end = text.length;
    while(start < end && space(text[start] ?? '')) start++;
    while(end > start && space(text[end - 1] ?? '')) end--;
    return text.slice(start, end);
}
export function visit(ctx: Batch4Context, index: number): void {
    if(!ctx.enabled('@typescript-eslint/ban-tslint-comment') || ctx.node(index).kind !== 'SourceFile') return;
    for(let cursor = 0; cursor < ctx.commentStarts.length; cursor++) {
        const start = ctx.commentStarts[cursor] ?? panic('comment start');
        const end = ctx.commentEnds[cursor] ?? panic('comment end');
        const block = ctx.source.slice(start, start + 2) === '/*';
        if(block && end - start < 4) continue;
        const interior = ctx.source.slice(start + 2, block ? end - 2 : end);
        // Go regexp's \s is the ASCII set, distinct from String.trim's set.
        let at = 0;
        while(at < interior.length && ' \t\n\r\f'.includes(interior[at] ?? '')) at++;
        const directive = interior.slice(at);
        let size = directive.startsWith('tslint:enable') ? 13 : directive.startsWith('tslint:disable') ? 14 : -1;
        if(size < 0) continue;
        if(directive.slice(size).startsWith('-next-line')) size += 10;
        else if(directive.slice(size).startsWith('-line')) size += 5;
        const next = directive[size] ?? '';
        if(next !== '' && next !== ':' && !' \t\n\r\f'.includes(next)) continue;
        const rendered = block ? `/* ${trimmed(interior)} */` : `// ${trimmed(interior)}`;
        const finding = ctx.report(
            index,
            '@typescript-eslint/ban-tslint-comment',
            'commentDetected',
            `This file carries the tslint directive ${rendered}. tslint has been deprecated since 2019 and nothing in this project reads its directives, so the comment suppresses nothing and instead reads as a live suppression to anyone who finds it. Delete it, and if the code it guarded still warrants a suppression, write the equivalent eslint-disable comment naming the rule.`,
            'fix',
        );
        finding.editStart =
            start > 0 &&
            ctx.source[start - 1] !== '\n' &&
            ctx.source[start - 1] !== '\r' &&
            blank(ctx.source[start - 1] ?? '')
                ? start - 1
                : start;
        finding.editEnd = end < ctx.source.length && blank(ctx.source[end] ?? '') ? end + 1 : end;
        ctx.replaceRange(finding, start, end);
    }
}
