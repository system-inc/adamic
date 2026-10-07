import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { Diagnostic, diagnostics } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { all } from '../../helpers/comments/all.ts';
import { trimGo } from '../tailwind-no-physical-direction/rule.ts';
const message = 'Unexpected undescribed directive comment. Include descriptions to explain why the comment is necessary: write the reason after ` -- ` at the end of the directive, as in `// eslint-disable-next-line some-rule -- why this line is the exception`. A suppression that does not say why cannot be told apart from one nobody remembers adding.';
function boundary(text: string): boolean { return text === '' || text.startsWith(' ') || text.startsWith('\t'); }
function recognize(text: string, block: boolean): string {
    let body = text;
    if(block && body.includes('\n')) {
        const lines = body.split('\n');
        for(let i = 0; i < lines.length; i++) {
            let line = trimGo(lines[i] ?? '');
            if(line.startsWith('*')) { line = line.slice(1); } lines[i] = line;
        }
        body = lines.join(' ');
    }
    // Go TrimSpace differs from JS trim at U+0085 and U+FEFF.
    body = trimGo(body);
    for(const tool of ['cohere','verify','eslint','oxlint']) {
        const disable = `${tool}-disable`;
        if(body.startsWith(disable)) {
            let rest = body.slice(disable.length); let kind = 'eslint-disable';
            if(rest.startsWith('-next-line')) { rest = rest.slice(10); kind = 'eslint-disable-next-line'; }
            else if(rest.startsWith('-line')) { rest = rest.slice(5); kind = 'eslint-disable-line'; }
            if(boundary(rest) && (block || kind !== 'eslint-disable')) { return kind; }
        }
        const enable = `${tool}-enable`;
        if(block && body.startsWith(enable) && boundary(body.slice(enable.length))) { return 'eslint-enable'; }
    }
    return '';
}
export class Rule {
    readonly context: RuleContext;
    readonly ignored: string[];
    readonly additional: string[];
    constructor(context: RuleContext) {
        this.context = context;
        this.ignored = context.settings.list('ignore',[]);
        this.additional = context.settings.list('additionaldirectives',[]);
    }
    visit(index: number, parent: number): void {
        const offsets = new Map<number,number>();
        let bytes = 0;
        for(let i = 0; i < this.context.source.length; i++) {
            offsets.set(bytes,i);
            const code = this.context.source.codePointAt(i) ?? 0;
            bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
            if(code > 65535) { i++; }
        }
        offsets.set(bytes,this.context.source.length);
        for(const comment of all(this.context.parser,index)) {
            if(comment.isBlock && !comment.text.endsWith('*/')) { continue; }
            const interior = comment.isBlock ? comment.text.slice(2,-2) : comment.text.slice(2);
            const pieces = interior.split(/\s-{2,}\s/u);
            const body = (pieces[0] ?? '').trim();
            const match = /^(eslint(?:-env|-enable|-disable(?:(?:-next)?-line)?)?|exported|globals?)(?:\s|$)/u.exec(body);
            let kind = match === null ? '' : match[1] ?? '';
            let found = match !== null;
            if(!found) {
                for(const candidate of this.additional) {
                    if(body.startsWith(candidate) && (body.length === candidate.length || /\s/u.test(body.slice(candidate.length,candidate.length + 1)))) { kind = candidate; found = true; break; }
                }
            }
            if(!comment.isBlock && kind !== 'eslint-disable-line' && kind !== 'eslint-disable-next-line' && !this.additional.includes(kind)) { found = false; }
            if(kind === 'eslint-disable-line' && comment.startLine !== comment.endLine) { found = false; }
            if(!found) { kind = recognize(interior,comment.isBlock); found = kind !== ''; }
            if(!found || this.ignored.includes(kind) || (pieces[1] ?? '').trim() !== '') { continue; }
            const start = offsets.get(comment.start) ?? panic('comment start outside source');
            const end = offsets.get(comment.end) ?? panic('comment end outside source');
            const name = '@eslint-community/eslint-comments/require-description';
            this.context.findings.push(new Finding(name,'missingDescription',message,start,end,'','',''));
            diagnostics.push(new Diagnostic(name,'missingDescription',message,start,end,[]));
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
