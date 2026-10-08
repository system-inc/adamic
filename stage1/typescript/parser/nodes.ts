// Node links are table indexes. No parent or child link owns another node.
import { panic } from 'adamic';

export class ParseNode {
    readonly kind: string;
    readonly pos: number;
    end: number;
    readonly children: number[];
    text = '';
    raw = '';
    operator = '';
    semantic = '';
    // Optional for-header roles, kept separate from the canonical child order/tree dump.
    slots: number[] = [];
    optional = false;
    literalFlags = 0;
    list = -1;
    trailing = false;
    multiLine = false;
    constructor(kind: string, pos: number, end: number, children: number[]) {
        this.kind = kind;
        this.pos = pos;
        this.end = end;
        this.children = children;
    }
}

export function written(text: string): string {
    let result = '';
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        result +=
            code >= 32 && code <= 126 && code !== 92
                ? text.slice(index, index + 1)
                : `\\u${code.toString(16).padStart(4, '0')}`;
    }
    return result;
}

const printedDepths = '0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31'.split(' ');

export function printTree(
    nodes: readonly ParseNode[],
    root: number,
    offsets: readonly number[],
    depth = 0,
    whole = false,
): number {
    const node = nodes[root] ?? panic('missing parse node');
    const depthText = depth >= 0 && depth < printedDepths.length ? printedDepths[depth] ?? panic('missing printed depth') : `${depth}`;
    console.log(
        `${depthText} ${node.kind} ${offsets[node.pos] ?? panic('node start outside source')} ${offsets[node.end] ?? panic('node end outside source')} ${node.optional ? '32' : '0'} ${node.literalFlags === 0 ? '0' : `${node.literalFlags}`} ${node.list === -1 ? '-1' : `${node.list}`} ${node.trailing ? '1' : '0'} ${node.multiLine ? '1' : '0'}\t${node.operator}\t${written(node.text)}\t${written(node.raw)}${whole ? '\t' : ''}${whole ? node.semantic : ''}`,
    );
    let count = 1;
    for(const child of node.children) {
        count += printTree(nodes, child, offsets, depth + 1, whole);
    }
    return count;
}

export function countTree(nodes: readonly ParseNode[], root: number): number {
    const node = nodes[root] ?? panic('missing parse node');
    let count = 1;
    for(const child of node.children) {
        count += countTree(nodes, child);
    }
    return count;
}
