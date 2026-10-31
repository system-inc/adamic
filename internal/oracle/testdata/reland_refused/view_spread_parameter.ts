interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
function copy(l: Lit) { return { ...l }; }
function show(n: Node): void { if (n.kind === 9) { const c = copy(n as Lit); console.log(c.text); } }
const v = { kind: 9, pos: 1, text: 5 }; show(v);
