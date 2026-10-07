// Escaped-line protocol shared by native parser probes. One prefix marks empty inputs.
export function decode(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) !== 92) continue;
        parts.push(text.slice(start, index));
        index++;
        const letter = text.slice(index, index + 1);
        parts.push(letter === 'n' ? '\n' : letter === 'r' ? '\r' : letter === 't' ? '\t' : letter);
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
export function encode(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code !== 92 && code !== 10 && code !== 13 && code !== 9) continue;
        parts.push(text.slice(start, index));
        parts.push(code === 92 ? '\\\\' : code === 10 ? '\\n' : code === 13 ? '\\r' : '\\t');
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
