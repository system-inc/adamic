// Index links avoid owning parent/child cycles.
import { panic } from 'adamic';
import { Node } from './nodes.ts';
import { visitorKeys } from './visitorKeys.ts';
export class Arena {
    readonly nodes: Node[] = [];
    newNode(type: string, start: number, end: number): number {
        this.nodes.push(new Node(type, start, end));
        return this.nodes.length - 1;
    }
    node(id: number): Node {
        return this.nodes[id] ?? panic('missing ESTree node');
    }
    type(id: number): string {
        return id < 0 ? '' : this.node(id).type;
    }
    children(id: number): number[] {
        const children: number[] = [];
        const keys: readonly string[] = visitorKeys.get(this.type(id)) ?? [];
        const node = this.node(id);
        for(const key of keys) {
            const value = node.get(key);
            if(value.kind === 'node') {
                children.push(value.node);
            }
            else if(value.kind === 'list') {
                for(const child of value.list) {
                    if(child >= 0) {
                        children.push(child);
                    }
                }
            }
        }
        return children;
    }
}
