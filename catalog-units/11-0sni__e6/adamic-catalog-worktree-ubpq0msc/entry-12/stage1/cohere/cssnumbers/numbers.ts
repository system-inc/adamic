// CSS numeric adjustment from cohere/internal/format/css/print_misc.go.
// No parser or regex runtime. Only ASCII digits and letters are special.
import { adjustStrings } from '../cssstrings/strings.ts';

// css-units-list 2.1.0; constant cases avoid repeated table-entry lowercase allocations.
function canonicalUnit(raw: string): string {
    if(raw === '') return '';
    const lower = raw.toLowerCase();
    switch(lower) {
        case 'q':
            return 'Q';
        case 'hz':
            return 'Hz';
        case 'khz':
            return 'kHz';
        case 'em':
        case 'rem':
        case 'ex':
        case 'rex':
        case 'cap':
        case 'rcap':
        case 'ch':
        case 'rch':
        case 'ic':
        case 'ric':
        case 'lh':
        case 'rlh':
        case 'vw':
        case 'svw':
        case 'lvw':
        case 'dvw':
        case 'vh':
        case 'svh':
        case 'lvh':
        case 'dvh':
        case 'vi':
        case 'svi':
        case 'lvi':
        case 'dvi':
        case 'vb':
        case 'svb':
        case 'lvb':
        case 'dvb':
        case 'vmin':
        case 'svmin':
        case 'lvmin':
        case 'dvmin':
        case 'vmax':
        case 'svmax':
        case 'lvmax':
        case 'dvmax':
        case 'cm':
        case 'mm':
        case 'in':
        case 'pt':
        case 'pc':
        case 'px':
        case 'deg':
        case 'grad':
        case 'rad':
        case 'turn':
        case 's':
        case 'ms':
        case 'dpi':
        case 'dpcm':
        case 'dppx':
        case 'x':
        case 'cqw':
        case 'cqh':
        case 'cqi':
        case 'cqb':
        case 'cqmin':
        case 'cqmax':
        case 'fr':
            return lower;
        default:
            return '';
    }
}
function digit(code: number): boolean {
    return code >= 48 && code <= 57;
}
function letter(code: number): boolean {
    return (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
}
function wordStart(code: number): boolean {
    return code === 95 || letter(code) || code >= 128;
}
function wordPart(code: number): boolean {
    return wordStart(code) || digit(code) || code === 45;
}
function skipDigits(value: string, start: number): number {
    let index = start;
    while(index < value.length && digit(value.charCodeAt(index))) index++;
    return index;
}
function numberEnd(value: string, start: number): number {
    const integerEnd = skipDigits(value, start);
    let end: number;
    if(value.charCodeAt(integerEnd) === 46 && digit(value.charCodeAt(integerEnd + 1))) {
        end = skipDigits(value, integerEnd + 1);
    }
    else if(integerEnd > start) {
        end = integerEnd;
        if(value.charCodeAt(end) === 46) end++;
    }
    else {
        return -1;
    }
    if(value.charCodeAt(end) === 69 || value.charCodeAt(end) === 101) {
        let exponent = end + 1;
        if(value.charCodeAt(exponent) === 43 || value.charCodeAt(exponent) === 45) exponent++;
        if(digit(value.charCodeAt(exponent))) end = skipDigits(value, exponent);
    }
    return end;
}
// On the matched `NUMBER` grammar, these operations are printNumber followed by
// printCssNumber: normalize the exponent, trim fraction zeros, and keep integer zeros.
function printNumber(raw: string): string {
    const value = raw.toLowerCase();
    const exponentPosition = value.indexOf('e');
    let mantissa = exponentPosition < 0 ? value : value.slice(0, exponentPosition);
    let suffix = '';
    if(exponentPosition >= 0) {
        const exponent = value.slice(exponentPosition + 1);
        let start = exponent.startsWith('+') || exponent.startsWith('-') ? 1 : 0;
        while(exponent.charCodeAt(start) === 48) start++;
        if(start < exponent.length) suffix = `e${exponent.startsWith('-') ? '-' : ''}${exponent.slice(start)}`;
    }
    if(mantissa.startsWith('.')) mantissa = `0${mantissa}`;
    const dot = mantissa.indexOf('.');
    if(dot >= 0) {
        let end = mantissa.length;
        while(end > dot + 1 && mantissa.charCodeAt(end - 1) === 48) end--;
        if(end === dot + 1) end = dot;
        mantissa = mantissa.slice(0, end);
    }
    return mantissa + suffix;
}
function stringEnd(value: string, start: number): number {
    const quote = value.charCodeAt(start);
    if(quote !== 34 && quote !== 39) return -1;
    for(let index = start + 1; index < value.length; index++) {
        const character = value.charCodeAt(index);
        if(character === quote) return index + 1;
        if(character === 92) index++;
    }
    return -1;
}
type MatchType = {
    readonly end: number;
    readonly numberStart: number;
    readonly numberEnd: number;
    readonly hasWordPart: boolean;
};
function finish(value: string, start: number, end: number, hasWordPart: boolean): MatchType {
    let unitEnd = end;
    while(unitEnd < value.length && letter(value.charCodeAt(unitEnd))) unitEnd++;
    return { end: unitEnd, numberStart: start, numberEnd: end, hasWordPart };
}
function numberMatch(value: string, start: number): MatchType {
    let word = start;
    if(value.charCodeAt(word) === 36 || value.charCodeAt(word) === 64) word++;
    if(word < value.length && wordStart(value.charCodeAt(word))) {
        let end = word + 1;
        while(end < value.length && wordPart(value.charCodeAt(end))) end++;
        for(let number = end; number > word; number--) {
            const next = numberEnd(value, number);
            if(next >= 0) return finish(value, number, next, true);
        }
    }
    const end = numberEnd(value, start);
    return end < 0
        ? { end: -1, numberStart: start, numberEnd: start, hasWordPart: false }
        : finish(value, start, end, false);
}
export function adjustNumbers(value: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < value.length; index++) {
        const quotedEnd = stringEnd(value, index);
        if(quotedEnd >= 0) {
            index = quotedEnd - 1;
            continue;
        }
        const match = numberMatch(value, index);
        if(match.end < 0) continue;
        if(!match.hasWordPart) {
            const rawUnit = value.slice(match.numberEnd, match.end);
            const unit = canonicalUnit(rawUnit);
            if(rawUnit === '' || rawUnit.toLowerCase() === 'n' || unit !== '') {
                parts.push(value.slice(start, index));
                parts.push(printNumber(value.slice(match.numberStart, match.numberEnd)));
                parts.push(unit === '' ? rawUnit.toLowerCase() : unit);
                start = match.end;
            }
        }
        index = match.end - 1;
    }
    parts.push(value.slice(start));
    return parts.join('');
}
export function formatValue(value: string, single: boolean): string {
    return adjustNumbers(adjustStrings(value, single));
}
