// markAlignedList's preorder, parent and following-code predicates.
import { panic } from 'adamic';
import type { AstArena } from './astArena.ts';
import type { AstWalk } from './astWalk.ts';
import { listSpaces } from './astSource.ts';
function contentStart(arena: AstArena, item: number): number {
    const child = arena.node(item).children[0];
    return child === undefined ? -1 : arena.node(child).startColumn - 1;
}
function spaces(arena: AstArena, item: number, source: string): number {
    const node = arena.node(item);
    const first = node.children[0];
    const end = first === undefined ? node.end : arena.node(first).start;
    const result = listSpaces(source.slice(node.start, end));
    if(result < 0) panic('markdown: an ordered list item without a number');
    return result;
}
function aligned(arena: AstArena, list: number, source: string, tabWidth: number): boolean {
    const node = arena.node(list);
    if(!node.ordered) return true;
    const first = node.children[0] ?? panic('first list item');
    if(spaces(arena, first, source) > 1) return true;
    const firstStart = contentStart(arena, first);
    if(firstStart === -1) return false;
    if(node.children.length === 1) return firstStart % tabWidth === 0;
    const second = node.children[1] ?? panic('second list item');
    if(firstStart !== contentStart(arena, second)) return false;
    if(firstStart % tabWidth === 0) return true;
    return spaces(arena, second, source) > 1;
}
export function markAlignedList(arena: AstArena, walk: AstWalk, source: string, tabWidth: number): void {
    const node = arena.node(walk.id);
    if(node.kind !== 'list' || node.children.length === 0) return;
    for(const parent of walk.ancestors) {
        if(arena.node(parent).kind === 'list' && !arena.node(parent).aligned) {
            node.aligned = false;
            return;
        }
    }
    if(walk.parent >= 0) {
        const next = arena.node(walk.parent).children[walk.index + 1];
        if(next !== undefined && arena.node(next).kind === 'code' && arena.node(next).indented) {
            node.aligned = false;
            return;
        }
    }
    node.aligned = aligned(arena, walk.id, source, tabWidth);
}
