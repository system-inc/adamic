interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
const positioned: Positioned = wide as Node;
console.log(String(positioned.name ?? 'none'));
