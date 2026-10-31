interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
const v = { kind: 9, pos: 1, text: 5 };
class Printer { constructor(private readonly h: HasText) {} run(): string { return this.h.text; } }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(new Printer(l).run()); } }
show(v);
