interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const h: HasText = l; for (const k in h) console.log(h[k as keyof HasText]); } }
show(v);
