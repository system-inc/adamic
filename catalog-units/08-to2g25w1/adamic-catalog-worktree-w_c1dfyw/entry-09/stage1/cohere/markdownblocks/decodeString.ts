// Cohere micromark DecodeString: escapes and bounded, semicolon-terminated references.
import { panic } from 'adamic';
import { upperEntityNames } from './upperEntityNames.ts';
import { upperEntityValues } from './upperEntityValues.ts';
import { lowerEntityNames } from './lowerEntityNames.ts';
import { lowerEntityValues } from './lowerEntityValues.ts';
function punctuation(code: number): boolean {
    return (
        (code >= 33 && code <= 47) ||
        (code >= 58 && code <= 64) ||
        (code >= 91 && code <= 96) ||
        (code >= 123 && code <= 126)
    );
}
function digit(code: number, hexadecimal: boolean): boolean {
    return (code >= 48 && code <= 57) || (hexadecimal && ((code >= 65 && code <= 70) || (code >= 97 && code <= 102)));
}
function alphanumeric(code: number): boolean {
    return digit(code, false) || (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
}
function lookupReference(
    name: string,
    entityNames: readonly string[],
    entityValues: readonly string[],
): string | undefined {
    let low = 0;
    let high = entityNames.length;
    while(low < high) {
        const middle = Math.floor((low + high) / 2);
        if((entityNames[middle] ?? panic('entity name')) < name) low = middle + 1;
        else high = middle;
    }
    if(entityNames[low] !== name) return undefined;
    return entityValues[low] ?? panic('entity value');
}
export function namedReference(name: string): string | undefined {
    const first = name.charCodeAt(0);
    if(first >= 65 && first <= 90) return lookupReference(name, upperEntityNames, upperEntityValues);
    return lookupReference(name, lowerEntityNames, lowerEntityValues);
}
// Called only on digit runs validated and bounded by decodeString, like cohere.
function numericReference(value: string, base: number): string {
    const code = Number.parseInt(value, base);
    if(
        code < 9 ||
        code === 11 ||
        (code > 13 && code < 32) ||
        (code > 126 && code < 160) ||
        (code > 55295 && code < 57344) ||
        (code > 64975 && code < 65008) ||
        (code & 65535) === 65535 ||
        (code & 65535) === 65534 ||
        code > 1114111
    )
        return '�';
    return String.fromCodePoint(code);
}
export function decodeString(value: string): string {
    const output: string[] = [];
    let index = 0;
    let start = 0;
    while(index < value.length) {
        const code = value.charCodeAt(index);
        if(code === 92 && index + 1 < value.length && punctuation(value.charCodeAt(index + 1))) {
            output.push(value.slice(start, index));
            output.push(value.slice(index + 1, index + 2));
            index += 2;
            start = index;
            continue;
        }
        if(code === 38) {
            let cursor = index + 1;
            let decoded = '';
            let matched = false;
            if(value.charCodeAt(cursor) === 35) {
                cursor++;
                const hexadecimal = value.charCodeAt(cursor) === 120 || value.charCodeAt(cursor) === 88;
                if(hexadecimal) cursor++;
                const digits = cursor;
                while(cursor < value.length && digit(value.charCodeAt(cursor), hexadecimal)) cursor++;
                if(cursor > digits && cursor - digits <= (hexadecimal ? 6 : 7) && value.charCodeAt(cursor) === 59) {
                    decoded = numericReference(value.slice(digits, cursor), hexadecimal ? 16 : 10);
                    matched = true;
                }
            }
            else {
                const letters = cursor;
                while(cursor < value.length && alphanumeric(value.charCodeAt(cursor))) cursor++;
                if(cursor > letters && cursor - letters <= 31 && value.charCodeAt(cursor) === 59) {
                    const named = namedReference(value.slice(letters, cursor));
                    if(named !== undefined) {
                        decoded = named;
                        matched = true;
                    }
                }
            }
            if(matched) {
                output.push(value.slice(start, index));
                output.push(decoded);
                index = cursor + 1;
                start = index;
                continue;
            }
        }
        index++;
    }
    output.push(value.slice(start));
    return output.join('');
}
