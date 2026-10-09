import { numberNode, type Node, type Identifier } from './nodes.a';
function identifier(node: Node): Identifier { return node as Identifier; }
console.log('casting');
console.log(identifier(numberNode(42)).name ?? 'missing');
