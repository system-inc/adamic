interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
interface TextGetter { get text(): string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; const g: TextGetter = l; console.log(g.text); } }
show(v);
