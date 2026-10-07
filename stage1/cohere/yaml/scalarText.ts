// Flow scalar folding and escape codes from yaml 2.9.0.
export function char(source: string, index: number): string {
    return source.slice(index, index + 1);
}
export function whitespace(character: string): boolean {
    return character === ' ' || character === '\t';
}
function lineEnd(source: string, from: number, newline: number): number {
    let end = newline;
    if(end - 1 >= from && char(source, end - 1) === '\r') end--;
    while(end - 1 >= from && whitespace(char(source, end - 1))) end--;
    return end;
}
export function foldLines(source: string): string {
    const first = source.indexOf('\n');
    if(first < 0) return source;
    let result = source.slice(0, lineEnd(source, 0, first));
    let separator = ' ';
    let position = first + 1;
    while(true) {
        let start = position;
        while(start < source.length && whitespace(char(source, start))) start++;
        const newline = source.indexOf('\n', start);
        if(newline < 0) break;
        const match = source.slice(start, lineEnd(source, start, newline));
        if(match === '') {
            if(separator === '\n') result += separator;
            else separator = '\n';
        }
        else {
            result += separator;
            result += match;
            separator = ' ';
        }
        position = newline + 1;
    }
    while(position < source.length && whitespace(char(source, position))) position++;
    result += separator;
    result += source.slice(position);
    return result;
}
export function escapeCode(character: string): string {
    switch(character) {
        case '0':
            return '\x00';
        case 'a':
            return '\x07';
        case 'b':
            return '\b';
        case 'e':
            return '\x1b';
        case 'f':
            return '\f';
        case 'n':
            return '\n';
        case 'r':
            return '\r';
        case 't':
            return '\t';
        case 'v':
            return '\v';
        case 'N':
            return '\x85';
        case '_':
            return '\xa0';
        case 'L':
            return '\u2028';
        case 'P':
            return '\u2029';
        case ' ':
        case '"':
        case '/':
        case '\\':
        case '\t':
            return character;
    }
    return '';
}
export function emptyContent(content: string): boolean {
    return content === '' || content === '\r';
}
