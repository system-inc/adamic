// Test transport projects every field the preprocessing passes read or write.
import { panic } from 'adamic';
import type { AstArena, AstNodeInterface } from './astArena.ts';
import { AstWalk } from './astWalk.ts';
import { decode, encode } from './codec.ts';
export function astInteger(text: string): number {
    const value = Number.parseInt(text, 10);
    if(!Number.isInteger(value)) panic('AST fixture integer');
    return value;
}
function flags(node: AstNodeInterface): number {
    return (
        (node.hasRaw ? 1 : 0) +
        (node.parent ? 2 : 0) +
        (node.ordered ? 4 : 0) +
        (node.indented ? 8 : 0) +
        (node.aligned ? 16 : 0) +
        (node.hasOriginalAlt ? 32 : 0) +
        (node.literal ? 64 : 0) +
        (node.cj ? 128 : 0) +
        (node.leading ? 256 : 0) +
        (node.trailing ? 512 : 0) +
        (node.hasPosition ? 1024 : 0)
    );
}
export function readAstNode(arena: AstArena, fields: readonly string[]): void {
    const id = arena.add(fields[1] ?? '');
    const node = arena.node(id);
    const mask = astInteger(fields[3] ?? '');
    node.value = decode(fields[2] ?? '');
    node.hasRaw = (mask & 1) !== 0;
    node.parent = (mask & 2) !== 0;
    node.ordered = (mask & 4) !== 0;
    node.indented = (mask & 8) !== 0;
    node.aligned = (mask & 16) !== 0;
    node.hasOriginalAlt = (mask & 32) !== 0;
    node.literal = (mask & 64) !== 0;
    node.cj = (mask & 128) !== 0;
    node.leading = (mask & 256) !== 0;
    node.trailing = (mask & 512) !== 0;
    node.hasPosition = (mask & 1024) !== 0;
    node.raw = decode(fields[4] ?? '');
    node.originalAlt = decode(fields[5] ?? '');
    node.position = astInteger(fields[6] ?? '');
    node.start = astInteger(fields[7] ?? '');
    node.end = astInteger(fields[8] ?? '');
    node.startLine = astInteger(fields[9] ?? '');
    node.endLine = astInteger(fields[10] ?? '');
    node.startColumn = astInteger(fields[11] ?? '');
    node.endColumn = astInteger(fields[12] ?? '');
    node.wordKind = fields[13] ?? '';
    if((fields[14] ?? '') !== '')
        for(const child of (fields[14] ?? '').split(',')) node.children.push(astInteger(child));
    arena.reservePosition(node.position);
}
export function serializeAst(arena: AstArena, root: number): string {
    const ids: number[] = [];
    const normalized = new Map<number, number>();
    const positions = new Map<number, number>();
    const walk = new AstWalk(arena, root);
    while(walk.next()) {
        if(normalized.has(walk.id)) panic('AST fixture is not a tree');
        normalized.set(walk.id, ids.length);
        ids.push(walk.id);
        walk.descend(walk.id);
    }
    const output: string[] = [];
    for(const id of ids) {
        const node = arena.node(id);
        const children: string[] = [];
        for(const child of node.children) children.push(`${normalized.get(child) ?? panic('serialized child')}`);
        let position = -1;
        if(node.hasPosition) {
            if(!positions.has(node.position)) positions.set(node.position, positions.size);
            position = positions.get(node.position) ?? panic('serialized position');
        }
        output.push(
            [
                'N',
                node.kind,
                encode(node.value),
                `${flags(node)}`,
                encode(node.raw),
                encode(node.originalAlt),
                `${position}`,
                `${node.start}`,
                `${node.end}`,
                `${node.startLine}`,
                `${node.endLine}`,
                `${node.startColumn}`,
                `${node.endColumn}`,
                node.wordKind,
                children.join(','),
            ].join('\t'),
        );
    }
    return output.join('\n');
}
