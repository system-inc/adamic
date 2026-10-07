/* cohere-disable max-classes-per-file -- keep the indexed getter reproducer in one standalone program */
// An indexed object getter makes the compiler's ownership traffic observable in perf.
import { panic, programArguments } from 'adamic';
class Entry {
    readonly kind: string;
    constructor(kind: string) {
        this.kind = kind;
    }
}
class IndexedNodes {
    readonly nodes: readonly Entry[];
    constructor(nodes: readonly Entry[]) {
        this.nodes = nodes;
    }
    node(index: number): Entry {
        return this.nodes[index] ?? panic('missing node');
    }
}
function matches(nodes: IndexedNodes, index: number): boolean {
    return nodes.node(index).kind === 'Identifier';
}
const args = programArguments();
const nodes = new IndexedNodes([new Entry(args[0] ?? panic('pass Identifier'))]);
const iterations = Number.parseInt(args[1] ?? '3000000', 10);
let findings = 0;
for(let index = 0; index < iterations; index++) {
    if(matches(nodes, 0)) findings++;
}
console.log(`${findings}`);
