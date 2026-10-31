interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const h: HasText = l; const c = { ...h }; console.log(c.text); } }
const v = { kind: 9, pos: 1, text: 5 }; show(v);
