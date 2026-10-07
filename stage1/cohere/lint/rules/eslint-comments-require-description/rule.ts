import { utf8Length } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { all } from '../../helpers/comments/all.ts';
import { message } from './messages.ts';

function whitespace(char: string): boolean { return char !== '' && char.trim() === ''; }
function pieces(text: string): string[] {
    const result: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length && result.length < 2; index++) {
        if(!whitespace(text[index] ?? '')) { continue; }
        let end = index + 1;
        while(text[end] === '-') { end++; }
        if(end < index + 3 || !whitespace(text[end] ?? '')) { continue; }
        result.push(text.slice(start, index));
        start = end + 1;
        index = end;
    }
    result.push(text.slice(start));
    return result;
}
function upstream(text: string, block: boolean, additional: readonly string[]): string {
    const body = (pieces(text)[0] ?? '').trim();
    const kinds = ['eslint-env', 'eslint-enable', 'eslint-disable-next-line', 'eslint-disable-line', 'eslint-disable', 'eslint', 'exported', 'globals', 'global'];
    let kind = '';
    for(const candidate of kinds) {
        if(body.startsWith(candidate) && (body.length === candidate.length || whitespace(body[candidate.length] ?? ''))) { kind = candidate; break; }
    }
    if(kind === '') {
        for(const candidate of additional) {
            if(body.startsWith(candidate) && (body.length === candidate.length || whitespace(body[candidate.length] ?? ''))) { kind = candidate; break; }
        }
    }
    if(!block && kind !== 'eslint-disable-line' && kind !== 'eslint-disable-next-line' && !additional.includes(kind)) { return ''; }
    if(kind === 'eslint-disable-line' && /[\n\r\u2028\u2029]/.test(text)) { return ''; }
    return kind;
}
function goTrim(text: string): string { return text.replace(/^[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+|[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+$/g, ''); }
function whole(rest: string): boolean { return rest === '' || rest.startsWith(' ') || rest.startsWith('\t'); }
function recognized(text: string, block: boolean): string {
    let body = text;
    if(block && body.includes('\n')) {
        body = body.split('\n').map(line => { const trimmed = goTrim(line); return trimmed.startsWith('*') ? trimmed.slice(1) : trimmed; }).join(' ');
    }
    body = goTrim(body);
    for(const tool of ['cohere', 'verify', 'eslint', 'oxlint']) {
        const disable = tool + '-disable';
        if(body.startsWith(disable)) {
            const rest = body.slice(disable.length);
            if(rest.startsWith('-next-line') && whole(rest.slice(10))) { return 'eslint-disable-next-line'; }
            if(rest.startsWith('-line') && whole(rest.slice(5))) { return 'eslint-disable-line'; }
            if(block && whole(rest)) { return 'eslint-disable'; }
        }
        const enable = tool + '-enable';
        if(block && body.startsWith(enable) && whole(body.slice(enable.length))) { return 'eslint-enable'; }
    }
    return '';
}
function unitAtByte(text: string, target: number): number {
    let bytes = 0;
    let unit = 0;
    while(unit < text.length && bytes < target) {
        const width = (text.codePointAt(unit) ?? 0) > 65535 ? 2 : 1;
        bytes += utf8Length(text.slice(unit, unit + width));
        unit += width;
    }
    return unit;
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number): void {
        const ignored = this.context.settings.list('ignore', []);
        const additional = this.context.settings.list('additionaldirectives', []);
        for(const comment of all(this.context.parser, index)) {
            if(comment.isBlock && !comment.text.endsWith('*/')) { continue; }
            const interior = comment.text.slice(2, comment.isBlock ? comment.text.length - 2 : comment.text.length);
            let kind = upstream(interior, comment.isBlock, additional);
            if(kind === '') { kind = recognized(interior, comment.isBlock); }
            if(kind === '' || ignored.includes(kind)) { continue; }
            const description = pieces(interior)[1] ?? '';
            if(description.trim() !== '') { continue; }
            this.context.findings.push(new Finding('@eslint-community/eslint-comments/require-description', 'missingDescription', message, unitAtByte(this.context.source, comment.start), unitAtByte(this.context.source, comment.end), '', '', ''));
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
