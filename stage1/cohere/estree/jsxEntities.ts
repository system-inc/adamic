import { xhtmlEntities } from './xhtmlEntities.ts';
export function jsxText(text: string): string {
    const parts: string[] = [];
    let index = 0;
    while(index < text.length) {
        if(text.slice(index, index + 1) !== '&') {
            parts.push(text.slice(index, index + 1));
            index++;
            continue;
        }
        const end = text.indexOf(';', index + 1);
        if(end < index + 2) {
            parts.push('&');
            index++;
            continue;
        }
        const item = text.slice(index + 1, end);
        let replacement = xhtmlEntities.get(item);
        if(/^#(?:[0-9]+|x[0-9a-fA-F]+)$/.test(item)) {
            const value = Number.parseInt(item.slice(item.startsWith('#x') ? 2 : 1), item.startsWith('#x') ? 16 : 10);
            if(value <= 0x10ffff) {
                replacement = value >= 0xd800 && value <= 0xdfff ? '\ufffd' : String.fromCodePoint(value);
            }
        }
        if(replacement === undefined) {
            parts.push('&');
            index++;
        }
        else {
            parts.push(replacement);
            index = end + 1;
        }
    }
    return parts.join('');
}
