interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function f(p: Positioned): string { return String(p.name ?? 'none'); }
const other: Positioned = { kind: 2, name: 'n' };
const flag = other.kind > 1;
console.log(f(flag ? wide as Node : other));
