import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { comments } from './comments.a';
import { recognize, registerSubject } from '../../../suppression/directives.ts';
import { message } from './messages.a';

const kinds: readonly string[] = ['eslint', 'eslint-disable', 'eslint-disable-line',
    'eslint-disable-next-line', 'eslint-enable', 'eslint-env', 'exported', 'global', 'globals'];
const separator = /\s-{2,}\s/u;
function directive(text: string, names: readonly string[]): string {
    for(const name of names) {
        if(text.startsWith(name) && (text.length === name.length || /\s/u.test(text.slice(name.length, name.length + 1)))) return name;
    }
    return '';
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        const context = this.context;
        const additional = context.settings.list('additionaldirectives', []);
        const ignored = context.settings.list('ignore', []);
        for(const comment of comments(context)) {
            const text = context.source.slice(comment.pos, comment.end);
            const block = text.startsWith('/*');
            if(block && (text.length < 4 || !text.endsWith('*/'))) continue;
            const interior = block ? text.slice(2, -2) : text.slice(2);
            const pieces = interior.split(separator);
            const head = (pieces[0] ?? '').trim();
            let kind = directive(head, kinds);
            if(kind === '') kind = directive(head, additional);
            const lineSupported = kind === 'eslint-disable-line' || kind === 'eslint-disable-next-line' || additional.includes(kind);
            if(!block && !lineSupported) kind = '';
            if(kind === 'eslint-disable-line' && /[\r\n\u2028\u2029]/u.test(text)) kind = '';
            if(kind === '') kind = recognize(text);
            if(kind === '' || ignored.includes(kind)) continue;
            if((pieces[1] ?? '').trim() !== '') continue;
            context.findings.push(new Finding('@eslint-community/eslint-comments/require-description',
                'missingDescription', message, comment.pos, comment.end, '', '', ''));
        }
    }
}
export function create(context: RuleContext): Rule {
    registerSubject('@eslint-community/eslint-comments/require-description');
    return new Rule(context);
}
