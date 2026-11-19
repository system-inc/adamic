interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
class Shout implements HasText { get text(): string { return "loud"; } }
function textOf(h: HasText): string { return h.text; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(textOf(l)); } }
show(v);
console.log(textOf(new Shout()));
