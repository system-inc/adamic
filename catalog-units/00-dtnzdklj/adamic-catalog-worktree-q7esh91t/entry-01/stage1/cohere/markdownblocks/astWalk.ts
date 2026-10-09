// Preorder mapAst with replacement-aware parent paths, without recursive closures.
import { panic } from 'adamic';
import type { AstArena } from './astArena.ts';
export class AstWalk {
    readonly arena: AstArena;
    readonly ids: number[] = [];
    readonly indices: number[] = [];
    readonly depths: number[] = [];
    readonly ancestors: number[] = [];
    root: number;
    id = -1;
    index = -1;
    depth = 0;
    parent = -1;
    constructor(arena: AstArena, root: number) {
        this.arena = arena;
        this.root = root;
        this.ids.push(root);
        this.indices.push(-1);
        this.depths.push(0);
    }
    next(): boolean {
        if(this.ids.length === 0) return false;
        this.id = this.ids.pop() ?? panic('walk node');
        this.index = this.indices.pop() ?? panic('walk sibling');
        this.depth = this.depths.pop() ?? panic('walk depth');
        while(this.ancestors.length > this.depth) this.ancestors.pop();
        this.parent = this.depth === 0 ? -1 : (this.ancestors[this.depth - 1] ?? panic('walk parent'));
        return true;
    }
    descend(replacement: number): void {
        if(this.parent < 0) this.root = replacement;
        else this.arena.node(this.parent).children[this.index] = replacement;
        this.ancestors.push(replacement);
        const children = this.arena.node(replacement).children;
        for(let index = children.length - 1; index >= 0; index--) {
            this.ids.push(children[index] ?? panic('walk child'));
            this.indices.push(index);
            this.depths.push(this.depth + 1);
        }
    }
}
