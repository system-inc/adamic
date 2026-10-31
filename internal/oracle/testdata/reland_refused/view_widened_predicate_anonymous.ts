interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
function isLit(n: Node): n is Lit { return n.kind === 9; }
function show(n: Node): void { if (isLit(n)) { const r: { readonly kind: 9; readonly pos: number; readonly text: string } = n; console.log(r.text); } }
show(v);
