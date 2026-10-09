// printMdast's shouldRemainTheSameContent branch, with preserved prose.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
import { splitText } from './splitText.ts';
export function printPreservedLabel(arena: DocumentArena, source: string): number {
    const parts: number[] = [arena.text('')];
    for(const token of splitText(source)) {
        if(token.type === 'whitespace' && token.value === '\n') {
            parts.push(arena.hardline());
            parts.push(arena.text(''));
        }
        else {
            const value = token.type === 'word' ? token.value : token.value === ' ' ? ' ' : '';
            const last = parts.length - 1;
            parts[last] = arena.concat([parts[last] ?? panic('reference label part'), arena.text(value)]);
        }
    }
    return arena.add('f', '', 0, parts);
}
