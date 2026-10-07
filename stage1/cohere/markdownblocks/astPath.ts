// printing/astpath.go, src/common/ast-path.js: Markdown's node/name/array stack.
import type { AstArena } from './astArena.ts';
export interface PathFrameInterface {
    readonly kind: string;
    readonly id: number;
    readonly name: string;
}
export function pathFrame(kind: string, id: number, name: string): PathFrameInterface {
    return { kind, id, name };
}
export function pathProperty(name: string): PathFrameInterface {
    return pathFrame('property', -1, name);
}
export function pathIndex(index: number): PathFrameInterface {
    return pathFrame('index', index, '');
}
export interface PathPredicateInterface {
    readonly present: boolean;
    readonly test: (value: PathFrameInterface, name: PathFrameInterface, index: number) => boolean;
}
export class AstPath {
    readonly arena: AstArena;
    readonly stack: PathFrameInterface[] = [];
    constructor(arena: AstArena, root: number) {
        this.arena = arena;
        this.stack.push(pathFrame('node', root, ''));
    }
    at(index: number): PathFrameInterface {
        const adjusted = index < 0 ? index + this.stack.length : index;
        return this.stack[adjusted] ?? pathFrame('missing', -1, '');
    }
    value(): PathFrameInterface {
        return this.at(-1);
    }
    node(): number {
        const value = this.value();
        return value.kind === 'node' ? value.id : -1;
    }
    siblings(): PathFrameInterface {
        const value = this.at(-3);
        return value.kind === 'array' ? value : pathFrame('missing', -1, '');
    }
    key(): string {
        const value = this.at(this.isInArray() ? -4 : -2);
        return value.kind === 'property' ? value.name : '';
    }
    index(): number {
        if(!this.isInArray()) return -1;
        const value = this.at(-2);
        return value.kind === 'index' ? value.id : -1;
    }
    name(): PathFrameInterface {
        return this.stack.length > 1 ? this.at(-2) : pathFrame('missing', -1, '');
    }
    arrayLength(value: PathFrameInterface): number {
        if(value.kind !== 'array' || value.name === 'comments') return 0;
        return this.arena.node(value.id).children.length;
    }
    arrayElement(value: PathFrameInterface, index: number): PathFrameInterface {
        if(value.kind !== 'array' || value.name === 'comments' || index < 0) return pathFrame('missing', -1, '');
        const child = this.arena.node(value.id).children[index];
        return child === undefined ? pathFrame('missing', -1, '') : pathFrame('node', child, '');
    }
    next(): number {
        const frame = this.arrayElement(this.siblings(), this.index() + 1);
        return frame.kind === 'node' ? frame.id : -1;
    }
    previous(): number {
        const frame = this.arrayElement(this.siblings(), this.index() - 1);
        return frame.kind === 'node' ? frame.id : -1;
    }
    isInArray(): boolean {
        return this.siblings().kind === 'array';
    }
    isFirst(): boolean {
        return this.isInArray() && this.index() === 0;
    }
    isLast(): boolean {
        return this.isInArray() && this.index() === this.arrayLength(this.siblings()) - 1;
    }
    isRoot(): boolean {
        return this.stack.length === 1;
    }
    root(): number {
        const frame = this.at(0);
        return frame.kind === 'node' ? frame.id : -1;
    }
    nodeStackIndex(count: number): number {
        let remaining = count;
        for(let index = this.stack.length - 1; index >= 0; index -= 2) {
            if(this.at(index).kind !== 'array') {
                remaining--;
                if(remaining < 0) return index;
            }
        }
        return -1;
    }
    getNode(count: number): number {
        const index = this.nodeStackIndex(count);
        if(index < 0) return -1;
        const frame = this.at(index);
        return frame.kind === 'node' ? frame.id : -1;
    }
    parent(): number {
        return this.getNode(1);
    }
    grandparent(): number {
        return this.getNode(2);
    }
    getParentNode(count: number): number {
        return this.getNode(count + 1);
    }
    ancestors(): number[] {
        const output: number[] = [];
        for(let index = this.stack.length - 3; index >= 0; index -= 2) {
            const frame = this.at(index);
            if(frame.kind === 'node') output.push(frame.id);
        }
        return output;
    }
    findAncestor(predicate: (id: number) => boolean): number {
        for(let index = this.stack.length - 3; index >= 0; index -= 2) {
            const frame = this.at(index);
            if(frame.kind === 'node' && predicate(frame.id)) return frame.id;
        }
        return -1;
    }
    hasAncestor(predicate: (id: number) => boolean): boolean {
        return this.findAncestor(predicate) >= 0;
    }
    field(value: PathFrameInterface, name: PathFrameInterface): PathFrameInterface {
        if(name.kind === 'index') return this.arrayElement(value, name.id);
        if(value.kind === 'node') {
            if(name.name === 'children' && this.arena.node(value.id).parent)
                return pathFrame('array', value.id, name.name);
            if(name.name === 'comments') return pathFrame('array', value.id, name.name);
        }
        return pathFrame('missing', -1, '');
    }
    enter(names: readonly PathFrameInterface[]): void {
        let value = this.value();
        for(const name of names) {
            value = this.field(value, name);
            this.stack.push(name);
            this.stack.push(value);
        }
    }
    checkpoint(): number {
        return this.stack.length;
    }
    truncate(length: number): void {
        while(this.stack.length > length) this.stack.pop();
    }
    call(callback: (path: AstPath) => string, names: readonly PathFrameInterface[]): string {
        const length = this.checkpoint();
        this.enter(names);
        const result = callback(this);
        this.truncate(length);
        return result;
    }
    callNumber(callback: (path: AstPath) => number, names: readonly PathFrameInterface[]): number {
        const length = this.checkpoint();
        this.enter(names);
        const result = callback(this);
        this.truncate(length);
        return result;
    }
    callParentNumber(callback: (path: AstPath) => number, count: number): number {
        const index = this.nodeStackIndex(count + 1);
        const saved = this.stack.slice(index + 1);
        this.truncate(index + 1);
        const result = callback(this);
        for(const frame of saved) this.stack.push(frame);
        return result;
    }
    callParentBoolean(callback: (path: AstPath) => boolean, count: number): boolean {
        const index = this.nodeStackIndex(count + 1);
        const saved = this.stack.slice(index + 1);
        this.truncate(index + 1);
        const result = callback(this);
        for(const frame of saved) this.stack.push(frame);
        return result;
    }
    callParent(callback: (path: AstPath) => string, count: number): string {
        const index = this.nodeStackIndex(count + 1);
        const saved = this.stack.slice(index + 1);
        this.truncate(index + 1);
        const result = callback(this);
        for(const frame of saved) this.stack.push(frame);
        return result;
    }
    each(
        callback: (path: AstPath, index: number, array: PathFrameInterface) => void,
        names: readonly PathFrameInterface[],
    ): void {
        const length = this.checkpoint();
        this.enter(names);
        const value = this.value();
        for(let index = 0; index < this.arrayLength(value); index++) {
            const before = this.checkpoint();
            this.stack.push(pathFrame('index', index, ''));
            this.stack.push(this.arrayElement(value, index));
            callback(this, index, value);
            this.truncate(before);
        }
        this.truncate(length);
    }
    map(
        callback: (path: AstPath, index: number, array: PathFrameInterface) => number,
        names: readonly PathFrameInterface[],
    ): number[] {
        const results: number[] = [];
        this.each(function(path: AstPath, index: number, array: PathFrameInterface): void {
            results.push(callback(path, index, array));
        }, names);
        return results;
    }
    match(predicates: readonly PathPredicateInterface[]): boolean {
        let pointer = this.stack.length - 1;
        let name = pathFrame('missing', -1, '');
        let node = this.at(pointer);
        pointer--;
        for(const predicate of predicates) {
            if(node.kind === 'missing') return false;
            let index = -1;
            if(name.kind === 'index') {
                index = name.id;
                name = pointer < 0 ? pathFrame('missing', -1, '') : this.at(pointer);
                pointer--;
                node = pointer < 0 ? pathFrame('missing', -1, '') : this.at(pointer);
                pointer--;
            }
            if(predicate.present && !predicate.test(node, name, index)) return false;
            name = pointer < 0 ? pathFrame('missing', -1, '') : this.at(pointer);
            pointer--;
            node = pointer < 0 ? pathFrame('missing', -1, '') : this.at(pointer);
            pointer--;
        }
        return true;
    }
}
