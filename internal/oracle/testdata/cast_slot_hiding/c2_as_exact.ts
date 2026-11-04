interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function f(p: Positioned): string { return String(p.name ?? 'none'); }
const exact: Positioned = { kind: 5, name: 'e' };
console.log(f(exact as Positioned));
