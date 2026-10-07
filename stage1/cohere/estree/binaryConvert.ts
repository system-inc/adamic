// A left binary spine needs no language or JavaScript call-stack recursion.
import { panic } from 'adamic';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Arena } from './arena.ts';
import { childValue, stringValue } from './values.ts';
class BinarySource {
    readonly scanner: Scanner;
    readonly nodes: readonly ParseNode[];
    readonly text: string;
    readonly offsets: readonly number[];
    constructor(nodes: readonly ParseNode[], text: string, offsets: readonly number[]) {
        this.nodes = nodes;
        this.text = text;
        this.offsets = offsets;
        this.scanner = new Scanner(text);
    }
    child(id: number, index: number): number {
        const node = this.nodes[id];
        return node === undefined ? -1 : (node.children[index] ?? -1);
    }
    start(id: number): number {
        this.scanner.pos = this.nodes[id]?.pos ?? 0;
        this.scanner.scan();
        return this.scanner.start;
    }
    raw(id: number): string {
        return this.text.slice(this.start(id), this.nodes[id]?.end ?? 0);
    }
    create(arena: Arena, id: number, type: string): number {
        return arena.newNode(
            type,
            this.offsets[this.start(id)] ?? panic('binary start outside source'),
            this.offsets[this.nodes[id]?.end ?? -1] ?? panic('binary end outside source'),
        );
    }
}
export function ordinaryBinary(
    arena: Arena,
    nodes: readonly ParseNode[],
    text: string,
    offsets: readonly number[],
    root: number,
    convert: (id: number, parent: number) => number,
): number {
    const source = new BinarySource(nodes, text, offsets);
    const spine: number[] = [];
    let current = root;
    while(nodes[current]?.kind === 'BinaryExpression') {
        const operator = source.raw(source.child(current, 1));
        if(operator === ',' || (operator.endsWith('=') && !['==', '===', '!=', '!==', '<=', '>='].includes(operator))) {
            break;
        }
        spine.push(current);
        current = source.child(current, 0);
    }
    let left = convert(current, spine[spine.length - 1] ?? -1);
    while(spine.length > 0) {
        const id = spine.pop() ?? panic('missing binary spine');
        const operator = source.raw(source.child(id, 1));
        const result = source.create(
            arena,
            id,
            ['&&', '||', '??'].includes(operator) ? 'LogicalExpression' : 'BinaryExpression',
        );
        const node = arena.node(result);
        node.set('operator', stringValue(operator));
        node.set('left', childValue(left));
        node.set('right', childValue(convert(source.child(id, 2), -1)));
        left = result;
    }
    return left;
}
