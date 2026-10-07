import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';

export class Comment {
    readonly start: number;
    readonly end: number;
    readonly text: string;
    constructor(start: number, end: number, text: string) { this.start = start; this.end = end; this.text = text; }
}
export function comments(context: RuleContext): Comment[] {
    const literalEnds = new Map<number, number>();
    for(let index = 0; index < context.parser.nodes.length; index++) {
        const node = context.node(index);
        if(['StringLiteral', 'RegularExpressionLiteral', 'NoSubstitutionTemplateLiteral', 'TemplateHead', 'TemplateMiddle', 'TemplateTail'].includes(node.kind)) { literalEnds.set(context.start(index), node.end); }
    }
    const source = context.source;
    const result: Comment[] = [];
    for(let index = 0; index < source.length;) {
        const literalEnd = literalEnds.get(index);
        if(literalEnd !== undefined) { index = literalEnd; continue; }
        const pair = source.slice(index, index + 2);
        if(index === 0 && pair === '#!') { const end = source.indexOf('\n', index); index = end < 0 ? source.length : end; continue; }
        if(pair === '//') {
            let end = index + 2;
            while(end < source.length && !['\n', '\r', '\u2028', '\u2029'].includes(source[end] ?? '')) { end++; }
            result.push(new Comment(index, end, source.slice(index, end))); index = end;
        }
        else if(pair === '/*') {
            const close = source.indexOf('*/', index + 2);
            const end = close < 0 ? source.length : close + 2;
            result.push(new Comment(index, end, source.slice(index, end))); index = end;
        }
        else { index++; }
    }
    return result;
}
