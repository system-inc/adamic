import { type Node, type Identifier } from './nodes.a';
function identifier(node: Node): Identifier { return node as Identifier; }
const malformed: Node = { kind: 'identifier', pos: 0 };
console.log(identifier(malformed).name ?? 'missing');
