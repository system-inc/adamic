import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { entities } from './entities.ts';
const missing = 'This Google Fonts <link> has no `display` parameter, so the browser uses the default and hides the text while the font downloads. On a slow connection that is a blank block where the words should be. Add `&display=optional` or `&display=swap` to the href so the text is painted in a fallback face first.';
const discouraged = 'This Google Fonts <link> sets `display` to auto, block, or fallback, each of which blocks or flashes the text while the font loads. Use `optional`, which paints a fallback immediately and only swaps if the font arrives in time, or `swap` if the exact face matters more than the reflow.';
export function decode(text: string): string {
    let result = '';
    for(let index = 0; index < text.length; index++) {
        if(text[index] !== '&') { result += text[index] ?? ''; continue; }
        const close = text.indexOf(';',index);
        if(close < index + 2) { result += '&'; continue; }
        const item = text.slice(index + 1,close);
        let replacement = '';
        let recognized = false;
        if(item.startsWith('#')) {
            const hex = item.startsWith('#x');
            const digits = item.slice(hex ? 2 : 1);
            if(digits !== '' && (hex ? /^[0-9a-fA-F]+$/.test(digits) : /^[0-9]+$/.test(digits))) {
                let value = 0;
                for(const digit of digits) {
                    const at = '0123456789abcdef'.indexOf(digit.toLowerCase());
                    value = Math.min(1114112,value * (hex ? 16 : 10) + at);
                }
                replacement = value > 1114111 ? text.slice(index,close + 1) : value >= 55296 && value <= 57343 ? '\ufffd' : String.fromCodePoint(value);
                recognized = true;
            }
        }
        else if(/^[0-9a-zA-Z]+$/.test(item)) { replacement = entities.get(item) ?? text.slice(index,close + 1); recognized = true; }
        if(recognized) { result += replacement; index = close; } else { result += '&'; }
    }
    return result;
}
export function displayFinding(reference: string): string {
    if(!reference.startsWith('https://fonts.googleapis.com/css')) { return ''; }
    const question = reference.indexOf('?');
    if(question < 0) { return 'googleFontDisplayMissing'; }
    for(const pair of reference.slice(question + 1).split('&')) {
        const equal = pair.indexOf('=');
        if(equal < 0 || pair.slice(0,equal) !== 'display') { continue; }
        return ['auto','block','fallback'].includes(pair.slice(equal + 1)) ? 'googleFontDisplayNotRecommended' : '';
    }
    return 'googleFontDisplayMissing';
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        const tag = this.context.node(node.children[0] ?? panic('missing JSX tag'));
        if(tag.kind !== 'Identifier' || tag.text !== 'link') { return; }
        let attributes = -1;
        for(const child of node.children) { if(this.context.node(child).kind === 'JsxAttributes') { attributes = child; } }
        if(attributes < 0) { return; }
        for(const child of this.context.node(attributes).children) {
            const attribute = this.context.node(child);
            if(attribute.kind !== 'JsxAttribute') { continue; }
            const name = this.context.node(attribute.children[0] ?? panic('missing attribute name'));
            if(name.kind !== 'Identifier' || name.text.toLowerCase() !== 'href') { continue; }
            const value = attribute.children[1] ?? -1;
            if(value < 0 || this.context.node(value).kind !== 'StringLiteral') { return; }
            const id = displayFinding(decode(this.context.node(value).text));
            if(id !== '') { record(this.context,index,'@next/next/google-font-display',id,id === 'googleFontDisplayMissing' ? missing : discouraged,[]); }
            return;
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
