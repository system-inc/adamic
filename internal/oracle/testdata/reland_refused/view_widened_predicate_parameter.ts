interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
function isLit(n: Node): n is Lit { return n.kind === 9; }
function textOf(x: HasText): string { return x.text; }
function show(n: Node): void { if (isLit(n)) { console.log(textOf(n)); } }
show(v);
