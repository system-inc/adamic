// Cohere blockquote construction and its child-block spacing on explicit AST facts.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
import type { ListBlockInterface } from './lists.ts';
function inline(kind: string): boolean {
    return [
        'liquidNode',
        'inlineCode',
        'emphasis',
        'esComment',
        'strong',
        'delete',
        'wikiLink',
        'link',
        'linkReference',
        'image',
        'imageReference',
        'footnote',
        'footnoteReference',
        'sentence',
        'whitespace',
        'word',
        'break',
        'inlineMath',
    ].includes(kind);
}
export function printQuote(arena: DocumentArena, children: readonly ListBlockInterface[]): number {
    const parts: number[] = [];
    for(let index = 0; index < children.length; index++) {
        const child = children[index] ?? panic('quote child');
        if(index > 0 && !inline(child.kind)) {
            const previous = children[index - 1] ?? panic('previous quote child');
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
    return arena.concat([arena.text('> '), arena.align('> ', arena.concat(parts))]);
}
