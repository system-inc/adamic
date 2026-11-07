interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 1; readonly text: string; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:1,text:'name'}]};
const base: Base = raw;
const view = base as SyntaxList;
const identifiers=view._children as readonly Identifier[];
const first=identifiers[0];
console.log(first === undefined ? "none" : first.text);
