interface Node { readonly kind: number; readonly pos: number; }
interface Inner { readonly label: string; }
interface Lit extends Node { readonly kind: 9; readonly parent: Inner; }
interface HasParent { readonly parent: { readonly label: string }; }
const v = { kind: 9, pos: 1, parent: { label: 5 } };
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const h: HasParent = l; console.log(h.parent.label); } }
show(v);
