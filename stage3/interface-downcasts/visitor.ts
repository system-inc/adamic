import { identifier, numberNode, stringNode, binaryNode, type Node, type Identifier, type NumberNode, type StringNode, type BinaryNode } from './nodes.a';
function visit(node: Node): string {
	switch (node.kind) {
		case 'identifier': return (node as Identifier).name;
		case 'number': return `${(node as NumberNode).value}`;
		case 'string': return (node as StringNode).text;
		case 'binary': return `${(node as BinaryNode).left + (node as BinaryNode).right}`;
	}
}
const nodes: readonly Node[] = [identifier('id'.repeat(2)), numberNode(42), stringNode('text'.repeat(2)), binaryNode(2, 3)];
for (const node of nodes) { console.log(visit(node)); }
let calls = 0;
function next(): Node { calls++; return identifier(`once${calls}`); }
const held: Node = next();
console.log(`${(held as Identifier) === held}`);
console.log((next() as Identifier).name);
console.log(`calls ${calls}`);
