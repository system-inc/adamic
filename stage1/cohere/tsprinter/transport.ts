import { panic } from 'adamic';
export function escaped(text: string): string {
    return text.split('\\').join('\\\\').split('\n').join('\\n').split('\r').join('\\r').split('\t').join('\\t');
}
export function unescaped(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) !== 92) continue;
        parts.push(text.slice(start, index));
        index++;
        const code = text.slice(index, index + 1);
        if(code !== 'n' && code !== 'r' && code !== 't' && code !== '\\') panic('invalid transport escape');
        parts.push(code === 'n' ? '\n' : code === 'r' ? '\r' : code === 't' ? '\t' : '\\');
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
