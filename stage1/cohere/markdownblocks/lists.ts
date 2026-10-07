// Cohere markdown/list.go and list-related utilities/preprocess, with explicit inputs.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface ListBlockInterface {
    readonly kind: string;
    readonly doc: number;
    readonly start: number;
    readonly end: number;
    readonly column: number;
    readonly indented: boolean;
    readonly ignoreNext: boolean;
}
export interface ListItemInterface {
    readonly marker: string;
    readonly start: number;
    readonly end: number;
    readonly spread: boolean;
    readonly checked: number;
    readonly children: readonly ListBlockInterface[];
}
export interface ListFrameInterface {
    readonly ordered: boolean;
    readonly start: number;
    readonly sibling: number;
    readonly ancestorAligned: boolean;
    readonly nextIndented: boolean;
    readonly nextCode: string;
    readonly items: readonly ListItemInterface[];
}
function whitespace(code: number): boolean {
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
export interface OrderedMarkerInterface {
    readonly number: number;
    readonly spaces: string;
}
export function orderedMarker(text: string): OrderedMarkerInterface {
    let index = 0;
    while(index < text.length && whitespace(text.charCodeAt(index))) index++;
    const start = index;
    let number = 0;
    while(index < text.length && text.charCodeAt(index) >= 48 && text.charCodeAt(index) <= 57) {
        number = number * 10 + text.charCodeAt(index) - 48;
        index++;
    }
    if(index === start || (text.charCodeAt(index) !== 46 && text.charCodeAt(index) !== 41))
        panic('ordered list item without a number');
    index++;
    const spaceStart = index;
    while(index < text.length && whitespace(text.charCodeAt(index))) index++;
    return { number, spaces: text.slice(spaceStart, index) };
}
function gitFriendly(frame: ListFrameInterface): boolean {
    if(!frame.ordered || frame.items.length < 2) return false;
    const second = frame.items[1] ?? panic('second list item');
    if(orderedMarker(second.marker).number !== 1) return false;
    const first = frame.items[0] ?? panic('first list item');
    if(orderedMarker(first.marker).number !== 0) return true;
    return frame.items.length > 2 && orderedMarker((frame.items[2] ?? panic('third list item')).marker).number === 1;
}
export function alignedList(frame: ListFrameInterface, tabWidth: number): boolean {
    if(!frame.ancestorAligned || frame.nextIndented) return false;
    if(!frame.ordered || frame.items.length === 0) return true;
    const first = frame.items[0] ?? panic('first list item');
    if(orderedMarker(first.marker).spaces.length > 1) return true;
    const firstBlock = first.children[0];
    if(firstBlock === undefined) return false;
    const firstStart = firstBlock.column - 1;
    if(frame.items.length === 1) return firstStart % tabWidth === 0;
    const second = frame.items[1] ?? panic('second list item');
    const secondBlock = second.children[0];
    const secondStart = secondBlock === undefined ? -1 : secondBlock.column - 1;
    if(firstStart !== secondStart) return false;
    return firstStart % tabWidth === 0 || orderedMarker(second.marker).spaces.length > 1;
}
function requiredIndent(frame: ListFrameInterface): number {
    if(!frame.nextIndented) return 0;
    let count = 0;
    for(let index = 0; index < frame.nextCode.length; index++) {
        const code = frame.nextCode.charCodeAt(index);
        if(code === 9) count += 4;
        else if(code === 32) count++;
        else break;
    }
    return 5 + count;
}
export function listPrefix(frame: ListFrameInterface, index: number, tabWidth: number): string {
    let prefix: string;
    if(frame.ordered) {
        const number = index === 0 ? frame.start : gitFriendly(frame) ? 1 : Math.min(frame.start + index, 999999999);
        prefix = `${number}${frame.sibling % 2 === 0 ? '. ' : ') '}`;
        if(alignedList(frame, tabWidth)) {
            const rest = prefix.length % tabWidth;
            const spaces = rest === 0 ? 0 : tabWidth - rest;
            if(spaces < 4) prefix += ' '.repeat(spaces);
        }
    }
    else prefix = frame.sibling % 2 === 0 ? '- ' : '* ';
    const required = requiredIndent(frame);
    if(prefix.length >= required) return prefix;
    while(prefix.length > 0 && whitespace(prefix.charCodeAt(prefix.length - 1))) prefix = prefix.slice(0, -1);
    prefix += ' '.repeat(Math.min(required - prefix.length, 4));
    prefix = ' '.repeat(Math.min(required - prefix.length, 3)) + prefix;
    return prefix;
}
function loose(item: ListItemInterface, next: ListItemInterface | undefined): boolean {
    return item.spread || (next !== undefined && item.end + 1 < next.start);
}
function doubleLine(
    item: ListItemInterface,
    next: ListItemInterface | undefined,
    previous: ListBlockInterface,
    block: ListBlockInterface,
): boolean {
    if(block.start !== block.end && block.start < previous.end) return false;
    if(
        block.kind === 'list' &&
        (previous.kind === 'code' || previous.kind === 'paragraph') &&
        previous.end + 1 < block.start
    )
        return true;
    if(block.kind === 'definition' && previous.kind === 'definition') return false;
    if(block.kind === 'list' || !loose(item, next)) return false;
    if(previous.ignoreNext) return false;
    if(
        block.kind === 'html' &&
        (previous.kind === 'html' || previous.kind === 'paragraph') &&
        previous.end + 1 === block.start
    )
        return false;
    return true;
}
export function printList(arena: DocumentArena, frame: ListFrameInterface, tabWidth: number): number {
    const parts: number[] = [];
    for(let index = 0; index < frame.items.length; index++) {
        const item = frame.items[index] ?? panic('list item');
        const next = frame.items[index + 1];
        if(index > 0) {
            const previous = frame.items[index - 1] ?? panic('previous list item');
            parts.push(arena.hardline());
            if(loose(previous, item)) parts.push(arena.hardline());
        }
        const prefix = listPrefix(frame, index, tabWidth);
        const task = item.checked < 0 ? '' : item.checked === 1 ? '[x] ' : '[ ] ';
        const itemParts: number[] = [arena.text(task)];
        for(let blockIndex = 0; blockIndex < item.children.length; blockIndex++) {
            const block = item.children[blockIndex] ?? panic('list block');
            if(blockIndex > 0) {
                const previous = item.children[blockIndex - 1] ?? panic('previous list block');
                itemParts.push(arena.hardline());
                if(doubleLine(item, next, previous, block)) itemParts.push(arena.hardline());
            }
            if((blockIndex === 0 && block.kind !== 'list') || (block.kind === 'code' && block.indented))
                itemParts.push(arena.align(' '.repeat(task.length), block.doc));
            else {
                const alignment = ' '.repeat(Math.max(0, Math.min(tabWidth - prefix.length, 3)));
                itemParts.push(arena.concat([arena.text(alignment), arena.align(alignment, block.doc)]));
            }
        }
        const body = arena.concat(itemParts);
        parts.push(arena.text(prefix));
        const first = item.children[0];
        const second = item.children[1];
        const specialHTML =
            item.children.length === 2 &&
            first !== undefined &&
            second !== undefined &&
            second.kind === 'html' &&
            first.column !== second.column;
        parts.push(specialHTML ? body : arena.align(' '.repeat(prefix.length), body));
    }
    return arena.concat(parts);
}
