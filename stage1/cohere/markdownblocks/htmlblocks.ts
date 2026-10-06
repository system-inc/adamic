// Cohere print.go HTML layout, including comments and the final root block.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface HTMLBlockInterface {
    readonly value: string;
    readonly rootLast: boolean;
}
function space(code: number): boolean {
    return (
        code === 9 ||
        code === 10 ||
        code === 11 ||
        code === 12 ||
        code === 13 ||
        code === 32 ||
        code === 0xa0 ||
        code === 0x1680 ||
        (code >= 0x2000 && code <= 0x200a) ||
        code === 0x2028 ||
        code === 0x2029 ||
        code === 0x202f ||
        code === 0x205f ||
        code === 0x3000 ||
        code === 0xfeff
    );
}
export function printHTMLBlock(arena: DocumentArena, frame: HTMLBlockInterface): number {
    let value = frame.value;
    if(frame.rootLast) {
        let end = value.length;
        while(end > 0 && space(value.charCodeAt(end - 1))) end--;
        value = value.slice(0, end);
    }
    const comment = value.length >= 7 && value.startsWith('<!--') && value.endsWith('-->');
    const literal = arena.concat([arena.add('h', '', 0, [], 6), arena.add('p', '', 0, [])]);
    const replacement = comment ? arena.hardline() : arena.add('r', '', 0, [literal]);
    const parts: number[] = [];
    const lines = value.split('\n');
    for(let index = 0; index < lines.length; index++) {
        if(index > 0) parts.push(replacement);
        parts.push(arena.text(lines[index] ?? panic('HTML line')));
    }
    return arena.concat(parts);
}
