interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly say: () => string; }
interface Says { readonly say: () => string; }
const v = { kind: 9, pos: 1, say: 5 };
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const s: Says = l; console.log(s.say()); } }
show(v);
