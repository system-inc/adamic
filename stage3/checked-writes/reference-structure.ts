type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface Parent { readonly name: string }
interface Node { readonly parent: Parent }
function replace<T extends Node>(node: T, parent: Parent): void { (node as Mutable<T>).parent = parent; }
const node: { readonly parent: { readonly name: 'left' } } = { parent: { name: 'left' } };
const parent: Parent = { name: 'right' };
replace(node, parent);
console.log(node.parent.name);
