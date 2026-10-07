// Source scans used by all six Markdown AST preprocessing passes, no RegExp.
export function javascriptSpace(code: number): boolean {
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
export function htmlSpace(code: number): boolean {
    return code === 9 || code === 10 || code === 12 || code === 13 || code === 32;
}
export function trimHTML(text: string, leading: boolean, trailing: boolean): string {
    let start = 0;
    let end = text.length;
    if(leading) while(start < end && htmlSpace(text.charCodeAt(start))) start++;
    if(trailing) while(end > start && htmlSpace(text.charCodeAt(end - 1))) end--;
    return text.slice(start, end);
}
function quoteEnd(text: string): number {
    let end = 0;
    let index = 0;
    while(index < text.length) {
        while(text.charCodeAt(index) === 32 || text.charCodeAt(index) === 9) index++;
        if(text.charCodeAt(index) !== 62) break;
        index++;
        while(text.charCodeAt(index) === 32 || text.charCodeAt(index) === 9) index++;
        end = index;
    }
    return end;
}
export function quoteRaw(text: string, value: string): string {
    const rawLines = text.split('\n');
    const valueLines = value.split('\n');
    const result: string[] = [];
    for(let index = 0; index < rawLines.length; index++) {
        const raw = rawLines[index] ?? '';
        const decoded = valueLines[index] ?? '';
        result.push(decoded.slice(0, quoteEnd(decoded)) + raw.slice(quoteEnd(raw)));
    }
    return result.join('\n');
}
export function bracketContent(text: string, start: number, end: number): string | undefined {
    const first = text.indexOf('[', start);
    if(first < 0 || first >= end) return undefined;
    let depth = 1;
    let index = first + 1;
    while(index < end) {
        const code = text.charCodeAt(index);
        if(code === 92) {
            index += 2;
            continue;
        }
        if(code === 91) depth++;
        else if(code === 93) {
            depth--;
            if(depth === 0) return text.slice(first + 1, index);
        }
        index++;
    }
    return undefined;
}
export function isIndented(text: string): boolean {
    const start = text.charCodeAt(0) === 10 ? 1 : 0;
    if(text.charCodeAt(start) === 9) return true;
    for(let index = 0; index < 4; index++) if(text.charCodeAt(start + index) !== 32) return false;
    return true;
}
export function listSpaces(text: string): number {
    let index = 0;
    while(index < text.length && javascriptSpace(text.charCodeAt(index))) index++;
    const start = index;
    while(text.charCodeAt(index) >= 48 && text.charCodeAt(index) <= 57) index++;
    if(index === start || (text.charCodeAt(index) !== 46 && text.charCodeAt(index) !== 41)) return -1;
    index++;
    const markerEnd = index;
    while(index < text.length && javascriptSpace(text.charCodeAt(index))) index++;
    return index - markerEnd;
}
