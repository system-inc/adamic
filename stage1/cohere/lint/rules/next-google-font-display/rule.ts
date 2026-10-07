import type { RuleContext } from '../../context.ts';
import { entities } from './entities.ts';
import { missing, notRecommended } from './messages.ts';
function decode(text: string): string {
    let result = '';
    for(let index = 0; index < text.length;) {
        if(text.charCodeAt(index) !== 38) { result += text.slice(index, index + 1); index++; continue; }
        const semicolon = text.indexOf(';', index);
        if(semicolon < index + 2) { result += '&'; index++; continue; }
        const body = text.slice(index + 1, semicolon);
        let replacement = entities.get(body);
        if(body.startsWith('#')) {
            const hexadecimal = body.startsWith('#x');
            const digits = body.slice(hexadecimal ? 2 : 1);
            const pattern = hexadecimal ? /^[0-9a-fA-F]+$/ : /^[0-9]+$/;
            if(pattern.test(digits)) {
                const value = Number.parseInt(digits, hexadecimal ? 16 : 10);
                if(value <= 1114111) { replacement = value >= 55296 && value <= 57343 ? '\ufffd' : String.fromCodePoint(value); }
            }
        }
        if(replacement === undefined) { result += '&'; index++; }
        else { result += replacement; index = semicolon + 1; }
    }
    return result;
}
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number): void {
        const node = this.context.node(index);
        const tag = node.children[0] ?? -1;
        if(tag < 0 || this.context.node(tag).kind !== 'Identifier' || this.context.node(tag).text !== 'link') { return; }
        let reference = '';
        let found = false;
        for(const child of node.children) {
            if(this.context.node(child).kind !== 'JsxAttributes') { continue; }
            for(const property of this.context.node(child).children) {
                const attribute = this.context.node(property);
                if(attribute.kind !== 'JsxAttribute') { continue; }
                const name = attribute.children[0] ?? -1;
                if(name < 0 || this.context.node(name).kind !== 'Identifier' || this.context.node(name).text.toLowerCase() !== 'href') { continue; }
                const value = attribute.children[1] ?? -1;
                if(value >= 0 && this.context.node(value).kind === 'StringLiteral') { reference = decode(this.context.node(value).text); found = true; }
                break;
            }
        }
        if(!found || !reference.startsWith('https://fonts.googleapis.com/css')) { return; }
        const question = reference.indexOf('?');
        let display = '';
        let hasDisplay = false;
        if(question >= 0) {
            for(const pair of reference.slice(question + 1).split('&')) {
                const equals = pair.indexOf('=');
                if(equals >= 0 && pair.slice(0, equals) === 'display') { display = pair.slice(equals + 1); hasDisplay = true; break; }
            }
        }
        if(!hasDisplay) { this.context.report(index, '@next/next/google-font-display', 'googleFontDisplayMissing', missing, '', '', ''); }
        else if(display === 'auto' || display === 'block' || display === 'fallback') { this.context.report(index, '@next/next/google-font-display', 'googleFontDisplayNotRecommended', notRecommended, '', '', ''); }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
