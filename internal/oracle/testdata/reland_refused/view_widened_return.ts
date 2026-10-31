interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
function view(n: Node): HasText | undefined { if (n.kind === 9) return n as Lit; return undefined; }
function show(n: Node): void { const h = view(n); console.log(h === undefined ? "none" : h.text); }
show(v);
