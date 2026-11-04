interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function f(p: Positioned): number { return p.name === undefined ? -1 : p.name.length; }
console.log(String(f(wide as Node)));
