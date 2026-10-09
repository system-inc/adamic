import type { RuleContext } from '../../context.ts';
import { messageJsxNoCommentTextnodes } from './messages.ts';
// Visit real JsxText nodes produced by the stage1 parser.
export function jsxNoCommentTextnodes(context: RuleContext, index: number): void {
    if(context.kind(index) !== 'JsxText') {
        return;
    }
    const node = context.node(index);
    const raw = context.source.slice(node.pos, node.end);
    let atLineStart = true;
    for(let position = 0; position < raw.length; position++) {
        const character = raw.slice(position, position + 1);
        if(character === '\n') {
            atLineStart = true;
            continue;
        }
        if([' ', '\t', '\r', '\v', '\f'].includes(character)) {
            continue;
        }
        if(atLineStart && character === '/' && ['/', '*'].includes(raw.slice(position + 1, position + 2))) {
            context.reportRange(
                node.pos,
                node.end,
                'react/jsx-no-comment-textnodes',
                'putCommentInBraces',
                messageJsxNoCommentTextnodes,
            );
            return;
        }
        atLineStart = false;
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        jsxNoCommentTextnodes(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
