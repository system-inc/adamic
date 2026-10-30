interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface Asserter { assertIdentifier: (node: Node, note: string) => asserts node is Identifier; }
const asserter: Asserter = { assertIdentifier: (node: Node, note: string): void => { console.log("old " + note); } };
function swap(a: Asserter): string { a.assertIdentifier = (n: Node, note: string): void => { console.log("new " + note); }; return "swapped"; }
function show(a: Asserter, node: Node): void { a.assertIdentifier(node, swap(a)); console.log(node.escapedText); }
show(asserter, { kind: 80, escapedText: "x" } as Node);
