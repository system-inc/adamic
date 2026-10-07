import { panic, utf8Length } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { forFile } from './comments/for_file.ts';
import { message } from './messages.ts';
import type { Suggestion } from './repairs.ts';

const directive = /^(eslint(?:-env|-enable|-disable(?:(?:-next)?-line)?)?|exported|globals?)(?:\s|$)/u;
const separator = /\s-{2,}\s/u;
const goWhitespace = /^[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]$/;

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    suggestions(index: number): Suggestion[] { return []; }
    goTrim(text: string): string {
        let start = 0;
        let end = text.length;
        while(start < end && goWhitespace.test(text.slice(start, start + 1))) { start++; }
        while(end > start && goWhitespace.test(text.slice(end - 1, end))) { end--; }
        return text.slice(start, end);
    }
    recognize(text: string, block: boolean): string {
        let body = block ? text.slice(2, -2) : text.slice(2);
        if(block && body.includes('\n')) {
            const lines: string[] = [];
            for(const line of body.split('\n')) {
                const trimmed = this.goTrim(line);
                lines.push(trimmed.startsWith('*') ? trimmed.slice(1) : trimmed);
            }
            body = lines.join(' ');
        }
        body = this.goTrim(body);
        for(const tool of ['cohere', 'verify', 'eslint', 'oxlint']) {
            const enable = `${tool}-enable`;
            if(block && body.startsWith(enable)) {
                const rest = body.slice(enable.length);
                if(rest === '' || rest.startsWith(' ') || rest.startsWith('\t')) { return 'eslint-enable'; }
            }
            const disable = `${tool}-disable`;
            if(!body.startsWith(disable)) { continue; }
            let rest = body.slice(disable.length);
            let kind = 'eslint-disable';
            if(rest.startsWith('-next-line')) { rest = rest.slice(10); kind = 'eslint-disable-next-line'; }
            else if(rest.startsWith('-line')) { rest = rest.slice(5); kind = 'eslint-disable-line'; }
            if(rest !== '' && !rest.startsWith(' ') && !rest.startsWith('\t')) { continue; }
            if(kind !== 'eslint-disable' || block) { return kind; }
        }
        return '';
    }
    visit(index: number, parent: number): void {
        if(this.context.node(index).kind !== 'SourceFile') { return; }
        const ignored = this.context.settings.list('ignore', []);
        const additional = this.context.settings.list('additionaldirectives', []);
        const units = new Map<number, number>();
        let byte = 0;
        for(let cursor = 0; cursor < this.context.source.length;) {
            units.set(byte, cursor);
            const width = (this.context.source.codePointAt(cursor) ?? 0) > 65535 ? 2 : 1;
            byte += utf8Length(this.context.source.slice(cursor, cursor + width));
            cursor += width;
        }
        units.set(byte, this.context.source.length);
        for(const comment of forFile(this.context.parser, index, undefined)) {
            if(comment.isBlock && (comment.text.length < 4 || !comment.text.endsWith('*/'))) { continue; }
            const interior = comment.isBlock ? comment.text.slice(2, -2) : comment.text.slice(2);
            const pieces = interior.split(separator);
            const text = (pieces[0] ?? '').trim();
            const match = directive.exec(text);
            let kind = match === null ? '' : match[1] ?? '';
            if(kind === '') {
                for(const candidate of additional) {
                    if(text.startsWith(candidate) && (text.length === candidate.length || /\s/u.test(text.slice(candidate.length, candidate.length + 1)))) { kind = candidate; break; }
                }
            }
            const supportedLine = ['eslint-disable-line', 'eslint-disable-next-line'].includes(kind) || additional.includes(kind);
            if((!comment.isBlock && !supportedLine) || (kind === 'eslint-disable-line' && comment.startLine !== comment.endLine)) { kind = ''; }
            if(kind === '') { kind = this.recognize(comment.text, comment.isBlock); }
            if(kind === '' || ignored.includes(kind) || (pieces.length > 1 && (pieces[1] ?? '').trim() !== '')) { continue; }
            const start = units.get(comment.start) ?? panic('comment start outside source');
            const end = units.get(comment.end) ?? panic('comment end outside source');
            this.context.findings.push(new Finding('@eslint-community/eslint-comments/require-description', 'missingDescription', message, start, end, '', '', ''));
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
