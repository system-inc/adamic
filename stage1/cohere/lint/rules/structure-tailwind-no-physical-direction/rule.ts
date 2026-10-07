export { syntaxKinds } from './listeners.a';
import type { RuleContext } from '../../context.ts';
import { message } from './messages.a';

interface Mapping {
    readonly prefix: string;
    readonly replacement: string;
    readonly exact: boolean;
    readonly negatable: boolean;
}
const mappings: readonly Mapping[] = [
    { prefix: 'ml-', replacement: 'ms-', exact: false, negatable: true },
    { prefix: 'mr-', replacement: 'me-', exact: false, negatable: true },
    { prefix: 'pl-', replacement: 'ps-', exact: false, negatable: true },
    { prefix: 'pr-', replacement: 'pe-', exact: false, negatable: true },
    { prefix: 'text-left', replacement: 'text-start', exact: true, negatable: false },
    { prefix: 'text-right', replacement: 'text-end', exact: true, negatable: false },
    { prefix: 'left-', replacement: 'start-', exact: false, negatable: true },
    { prefix: 'right-', replacement: 'end-', exact: false, negatable: true },
    { prefix: 'border-l', replacement: 'border-s', exact: true, negatable: false },
    { prefix: 'border-l-', replacement: 'border-s-', exact: false, negatable: false },
    { prefix: 'border-r', replacement: 'border-e', exact: true, negatable: false },
    { prefix: 'border-r-', replacement: 'border-e-', exact: false, negatable: false },
    { prefix: 'rounded-l', replacement: 'rounded-s', exact: true, negatable: false },
    { prefix: 'rounded-l-', replacement: 'rounded-s-', exact: false, negatable: false },
    { prefix: 'rounded-r', replacement: 'rounded-e', exact: true, negatable: false },
    { prefix: 'rounded-r-', replacement: 'rounded-e-', exact: false, negatable: false },
    { prefix: 'scroll-ml-', replacement: 'scroll-ms-', exact: false, negatable: false },
    { prefix: 'scroll-mr-', replacement: 'scroll-me-', exact: false, negatable: false },
    { prefix: 'scroll-pl-', replacement: 'scroll-ps-', exact: false, negatable: false },
    { prefix: 'scroll-pr-', replacement: 'scroll-pe-', exact: false, negatable: false },
];

// strings.Fields uses Go Unicode whitespace, which differs from JavaScript \s.
function space(code: number): boolean {
    return (code >= 9 && code <= 13) || code === 32 || code === 133 || code === 160 ||
        code === 5760 || (code >= 8192 && code <= 8202) || code === 8232 || code === 8233 ||
        code === 8239 || code === 8287 || code === 12288;
}
function fields(text: string): string[] {
    const result: string[] = [];
    let start = 0;
    for(let index = 0; index <= text.length; index++) {
        if(index === text.length || space(text.charCodeAt(index))) {
            if(start < index) result.push(text.slice(start, index));
            start = index + 1;
        }
    }
    return result;
}
const pattern = /^-?(?:[a-z]+:)*(?:[a-z]+-)?[a-z]+(?:-[a-z0-9.\[\]/]+)?$/;

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    report(index: number): void {
        const tokens = fields(this.context.node(index).text);
        if(!tokens.some((token) => pattern.test(token))) return;
        for(const token of tokens) {
            const parts = token.split(':');
            const base = parts.pop() ?? '';
            if(parts.includes('rtl') || parts.includes('ltr')) continue;
            const negative = base.startsWith('-') ? '-' : '';
            const body = negative === '' ? base : base.slice(1);
            for(const mapping of mappings) {
                if(negative !== '' && !mapping.negatable) continue;
                if(mapping.exact ? body !== mapping.prefix : !body.startsWith(mapping.prefix) || body.length === mapping.prefix.length) continue;
                const replacement = (parts.length === 0 ? '' : `${parts.join(':')}:`) + negative +
                    mapping.replacement + (mapping.exact ? '' : body.slice(mapping.prefix.length));
                this.context.report(index, 'structure/tailwind-no-physical-direction', 'useLogicalClass',
                    message(token, replacement), '', '', '');
                break;
            }
        }
    }
    visit(index: number): void {
        const path = this.context.parser.path;
        if(!path.endsWith('.ts') && !path.endsWith('.tsx')) return;
        const node = this.context.node(index);
        if(node.kind === 'TemplateExpression') {
            for(const child of node.children) {
                if(this.context.node(child).kind === 'TemplateHead') this.report(child);
                if(this.context.node(child).kind === 'TemplateSpan') {
                    for(const literal of this.context.node(child).children) {
                        if(['TemplateMiddle', 'TemplateTail'].includes(this.context.node(literal).kind)) this.report(literal);
                    }
                }
            }
        }
        else this.report(index);
    }
}
export function create(context: RuleContext): Rule {
    return new Rule(context);
}
