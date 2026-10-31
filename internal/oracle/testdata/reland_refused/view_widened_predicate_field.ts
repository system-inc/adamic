interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
function isLit(n: Node): n is Lit { return n.kind === 9; }
interface Box { readonly item: HasText }
function show(n: Node): void { if (isLit(n)) { const b: Box = { item: n }; console.log(b.item.text); } }
show(v);
