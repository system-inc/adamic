// Leading triple-slash pragmas follow typescript-go, including incomplete attributes.
import { utf8Length } from 'adamic';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import type { Arena } from './arena.ts';
function blanks(text: string, start: number): number {
    let pos = start;
    while(pos < text.length && [' ', '\t'].includes(text.slice(pos, pos + 1))) {
        pos++;
    }
    return pos;
}
function nameEnd(text: string, start: number): number {
    let pos = start;
    while(pos < text.length) {
        const code = text.charCodeAt(pos);
        if(!((code >= 65 && code <= 90) || (code >= 97 && code <= 122) || code === 45)) {
            break;
        }
        pos++;
    }
    return pos;
}
export function referenceError(text: string, arena: Arena, comments: readonly number[]): string {
    const scanner = new Scanner(text);
    scanner.scan();
    const firstToken = utf8Length(text.slice(0, scanner.start));
    for(const id of comments) {
        const comment = arena.node(id);
        if(comment.type !== 'Line' || comment.end > firstToken) {
            continue;
        }
        const value = comment.string('value');
        if(!value.startsWith('/')) {
            continue;
        }
        let pos = blanks(value, 1);
        if(value.slice(pos, pos + 1) !== '<') {
            continue;
        }
        pos++;
        const tagEnd = nameEnd(value, pos);
        if(value.slice(pos, tagEnd).toLowerCase() !== 'reference') {
            continue;
        }
        pos = tagEnd;
        const attributes = new Map<string, string>();
        while(true) {
            pos = blanks(value, pos);
            if(value.slice(pos, pos + 2) === '/>') {
                break;
            }
            const end = nameEnd(value, pos);
            if(end === pos) {
                break;
            }
            const name = value.slice(pos, end).toLowerCase();
            pos = blanks(value, end);
            if(value.slice(pos, pos + 1) !== '=') {
                break;
            }
            pos = blanks(value, pos + 1);
            const quote = value.slice(pos, pos + 1);
            if(quote !== "'" && quote !== '"') {
                break;
            }
            const close = value.indexOf(quote, pos + 1);
            if(close < 0) {
                break;
            }
            attributes.set(name, value.slice(pos + 1, close));
            pos = close + 1;
        }
        if(attributes.get('no-default-lib') === 'true') {
            continue;
        }
        if(attributes.has('types')) {
            const mode = attributes.get('resolution-mode');
            if(mode !== undefined && mode !== 'import' && mode !== 'require') {
                return 'invalid reference resolution mode';
            }
        }
        else if(!attributes.has('lib') && !attributes.has('path')) {
            return 'invalid reference directive syntax';
        }
    }
    return '';
}
