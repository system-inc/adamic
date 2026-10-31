interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
interface Box { readonly item: HasText }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const b: Box = { item: l }; console.log(b.item.text); } }
show(v);
