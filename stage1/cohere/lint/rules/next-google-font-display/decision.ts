import { panic } from 'adamic';
import { names, values } from './entities.ts';

export function decode(text: string): string {
    let result = '';
    for(let index = 0; index < text.length;) {
        if(text[index] !== '&') { result += text.slice(index, index + 1); index++; continue; }
        const semicolon = text.indexOf(';', index);
        if(semicolon < index + 2) { result += '&'; index++; continue; }
        const body = text.slice(index + 1, semicolon);
        let replacement = '';
        if(/^#(?:[0-9]+|x[0-9a-fA-F]+)$/.test(body)) {
            const value = Number.parseInt(body.slice(body.startsWith('#x') ? 2 : 1), body.startsWith('#x') ? 16 : 10);
            replacement = value > 1114111 ? text.slice(index, semicolon + 1) : String.fromCodePoint(value >= 55296 && value <= 57343 ? 65533 : value);
        }
        else if(/^[0-9a-zA-Z]+$/.test(body)) {
            const found = names.indexOf(body);
            replacement = found < 0 ? text.slice(index, semicolon + 1) : values[found] ?? panic('missing entity');
        }
        if(replacement === '') { result += '&'; index++; continue; }
        result += replacement; index = semicolon + 1;
    }
    return result;
}

export function diagnostic(raw: string): string {
    const href = decode(raw);
    if(!href.startsWith('https://fonts.googleapis.com/css')) { return ''; }
    const question = href.indexOf('?');
    if(question >= 0) {
        for(const pair of href.slice(question + 1).split('&')) {
            const equals = pair.indexOf('=');
            if(equals < 0 || pair.slice(0, equals) !== 'display') { continue; }
            const value = pair.slice(equals + 1);
            return ['auto', 'block', 'fallback'].includes(value) ? 'googleFontDisplayNotRecommended' : '';
        }
    }
    return 'googleFontDisplayMissing';
}
