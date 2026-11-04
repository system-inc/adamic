interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
const holder: { readonly held: Positioned } = { held: wide as Node };
console.log(String(holder.held.name ?? 'none'));
