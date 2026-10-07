import { panic, utf8Length } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { all } from '../../helpers/comments/all.ts';
import { missingDescription } from './messages.ts';
import { checked } from './options.ts';
function space(code: number): boolean {
    return code === 9 || code === 10 || code === 11 || code === 12 || code === 13 || code === 32 || code === 160 || code === 5760 || (code >= 8192 && code <= 8202) || code === 8232 || code === 8233 || code === 8239 || code === 8287 || code === 12288 || code === 65279;
}
function goTrim(text: string): string {
    let start = 0;
    let end = text.length;
    while(start < end && (space(text.charCodeAt(start)) && text.charCodeAt(start) !== 65279 || text.charCodeAt(start) === 133)) { start++; }
    while(end > start && (space(text.charCodeAt(end - 1)) && text.charCodeAt(end - 1) !== 65279 || text.charCodeAt(end - 1) === 133)) { end--; }
    return text.slice(start, end);
}
function separator(text: string, from: number): number {
    for(let index = from; index < text.length; index++) {
        if(!space(text.charCodeAt(index))) { continue; }
        let end = index + 1;
        while(text.charCodeAt(end) === 45) { end++; }
        if(end >= index + 3 && space(text.charCodeAt(end))) { return index; }
    }
    return -1;
}
function afterSeparator(text: string, start: number): number {
    let cursor = start + 1;
    while(text.charCodeAt(cursor) === 45) { cursor++; }
    return cursor + 1;
}
function hasDescription(text: string): boolean {
    const first = separator(text, 0);
    if(first < 0) { return false; }
    const start = afterSeparator(text, first);
    const next = separator(text, start);
    return text.slice(start, next < 0 ? text.length : next).trim() !== '';
}
function whole(text: string, word: string): boolean {
    return text === word || text.startsWith(`${word} `) || text.startsWith(`${word}\t`);
}
function recognized(raw: string, block: boolean): string {
    let body = raw.slice(2, block ? -2 : raw.length);
    if(block && body.includes('\n')) {
        body = body.split('\n').map((line) => { const trimmed = goTrim(line); return trimmed.startsWith('*') ? trimmed.slice(1) : trimmed; }).join(' ');
    }
    body = goTrim(body);
    for(const tool of ['cohere', 'verify', 'eslint', 'oxlint']) {
        for(const suffix of ['-disable-next-line', '-disable-line', '-disable', '-enable']) {
            if(!block && (suffix === '-disable' || suffix === '-enable')) { continue; }
            if(whole(body, tool + suffix)) { return 'eslint' + suffix; }
        }
    }
    return '';
}
export class Rule {
    readonly context: RuleContext;
    readonly rawOptions: string;
    constructor(context: RuleContext, rawOptions: string = '') { this.context = context; this.rawOptions = rawOptions; }
    visit(root: number): void {
        const context = this.context;
        if(this.rawOptions === '' && context.settings.values.size > 0) { panic('NotYet: shared factory must supply raw description options'); }
        const options = checked(this.rawOptions);
        const ignored = options.ignored;
        const additional = options.additional;
        // Helpers expose UTF-8 bytes; findings expose UTF-16 units.
        const units = new Map<number, number>();
        let byte = 0;
        for(let unit = 0; unit < context.source.length;) {
            units.set(byte, unit);
            const width = (context.source.codePointAt(unit) ?? 0) > 65535 ? 2 : 1;
            byte += utf8Length(context.source.slice(unit, unit + width));
            unit += width;
        }
        units.set(byte, context.source.length);
        for(const comment of all(context.parser, root)) {
            if(comment.isBlock && (comment.text.length < 4 || !comment.text.endsWith('*/'))) { continue; }
            const interior = comment.text.slice(2, comment.isBlock ? -2 : comment.text.length);
            const split = separator(interior, 0);
            const text = (split < 0 ? interior : interior.slice(0, split)).trim();
            let kind = '';
            for(const candidate of ['eslint-env', 'eslint-enable', 'eslint-disable-next-line', 'eslint-disable-line', 'eslint-disable', 'eslint', 'exported', 'globals', 'global']) {
                if(text === candidate || text.startsWith(candidate) && space(text.charCodeAt(candidate.length))) { kind = candidate; break; }
            }
            if(kind === '') {
                for(const candidate of additional) {
                    if(text === candidate || text.startsWith(candidate) && space(text.charCodeAt(candidate.length))) { kind = candidate; break; }
                }
            }
            if(!comment.isBlock && kind !== 'eslint-disable-line' && kind !== 'eslint-disable-next-line' && !additional.includes(kind)) { kind = ''; }
            if(kind === 'eslint-disable-line' && comment.startLine !== comment.endLine) { kind = ''; }
            if(kind === '') { kind = recognized(comment.text, comment.isBlock); }
            if(kind === '' || ignored.includes(kind) || hasDescription(interior)) { continue; }
            context.findings.push(new Finding('@eslint-community/eslint-comments/require-description', 'missingDescription', missingDescription, units.get(comment.start) ?? 0, units.get(comment.end) ?? 0, '', '', ''));
        }
    }
}
export function create(context: RuleContext, rawOptions: string = ''): Rule { return new Rule(context, rawOptions); }
