// Cohere printChildren for block wrappers outside list items and the root.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
import type { ListBlockInterface } from './lists.ts';
export function printBlockChildren(arena: DocumentArena, children: readonly ListBlockInterface[]): number {
    const parts: number[] = [];
    for(let index = 0; index < children.length; index++) {
        const child = children[index] ?? panic('block child');
        if(index > 0) {
            const previous = children[index - 1] ?? panic('previous block child');
            parts.push(arena.hardline());
            const overlapping = child.start !== child.end && child.start < previous.end;
            const definitions = child.kind === 'definition' && previous.kind === 'definition';
            const adjacentHTML =
                child.kind === 'html' &&
                (previous.kind === 'html' || previous.kind === 'paragraph') &&
                previous.end + 1 === child.start;
            if(!overlapping && !definitions && !previous.ignoreNext && !adjacentHTML) parts.push(arena.hardline());
        }
        parts.push(child.doc);
    }
    return arena.concat(parts);
}
