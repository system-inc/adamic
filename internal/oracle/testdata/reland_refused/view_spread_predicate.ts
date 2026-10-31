interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
function isLit(n: Node): n is Lit { return n.kind === 9; }
function show(n: Node): void { if (isLit(n)) { const c = { ...n }; console.log(c.text); } }
const v = { kind: 9, pos: 1, text: 5 }; show(v);
