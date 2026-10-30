interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface Asserter { assertIdentifier: (node: Node, other: Node) => asserts node is Identifier; }
const asserter: Asserter = { assertIdentifier: (node: Node, other: Node): void => { console.log("old"); } };
class Holder {
    readonly asserter: Asserter;
    readonly held: Node;
    constructor(asserter: Asserter, held: Node) { this.asserter = asserter; this.held = held; }
    get node(): Node { this.asserter.assertIdentifier = (n: Node, o: Node): void => { console.log("new"); }; return this.held; }
}
interface HasNode { readonly node: Node; }
function show(a: Asserter, node: Node, holder: HasNode): void { a.assertIdentifier(node, holder.node); console.log(node.escapedText); }
const identifier = { kind: 80, escapedText: "x" } as Node;
show(asserter, identifier, new Holder(asserter, identifier));
