// Literal leaves shield slashes in strings, regexp and templates from comment collection.
import { panic } from 'adamic';
import type { Converter } from './convert.ts';
import { stringValue } from './values.ts';
export function collectComments(converter: Converter, root: number): number[] {
    const starts = new Map<number, number>();
    const stack = [root];
    while(stack.length > 0) {
        const id = stack.pop() ?? panic('missing parser node');
        const node = converter.ts(id);
        if(
            node.children.length === 0 &&
            [
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
                'RegularExpressionLiteral',
                'NoSubstitutionTemplateLiteral',
                'TemplateHead',
                'TemplateMiddle',
                'TemplateTail',
                'JsxText',
                'JsxString',
            ].includes(node.kind)
        ) {
            starts.set(node.kind.startsWith('Jsx') ? node.pos : converter.startUnit(id), node.end);
        }
        for(const child of node.children) {
            stack.push(child);
        }
    }
    const comments: number[] = [];
    for(let index = 0; index < converter.text.length;) {
        const end = starts.get(index);
        if(end !== undefined) {
            index = end;
            continue;
        }
        const prefix = converter.text.slice(index, index + 2);
        if(prefix !== '//' && prefix !== '/*') {
            index++;
            continue;
        }
        const start = index;
        const block = prefix === '/*';
        index += 2;
        const valueStart = index;
        if(block) {
            const close = converter.text.indexOf('*/', index);
            if(close < 0) {
                panic('unterminated block comment');
            }
            index = close + 2;
        }
        else {
            while(
                index < converter.text.length &&
                !['\n', '\r', '\u2028', '\u2029'].includes(converter.text.slice(index, index + 1))
            ) {
                index++;
            }
        }
        const id = converter.arena.newNode(block ? 'Block' : 'Line', converter.byte(start), converter.byte(index));
        converter.arena.node(id).set('value', stringValue(converter.text.slice(valueStart, block ? index - 2 : index)));
        comments.push(id);
    }
    return comments;
}
