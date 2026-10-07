// Cohere heading.go, sentence.go and paragraph.go on owned document IDs.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface SentenceChildInterface {
    readonly doc: number;
    readonly whitespace: boolean;
}
export function printSentence(arena: DocumentArena, children: readonly SentenceChildInterface[]): number {
    const parts: number[] = [];
    let contents: number[] = [arena.text('')];
    for(const child of children) {
        if(child.whitespace && arena.node(child.doc).kind !== 't') {
            parts.push(contents.length === 1 ? (contents[0] ?? panic('sentence head')) : arena.concat(contents));
            parts.push(child.doc);
            contents = [arena.text('')];
        }
        else contents.push(child.doc);
    }
    parts.push(contents.length === 1 ? (contents[0] ?? panic('sentence tail')) : arena.concat(contents));
    return arena.add('f', '', 0, parts);
}
export function printParagraph(arena: DocumentArena, children: readonly number[]): number {
    const parts: number[] = [arena.text('')];
    const stack = children.slice().reverse();
    while(stack.length > 0) {
        const id = stack.pop() ?? panic('paragraph stack');
        const node = arena.node(id);
        if(node.kind === 'a') {
            for(let index = node.children.length - 1; index >= 0; index--)
                stack.push(node.children[index] ?? panic('paragraph concat child'));
            continue;
        }
        const head = node.kind === 'f' ? (node.children[0] ?? panic('fill head')) : id;
        const last = parts.length - 1;
        parts[last] = arena.concat([parts[last] ?? panic('paragraph part'), head]);
        if(node.kind === 'f') {
            for(let index = 1; index < node.children.length; index++)
                parts.push(node.children[index] ?? panic('fill rest'));
        }
    }
    return arena.add('f', '', 0, parts);
}
export function printHeading(
    arena: DocumentArena,
    children: readonly number[],
    depth: number,
    setext: boolean,
    source: string,
): number {
    const contents = arena.concat(children);
    if(!setext) return arena.concat([arena.text(`${'#'.repeat(depth)} `), contents]);
    const lastLine = source.slice(source.lastIndexOf('\n') + 1);
    const start = Math.max(lastLine.indexOf('='), lastLine.indexOf('-'));
    const underline = lastLine.slice(start < 0 ? Math.max(lastLine.length - 1, 0) : start);
    return arena.concat([contents, arena.hardline(), arena.text(underline)]);
}
