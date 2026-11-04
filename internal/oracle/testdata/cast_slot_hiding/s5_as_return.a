interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function g(): Positioned { return wide as Node; }
console.log(String(g().name ?? 'none'));
