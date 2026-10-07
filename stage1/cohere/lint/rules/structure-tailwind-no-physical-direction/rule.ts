import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';

const shape = /^-?(?:[a-z]+:)*(?:[a-z]+-)?[a-z]+(?:-[a-z0-9.\[\]/]+)?$/;
// Go strings.Fields uses Unicode White_Space, which excludes BOM and includes NEL.
const fields = /[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+/;
const prefixes = ['ml-', 'mr-', 'pl-', 'pr-', 'text-left', 'text-right', 'left-', 'right-', 'border-l', 'border-l-', 'border-r', 'border-r-', 'rounded-l', 'rounded-l-', 'rounded-r', 'rounded-r-', 'scroll-ml-', 'scroll-mr-', 'scroll-pl-', 'scroll-pr-'];
const replacements = ['ms-', 'me-', 'ps-', 'pe-', 'text-start', 'text-end', 'start-', 'end-', 'border-s', 'border-s-', 'border-e', 'border-e-', 'rounded-s', 'rounded-s-', 'rounded-e', 'rounded-e-', 'scroll-ms-', 'scroll-me-', 'scroll-ps-', 'scroll-pe-'];
const exact = [false, false, false, false, true, true, false, false, true, false, true, false, true, false, true, false, false, false, false, false];
const negatable = [true, true, true, true, false, false, true, true, false, false, false, false, false, false, false, false, false, false, false, false];
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number): void {
        const path = this.context.parser.path;
        if(!path.endsWith('.ts') && !path.endsWith('.tsx')) { return; }
        const tokens = this.context.node(index).text.split(fields).map((token) => token ?? '').filter((token) => token !== '');
        if(!tokens.some((token) => shape.test(token))) { return; }
        for(const token of tokens) {
            const parts = token.split(':');
            const base = parts.pop() ?? '';
            if(parts.includes('rtl') || parts.includes('ltr')) { continue; }
            const negative = base.startsWith('-');
            const body = negative ? base.slice(1) : base;
            for(let entry = 0; entry < prefixes.length; entry++) {
                const prefix = prefixes[entry] ?? '';
                const matches = exact[entry] === true ? body === prefix : body.startsWith(prefix) && body.length > prefix.length;
                if(!matches || (negative && negatable[entry] !== true)) { continue; }
                const suffix = exact[entry] === true ? '' : body.slice(prefix.length);
                const rewritten = (parts.length > 0 ? `${parts.join(':')}:` : '') + (negative ? '-' : '') + (replacements[entry] ?? '') + suffix;
                this.context.report(index, 'structure/tailwind-no-physical-direction', 'useLogicalClass', message(token, rewritten), '', '', '');
                break;
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
