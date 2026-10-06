// Cohere print.go code blocks, with embeddedLanguageFormatting off.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface CodeBlockInterface {
    readonly indented: boolean;
    readonly value: string;
    readonly language: string;
    readonly metadata: string;
    readonly languageWidth: number;
    readonly metadataWidth: number;
    readonly lineWidths: readonly number[];
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
        body.push(arena.text(lines[index] ?? panic('code line'), frame.lineWidths[index] ?? panic('code line width')));
    }
    if(frame.indented) {
        const alignment = ' '.repeat(4);
        return arena.align(alignment, arena.concat([arena.text(alignment, alignment.length), arena.concat(body)]));
    }
    const style = fenceStyle(frame.value);
    const metadata = frame.metadata === '' ? '' : ` ${frame.metadata}`;
    return arena.concat([
        arena.text(style, style.length),
        arena.text(frame.language, frame.languageWidth),
        arena.text(metadata, frame.metadataWidth),
        arena.hardline(),
        arena.concat(body),
        arena.hardline(),
        arena.text(style, style.length),
    ]);
}
