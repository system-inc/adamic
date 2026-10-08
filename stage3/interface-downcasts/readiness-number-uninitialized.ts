interface Node { readonly kind: string; }
interface NumberNode extends Node { readonly kind: 'number'; value: number; }
function createNumber(): NumberNode { return { kind: 'number', value: null! }; }
function visit(node: NumberNode): void { console.log(`${node.value}`); }
const numberNode = createNumber();
const held: Node = numberNode;
console.log(held.kind);
visit(numberNode);
