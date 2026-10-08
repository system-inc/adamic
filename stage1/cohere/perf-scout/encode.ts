// The lint wire protocol uses ASCII and escaped UTF-16 units. Copy ordinary spans once.
export function written(text: string): string {
    let result = '';
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code < 32 || code > 126 || code === 92) {
            if(index > start) result += text.slice(start, index);
            result += `\\u${code.toString(16).padStart(4, '0')}`;
            start = index + 1;
        }
    }
    if(start === 0) return text;
    if(start < text.length) result += text.slice(start);
    return result;
}
