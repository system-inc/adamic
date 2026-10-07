// A port of cohere's internal/format/css/mediaquery/nodes.go to Adamic 0.1: postcss-media-query-parser
// 0.2.3, dist/nodes/Node.js and dist/nodes/Container.js.
//
// The Go builds an *estree.Node, a bag of properties in the constructors' assignment order, because
// Prettier's printer reads it by name. The port has a class with exactly the fields the library's nodes
// carry, since 0.1 has no property bags (index signatures are refused; a Map can't hold a string, a
// number and a list together without a union for every read). Every node has after, before, value and
// sourceIndex, and a container has nodes too, which is the same set the Go's carry; the test's Go side
// checks that every Go node has exactly these keys, so the class can't be leaving one out.
//
// type is '' where the library leaves it undefined, as the Go's is: a node whose type the post-pass in
// parseMediaQuery could not decide (Prettier's addMissingType fills it).
//
// Not ported, as in the Go: parent, and Container's walk and each, which Prettier never calls.

export class MediaNode {
    type: string;
    after: string;
    before: string;
    readonly value: string;
    sourceIndex: number;

    // The container's children, or undefined for a node that is not a container.
    readonly nodes: readonly MediaNode[] | undefined;

    constructor(
        type: string,
        after: string,
        before: string,
        value: string,
        sourceIndex: number,
        nodes: readonly MediaNode[] | undefined,
    ) {
        this.type = type;
        this.after = after;
        this.before = before;
        this.value = value;
        this.sourceIndex = sourceIndex;
        this.nodes = nodes;
    }
}

// nodes.go: newNode, `new Node(opts)`: a very generic node. Pretty much any element of a media query.
export function newNode(after: string, before: string, type: string, value: string, sourceIndex: number): MediaNode {
    return new MediaNode(type, after, before, value, sourceIndex, undefined);
}

// nodes.go: newContainer, `new Container(opts)`: a node that contains other nodes. after, before and
// sourceIndex are undefined only for the root, which index.js builds without them, and Container derives
// them from its children.
export function newContainer(
    after: string | undefined,
    before: string | undefined,
    type: string,
    value: string,
    sourceIndex: number | undefined,
    nodes: readonly MediaNode[],
): MediaNode {
    const first = nodes[0];
    const last = nodes.at(-1);
    const derivedAfter = after ?? (last !== undefined ? last.after : '');
    const derivedBefore = before ?? (first !== undefined ? first.before : '');
    // this.before.length: before is whitespace from the input, so its length is the offset. The Go counts
    // bytes and the port UTF-16 units, as the library does.
    const derivedSourceIndex = sourceIndex ?? derivedBefore.length;
    return new MediaNode(type, derivedAfter, derivedBefore, value, derivedSourceIndex, nodes);
}
