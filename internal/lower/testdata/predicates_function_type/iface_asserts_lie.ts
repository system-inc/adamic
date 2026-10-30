interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface Asserter { assertIdentifier(node: Node): asserts node is Identifier; }
const asserter: Asserter = { assertIdentifier: (node: Node): void => {} };
function show(a: Asserter, node: Node): void { a.assertIdentifier(node); console.log(node.escapedText); }
show(asserter, { kind: 11 } as Node);
