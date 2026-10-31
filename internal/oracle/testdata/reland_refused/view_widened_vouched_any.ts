interface Loose { readonly kind: number; readonly pos: number; readonly text: any; }
interface LitL extends Loose { readonly kind: 9; readonly text: string; }
interface HasText { readonly text: string; }
function show(n: Loose): void { if (n.kind === 9) { const l = n as LitL; const h: HasText = l; console.log(h.text); } }
const w = { kind: 9, pos: 1, text: 5 }; show(w);
