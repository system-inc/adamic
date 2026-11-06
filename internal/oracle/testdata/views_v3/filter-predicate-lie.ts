interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 1; readonly text: string; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 1; }
const raw={tag:1 as const,_children:[7]};
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
const identifiers=view._children.filter(isIdentifier);
const first=identifiers[0];
console.log('after');
