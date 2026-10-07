import type { RuleContext } from '../../context.ts';
import { messageOctalEscape } from './messages.ts';
export function noOctalEscape(context: RuleContext, index: number): void {
    if(context.kind(index) !== 'StringLiteral') {
        return;
    }
    const raw = context.raw(index);
    for(let position = 0; position < raw.length; position++) {
        if(raw.slice(position, position + 1) !== '\\') {
            continue;
        }
        const start = position + 1;
        const first = raw.slice(start, start + 1);
        let width = 0;
        if(first >= '0' && first <= '3') {
            width = 1;
            while(
                width < 3 &&
                raw.slice(start + width, start + width + 1) >= '0' &&
                raw.slice(start + width, start + width + 1) <= '7'
            ) {
                width++;
            }
            if(width < 2) {
                width = 0;
            }
        }
        if(
            width === 0 &&
            first >= '4' &&
            first <= '7' &&
            raw.slice(start + 1, start + 2) >= '0' &&
            raw.slice(start + 1, start + 2) <= '7'
        ) {
            width = 2;
        }
        if(
            width === 0 &&
            ((first >= '1' && first <= '7') || (first === '0' && ['8', '9'].includes(raw.slice(start + 1, start + 2))))
        ) {
            width = 1;
        }
        if(width !== 0) {
            context.reportNode(
                index,
                'no-octal-escape',
                'octalEscapeSequence',
                `Do not use the octal escape \`\\${raw.slice(start, start + width)}\`. ${messageOctalEscape}`,
            );
            return;
        }
        position++;
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        noOctalEscape(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
