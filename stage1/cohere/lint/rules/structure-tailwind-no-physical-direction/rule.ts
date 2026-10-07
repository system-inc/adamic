import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';

function fields(text: string): string[] {
    return text.split(/[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+/).map(value => value ?? '').filter(value => value !== '');
}
const shape = /^-?(?:[a-z]+:)*(?:[a-z]+-)?[a-z]+(?:-[a-z0-9.\[\]/]+)?$/;
const prefixes = ['ml-', 'mr-', 'pl-', 'pr-', 'text-left', 'text-right', 'left-', 'right-', 'border-l', 'border-l-', 'border-r', 'border-r-', 'rounded-l', 'rounded-l-', 'rounded-r', 'rounded-r-', 'scroll-ml-', 'scroll-mr-', 'scroll-pl-', 'scroll-pr-'];
const replacements = ['ms-', 'me-', 'ps-', 'pe-', 'text-start', 'text-end', 'start-', 'end-', 'border-s', 'border-s-', 'border-e', 'border-e-', 'rounded-s', 'rounded-s-', 'rounded-e', 'rounded-e-', 'scroll-ms-', 'scroll-me-', 'scroll-ps-', 'scroll-pe-'];
function logical(base: string): string {
    const negative = base.startsWith('-');
    const body = negative ? base.slice(1) : base;
    for(let index = 0; index < prefixes.length; index++) {
        const prefix = prefixes[index] ?? '';
        const replacement = replacements[index] ?? '';
        if(!prefix.endsWith('-')) {
            if(!negative && body === prefix) { return replacement; }
        }
        else if(body.startsWith(prefix) && body.length > prefix.length && (!negative || index < 4 || index === 6 || index === 7)) {
            return (negative ? '-' : '') + replacement + body.slice(prefix.length);
        }
    }
    return '';
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    report(index: number): void {
        const tokens = fields(this.context.node(index).text);
        if(!tokens.some(token => shape.test(token))) { return; }
        for(const token of tokens) {
            const parts = token.split(':');
            const base = parts.pop() ?? '';
            if(parts.includes('rtl') || parts.includes('ltr')) { continue; }
            const replacement = logical(base);
            if(replacement === '') { continue; }
            const complete = (parts.length === 0 ? '' : parts.join(':') + ':') + replacement;
            this.context.report(index, 'structure/tailwind-no-physical-direction', 'useLogicalClass', message(token, complete), '', '', '');
        }
    }
    visit(index: number): void {
        const path = this.context.parser.path;
        if(!path.endsWith('.ts') && !path.endsWith('.tsx')) { return; }
        const node = this.context.node(index);
        if(node.kind === 'StringLiteral' || node.kind === 'NoSubstitutionTemplateLiteral') { this.report(index); }
        else if(node.kind === 'TemplateExpression') {
            for(const child of node.children) {
                if(this.context.node(child).kind === 'TemplateHead') { this.report(child); }
                else if(this.context.node(child).kind === 'TemplateSpan') {
                    const children = this.context.node(child).children;
                    const literal = children[children.length - 1] ?? -1;
                    if(literal >= 0) { this.report(literal); }
                }
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
