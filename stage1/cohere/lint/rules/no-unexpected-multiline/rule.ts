import { isIdentifierPart } from '../../../../typescript/scanner/characters.ts';
import type { RuleContext } from '../../context.ts';
import {
    messageUnexpectedMultilineDivision,
    messageUnexpectedMultilineFunction,
    messageUnexpectedMultilineProperty,
    messageUnexpectedMultilineTaggedTemplate,
} from './messages.ts';

function after(context: RuleContext, expression: number, id: string, message: string): void {
    context.scanAt(context.node(expression).end);
    const start = context.scanStart();
    if(start < context.source.length && context.line(start) !== context.line(context.node(expression).end - 1)) {
        context.reportRange(start, start + 1, 'no-unexpected-multiline', id, message);
    }
}
export function noUnexpectedMultiline(context: RuleContext, index: number): void {
    const kind = context.kind(index);
    if(kind === 'CallExpression' && !context.questionDot(index) && context.node(index).list > 0) {
        after(context, context.child(index, 0), 'function', messageUnexpectedMultilineFunction);
    }
    if(kind === 'ElementAccessExpression' && !context.questionDot(index)) {
        after(context, context.child(index, 0), 'property', messageUnexpectedMultilineProperty);
    }
    if(kind === 'TaggedTemplateExpression') {
        const children = context.node(index).children;
        const template = context.property(index);
        let last = context.node(context.child(index, 0)).end;
        if(children.length > 2) {
            context.scanAt(context.node(context.child(index, children.length - 2)).end);
            last = context.scanEnd();
        }
        const start = context.start(template);
        if(context.line(start) !== context.line(last - 1)) {
            context.reportRange(
                start,
                start + 1,
                'no-unexpected-multiline',
                'taggedTemplate',
                messageUnexpectedMultilineTaggedTemplate,
            );
        }
    }
    if(kind !== 'BinaryExpression' || context.node(context.child(index, 1)).kind !== 'SlashToken') {
        return;
    }
    const left = context.child(index, 0);
    if(context.kind(left) !== 'BinaryExpression' || context.kind(context.child(left, 1)) !== 'SlashToken') {
        return;
    }
    const start = context.node(context.child(index, 1)).end;
    let end = start;
    while(end < context.source.length && 'gimsuy'.includes(context.source.slice(end, end + 1))) {
        end++;
    }
    if(end === start) {
        return;
    }
    if(end < context.source.length) {
        const code = context.source.codePointAt(end) ?? 0;
        // Go checks rune(text[end]): its UTF-8 byte, not the decoded identifier or scalar.
        const byte =
            code < 128
                ? code
                : code < 2048
                  ? 192 + Math.floor(code / 64)
                  : code < 65536
                    ? 224 + Math.floor(code / 4096)
                    : 240 + Math.floor(code / 262144);
        if(isIdentifierPart(byte)) {
            return;
        }
    }
    after(context, context.child(left, 0), 'division', messageUnexpectedMultilineDivision);
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        noUnexpectedMultiline(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
