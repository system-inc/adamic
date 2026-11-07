// A port of cohere's internal/format/css/values/nodes.go to Adamic 0.1: postcss-values-parser 2.0.1,
// lib/node.js, lib/container.js and the node classes (root.js, value.js, atword.js, colon.js, comma.js,
// comment.js, function.js, number.js, operator.js, paren.js, string.js, unicode-range.js, word.js).
//
// The Go builds an *estree.Node, a bag of the own fields each class leaves on the JavaScript object.
// The port has one class with every field any of them has, and a shape saying which constructor made
// the node, which decides which of the fields it has (keysOf). That is the library's own rule: a node's
// fields are its class's. The Go side of the test checks every Go node has exactly the keys its shape
// gives, so the class can't be leaving one out.
//
// The shape carries presence because stage 0 doesn't lower a boolean | undefined field yet (gap 2 in
// GAPS.md), which is how the flags (inline, isHex, isColor, quoted) present on some nodes and absent on
// others would otherwise be written. The other optional fields are plain values the shape ignores
// where it leaves them out.
//
// A container's children are not a field of its node while the parse runs. The parser appends to them
// long after it made the node (a func's arguments, a value's words), and a mutable list of nodes that
// a node keeps could be made to hold that node itself, a cycle stage 0 refuses (adamic/cycle-capable;
// mediaquery's GAPS.md, gaps 4 and 5). So a container has an id, the parser keeps each container's
// children by it, and when the parse is done it builds the tree the library returns, a ValueTree, whose
// lists are readonly.
//
// Not ported, as in the Go: parent, and everything functional (toString, clone, walk and the rest),
// which the parser never calls.

// Which constructor made a node. 'leaf' is colon, comma, operator, unicode-range, and the word parenOpen
// builds for a url's argument; 'word' is the word splitWord builds, with isHex and isColor.
export type Shape = 'root' | 'value' | 'leaf' | 'word' | 'paren' | 'comment' | 'atword' | 'number' | 'func' | 'string';

// nodes.go: position, a source start or end: { line, column }. column is in UTF-16 units, the library's
// own, and is NaN where upstream computes it from an undefined variable.
export class Position {
    readonly line: number;
    readonly column: number;

    constructor(line: number, column: number) {
        this.line = line;
        this.column = column;
    }
}

// nodes.go: source, { start, end }.
export class Source {
    readonly start: Position;
    readonly end: Position;

    constructor(start: Position, end: Position) {
        this.start = start;
        this.end = end;
    }
}

// nodes.go: newRaws, Node's `this.raws = { before: "", after: "" }`. quote is the key the parser's
// string adds; undefined where it doesn't.
export class Raws {
    before = '';
    after = '';
    quote: string | undefined = undefined;
}

export class ValueNode {
    readonly type: string;
    readonly shape: Shape;
    readonly raws = new Raws();
    readonly value: string;
    readonly source: Source | undefined;
    readonly sourceIndex: number;
    // The container's index in the parser's children, or -1 for a node that is not a container.
    readonly id: number;
    unbalanced = 0;
    readonly unit: string;
    readonly flag: boolean;
    isHex = false;
    isColor = false;

    // flag is comment's inline and string's quoted, the one boolean each of them has.
    constructor(
        type: string,
        shape: Shape,
        value: string,
        source: Source | undefined,
        sourceIndex: number,
        unit: string,
        flag: boolean,
        id: number,
    ) {
        this.type = type;
        this.shape = shape;
        this.value = value;
        this.source = source;
        this.sourceIndex = sourceIndex;
        this.unit = unit;
        this.flag = flag;
        this.id = id;
    }
}

// A node with its children: the tree a parse returns.
export class ValueTree {
    readonly node: ValueNode;
    readonly nodes: readonly ValueTree[];

    constructor(node: ValueNode, nodes: readonly ValueTree[]) {
        this.node = node;
        this.nodes = nodes;
    }
}

// keysOf is the node's own fields but type, parent and nodes, sorted, as the library's class leaves
// them: what the test's dump prints.
export function keysOf(shape: Shape): readonly string[] {
    switch(shape) {
        case 'root':
            return ['raws'];
        case 'value':
            return ['raws', 'unbalanced'];
        case 'leaf':
            return ['raws', 'source', 'sourceIndex', 'value'];
        case 'word':
            return ['isColor', 'isHex', 'raws', 'source', 'sourceIndex', 'value'];
        case 'paren':
            return ['parenType', 'raws', 'source', 'sourceIndex', 'value'];
        case 'comment':
            return ['inline', 'raws', 'source', 'sourceIndex', 'value'];
        case 'atword':
            return ['raws', 'source', 'sourceIndex', 'value'];
        case 'number':
            return ['raws', 'source', 'sourceIndex', 'unit', 'value'];
        case 'func':
            return ['raws', 'source', 'sourceIndex', 'unbalanced', 'value'];
        case 'string':
            return ['quoted', 'raws', 'source', 'sourceIndex', 'value'];
    }
}

// hasNodes is whether the shape is a Container's.
export function hasNodes(shape: Shape): boolean {
    return shape === 'root' || shape === 'value' || shape === 'atword' || shape === 'func';
}

// nodes.go: newRoot, Container with type 'root'.
export function newRoot(id: number): ValueNode {
    return new ValueNode('root', 'root', '', undefined, 0, '', false, id);
}

// nodes.go: newValue, Container with type 'value' and unbalanced 0.
export function newValue(id: number): ValueNode {
    return new ValueNode('value', 'value', '', undefined, 0, '', false, id);
}

// nodes.go: leaf, the Node subclasses that take { value, source, sourceIndex }: colon, comma, operator,
// unicode-range and word.
export function leaf(type: string, value: string, source: Source, sourceIndex: number): ValueNode {
    return new ValueNode(type, 'leaf', value, source, sourceIndex, '', false, -1);
}

// The word splitWord builds, which sets isHex and isColor on it.
export function newWord(
    value: string,
    source: Source,
    sourceIndex: number,
    isHex: boolean,
    isColor: boolean,
): ValueNode {
    const node = new ValueNode('word', 'word', value, source, sourceIndex, '', false, -1);
    node.isHex = isHex;
    node.isColor = isColor;
    return node;
}

// nodes.go: newParen, Node with type 'paren' and parenType "".
export function newParen(value: string, source: Source, sourceIndex: number): ValueNode {
    return new ValueNode('paren', 'paren', value, source, sourceIndex, '', false, -1);
}

// nodes.go: newComment, Node with type 'comment' and inline from its options.
export function newComment(value: string, inline: boolean, source: Source, sourceIndex: number): ValueNode {
    return new ValueNode('comment', 'comment', value, source, sourceIndex, '', inline, -1);
}

// nodes.go: newAtWord, Container with type 'atword'.
export function newAtWord(value: string, source: Source, sourceIndex: number, id: number): ValueNode {
    return new ValueNode('atword', 'atword', value, source, sourceIndex, '', false, id);
}

// nodes.go: newNumber, Node with type 'number' and unit from its options, or "".
export function newNumber(value: string, source: Source, sourceIndex: number, unit: string): ValueNode {
    return new ValueNode('number', 'number', value, source, sourceIndex, unit, false, -1);
}

// nodes.go: newFunc, Container with type 'func', unbalanced starting at -1 so the parser knows no parens
// have been added yet.
export function newFunc(value: string, source: Source, sourceIndex: number, id: number): ValueNode {
    const node = new ValueNode('func', 'func', value, source, sourceIndex, '', false, id);
    node.unbalanced = -1;
    return node;
}

// nodes.go: newString, Node with type 'string' and quoted from its options.
export function newString(value: string, source: Source, sourceIndex: number, quoted: boolean): ValueNode {
    return new ValueNode('string', 'string', value, source, sourceIndex, '', quoted, -1);
}
