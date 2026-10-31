interface Node { readonly kind: number; readonly pos: number; }
interface Inner { readonly text: string; }
interface Outer extends Node { readonly kind: 9; readonly inner: Inner; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, inner: { text: 5 } };
function show(n: Node): void { if (n.kind === 9) { const o = n as Outer; const h: HasText = o.inner; console.log(h.text); } }
show(v);
