import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { prefixes, replacements, exact, negatable, message } from './table.ts';
function space(code: number): boolean {
    return (code >= 9 && code <= 13) || [32,133,160,5760,8232,8233,8239,8287,12288].includes(code) || (code >= 8192 && code <= 8202);
}
export function trimGo(text: string): string {
    let start = 0; let end = text.length;
    while(start < end && space(text.charCodeAt(start))) { start++; }
    while(end > start && space(text.charCodeAt(end - 1))) { end--; }
    return text.slice(start,end);
}
export function fields(text: string): string[] {
    const result: string[] = []; let start = 0;
    for(let index = 0; index <= text.length; index++) {
        if(index === text.length || space(text.charCodeAt(index))) {
            if(index > start) { result.push(text.slice(start,index)); } start = index + 1;
        }
    }
    return result;
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    literal(index: number): void {
        const tokens = fields(this.context.node(index).text);
        const pattern = /^-?(?:[a-z]+:)*(?:[a-z]+-)?[a-z]+(?:-[a-z0-9.\[\]/]+)?$/;
        let qualifies = false;
        for(const token of tokens) { if(pattern.test(token)) { qualifies = true; } }
        if(!qualifies) { return; }
        for(const token of tokens) {
            const parts = token.split(':');
            const last = parts.length - 1;
            let directional = false;
            for(let i = 0; i < last; i++) { if(parts[i] === 'rtl' || parts[i] === 'ltr') { directional = true; } }
            if(directional) { continue; }
            let body = parts[last] ?? panic('missing class');
            const negative = body.startsWith('-') ? '-' : '';
            if(negative !== '') { body = body.slice(1); }
            for(let i = 0; i < prefixes.length; i++) {
                const prefix = prefixes[i] ?? panic('missing prefix');
                const isExact = exact[i] ?? false;
                if(isExact ? body !== prefix || negative !== '' : !body.startsWith(prefix) || body.length === prefix.length || negative !== '' && !(negatable[i] ?? false)) { continue; }
                const replacement = negative + (replacements[i] ?? panic('missing replacement')) + (isExact ? '' : body.slice(prefix.length));
                parts[last] = replacement;
                record(this.context,index,'structure/tailwind-no-physical-direction','useLogicalClass',message.replace('{{original}}',token).replace('{{replacement}}',parts.join(':')),[]);
                break;
            }
        }
    }
    visit(index: number, parent: number): void {
        const path = this.context.parser.path;
        if(!path.endsWith('.ts') && !path.endsWith('.tsx')) { return; }
        const node = this.context.node(index);
        if(node.kind !== 'TemplateExpression') { this.literal(index); return; }
        for(const child of node.children) {
            const part = this.context.node(child);
            if(part.kind === 'TemplateHead') { this.literal(child); }
            else if(part.kind === 'TemplateSpan') {
                const literal = part.children[part.children.length - 1] ?? panic('missing template literal');
                this.literal(literal);
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
