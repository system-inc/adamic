interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function f(p: Positioned): string { return String(p.name ?? 'none'); }
console.log(f({ kind: 3 } as const));
console.log(f({ kind: 4, name: 'x' } as const));
