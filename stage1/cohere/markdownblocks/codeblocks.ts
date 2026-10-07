// Cohere print.go code blocks, with embeddedLanguageFormatting off.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface CodeBlockInterface {
    readonly indented: boolean;
    readonly value: string;
    readonly language: string;
    readonly metadata: string;
}
export function fenceStyle(value: string): string {
    let longest = 0;
    let count = 0;
    for(let index = 0; index < value.length; index++) {
        if(value.charCodeAt(index) === 96) {
            count++;
            longest = Math.max(longest, count);
        }
        else count = 0;
    }
    return '`'.repeat(Math.max(3, longest + 1));
}
export function printCodeBlock(arena: DocumentArena, frame: CodeBlockInterface): number {
    const body: number[] = [];
    const lines = frame.value.split('\n');
    for(let index = 0; index < lines.length; index++) {
        if(index > 0) body.push(arena.hardline());
        body.push(arena.text(lines[index] ?? panic('code line')));
    }
    if(frame.indented) {
        const alignment = ' '.repeat(4);
        return arena.align(alignment, arena.concat([arena.text(alignment), arena.concat(body)]));
    }
    const style = fenceStyle(frame.value);
    const metadata = frame.metadata === '' ? '' : ` ${frame.metadata}`;
    return arena.concat([
        arena.text(style),
        arena.text(frame.language),
        arena.text(metadata),
        arena.hardline(),
        arena.concat(body),
        arena.hardline(),
        arena.text(style),
    ]);
}
