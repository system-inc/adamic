import { pathIndex, pathProperty } from './astPath.ts';
import type { AstPath, PathFrameInterface, PathPredicateInterface } from './astPath.ts';
export function frameText(frame: PathFrameInterface): string {
    if(frame.kind === 'node') return `N${frame.id}`;
    if(frame.kind === 'array') return `A${frame.id}`;
    if(frame.kind === 'index') return `I${frame.id}`;
    if(frame.kind === 'property') return `P${frame.name}`;
    return 'U';
}
const visits: number[] = [];
function identity(path: AstPath, index: number, array: PathFrameInterface): number {
    return (path.node() + 1) * 1000000 + (index + 1) * 1000 + array.id;
}
function visit(path: AstPath, index: number, array: PathFrameInterface): void {
    visits.push(identity(path, index, array));
}
const zeroPredicates: PathPredicateInterface[] = [];
function directNode(path: AstPath): number {
    return path.node();
}
function directFirst(path: AstPath): boolean {
    return path.isFirst();
}
function even(id: number): boolean {
    return id % 2 === 0;
}
function someNode(value: PathFrameInterface, name: PathFrameInterface, index: number): boolean {
    return value.kind === 'node' && (name.kind === 'missing' || name.kind === 'property') && index >= -1;
}
function neverNode(value: PathFrameInterface, name: PathFrameInterface, index: number): boolean {
    return value.kind === 'node' && name.kind === 'property' && index === -9;
}
export function pathSnapshot(path: AstPath): string {
    const stack: string[] = [];
    for(const frame of path.stack) stack.push(frameText(frame));
    const nodes: string[] = [];
    for(let count = 0; count < 5; count++) nodes.push(`${path.getNode(count)}`);
    return [
        stack.join(','),
        frameText(path.value()),
        path.key(),
        `${path.index()}`,
        frameText(path.name()),
        `${path.node()}`,
        `${path.parent()}`,
        `${path.grandparent()}`,
        `${path.root()}`,
        `${path.next()}`,
        `${path.previous()}`,
        `${path.isRoot()}`,
        `${path.isInArray()}`,
        `${path.isFirst()}`,
        `${path.isLast()}`,
        path.ancestors().join(','),
        nodes.join(','),
        `${path.getParentNode(1)}`,
        `${path.findAncestor(even)}`,
        `${path.hasAncestor(even)}`,
        `${path.match(zeroPredicates)}`,
        `${path.match([{ present: false, test: someNode }])}`,
        `${path.match([
            { present: true, test: someNode },
            { present: false, test: someNode },
        ])}`,
        `${path.match([
            { present: true, test: someNode },
            { present: true, test: someNode },
        ])}`,
        `${path.match([{ present: true, test: neverNode }])}`,
    ].join('|');
}
export function observePath(path: AstPath): string {
    const rows = [pathSnapshot(path)];
    rows.push(path.call(pathSnapshot, [pathProperty('children')]));
    rows.push(path.call(pathSnapshot, [pathProperty('children'), pathIndex(-1)]));
    rows.push(path.call(pathSnapshot, [pathProperty('children'), pathIndex(999999)]));
    rows.push(path.call(pathSnapshot, [pathProperty('absent')]));
    for(let count = 0; count < Math.min(3, path.ancestors().length); count++) {
        rows.push(path.callParent(pathSnapshot, count));
        rows.push(`${path.callParentNumber(directNode, count)}`);
        rows.push(`${path.callParentBoolean(directFirst, count)}`);
    }
    rows.push(`${path.callNumber(directNode, [pathProperty('children'), pathIndex(0)])}`);
    rows.push(pathSnapshot(path));
    rows.push(path.map(identity, [pathProperty('children')]).join(','));
    while(visits.length > 0) visits.pop();
    path.each(visit, [pathProperty('children')]);
    rows.push(visits.join(','));
    rows.push(pathSnapshot(path));
    return rows.join('\n');
}
