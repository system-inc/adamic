// Matching follows cohere's JavaScript iu pattern with literal terms and decorations.
// Quoting counts UTF-8 bytes, as cohere does, rather than JS code units.
import { utf8Length } from 'adamic';
import { foldPoint, printable } from './unicode.ts';

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
    const name = 'no-warning-comments';
    let position = value.indexOf(name);
    let named = false;
    while(position >= 0) {
        if(!word(value[position - 1] ?? '') && !word(value[position + name.length] ?? '')) {
            named = true;
            break;
        }
        position = value.indexOf(name, position + 1);
    }
    if(!named) {
        return false;
    }
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
function decorated(character: string, decoration: string[]): boolean {
    for(const entry of decoration) {
        if(foldPoint(character.codePointAt(0) ?? -1) === foldPoint(entry.codePointAt(0) ?? -2)) {
            return true;
        }
    }
    return false;
}
// JavaScript iu boundaries include the long s and Kelvin sign through their ASCII folds.
function patternWord(character: string): boolean {
    return character !== '' && word(String.fromCodePoint(foldPoint(character.codePointAt(0) ?? 0)));
}
export function matches(value: string, term: string, location: string, decoration: string[]): boolean {
    // Cohere escapes every decoration, including '-' as \x2d, so none spells a range.
    let prefix = 0;
    if(location === 'start') {
        while(prefix < value.length) {
            const point = value.codePointAt(prefix) ?? 0;
            const character = String.fromCodePoint(point);
            if(!space(character) && !decorated(character, decoration)) {
                break;
            }
            prefix += point > 65535 ? 2 : 1;
        }
    }
    for(let index = 0; index <= (location === 'start' ? prefix : value.length); index++) {
        // Compare folded scalars directly, without allocating candidate strings.
        let end = index;
        let equal = true;
        for(let cursor = 0; cursor < term.length;) {
            const expected = term.codePointAt(cursor) ?? 0;
            const actual = value.codePointAt(end);
            if(actual === undefined || foldPoint(actual) !== foldPoint(expected)) {
                equal = false;
                break;
            }
            cursor += expected > 65535 ? 2 : 1;
            end += actual > 65535 ? 2 : 1;
        }
        if(
            equal &&
            (location === 'start' ||
                !word(term[0] ?? '') ||
                patternWord(value[index - 1] ?? '') !== patternWord(value[index] ?? '')) &&
            (!word(term[term.length - 1] ?? '') || patternWord(value[end - 1] ?? '') !== patternWord(value[end] ?? ''))
        ) {
            return true;
        }
        if((value.codePointAt(index) ?? 0) > 65535) {
            index++;
        }
    }
    return false;
}
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
