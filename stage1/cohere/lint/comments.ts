// Matching follows cohere's JavaScript iu pattern with literal terms and decorations.
// Quoting counts UTF-8 bytes, as cohere does, rather than JS code units.
import { utf8Length } from 'adamic';
import { printable } from './unicode.ts';
import { hasWarningRuleName } from './regex/no_warning_comments.a';

export function word(character: string): boolean {
    return (
        character !== '' &&
        ((character >= 'a' && character <= 'z') ||
            (character >= 'A' && character <= 'Z') ||
            (character >= '0' && character <= '9') ||
            character === '_')
    );
}
export function space(character: string): boolean {
    const code = character.codePointAt(0) ?? -1;
    return (
        (code >= 9 && code <= 13) ||
        code === 32 ||
        code === 160 ||
        code === 5760 ||
        (code >= 8192 && code <= 8202) ||
        code === 8232 ||
        code === 8233 ||
        code === 8239 ||
        code === 8287 ||
        code === 12288 ||
        code === 65279
    );
}
export function selfDirective(value: string): boolean {
    if(!hasWarningRuleName(value)) return false;
    let start = 0;
    while(start < value.length && space(value[start] ?? '')) {
        start++;
    }
    const trimmed = value.slice(start);
    for(const prefix of [
        'eslint',
        'eslint-disable',
        'eslint-enable',
        'eslint-disable-line',
        'eslint-disable-next-line',
        'globals',
        'exported',
    ]) {
        if(trimmed.startsWith(prefix) && ['', ' ', '\t', '\n', '\r'].includes(trimmed[prefix.length] ?? '')) {
            return true;
        }
    }
    return false;
}
export { matches } from './regex/no_warning_comments.a';
export function quote(value: string): string {
    let normalized = '';
    let cursor = 0;
    while(cursor < value.length) {
        while(cursor < value.length && space(value[cursor] ?? '')) {
            cursor++;
        }
        const start = cursor;
        while(cursor < value.length && !space(value[cursor] ?? '')) {
            cursor++;
        }
        if(start === cursor) {
            break;
        }
        const wordText = value.slice(start, cursor);
        const candidate = normalized === '' ? wordText : `${normalized} ${wordText}`;
        if(utf8Length(candidate) > 40) {
            normalized += '...';
            break;
        }
        normalized = candidate;
    }
    let result = '"';
    for(let index = 0; index < normalized.length; index++) {
        const point = normalized.codePointAt(index) ?? 0;
        if(point > 65535) {
            index++;
        }
        if(point === 34) {
            result += '\\"';
        }
        else if(point === 92) {
            result += '\\\\';
        }
        else if(point === 7) {
            result += '\\a';
        }
        else if(point === 8) {
            result += '\\b';
        }
        else if(point === 9) {
            result += '\\t';
        }
        else if(point === 10) {
            result += '\\n';
        }
        else if(point === 11) {
            result += '\\v';
        }
        else if(point === 12) {
            result += '\\f';
        }
        else if(point === 13) {
            result += '\\r';
        }
        else if(point < 32 || point === 127) {
            result += `\\x${point.toString(16).padStart(2, '0')}`;
        }
        else if(!printable(point)) {
            result +=
                point <= 65535
                    ? `\\u${point.toString(16).padStart(4, '0')}`
                    : `\\U${point.toString(16).padStart(8, '0')}`;
        }
        else {
            result += String.fromCodePoint(point);
        }
    }
    return `${result}"`;
}
