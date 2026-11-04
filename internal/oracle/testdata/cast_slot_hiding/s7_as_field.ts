interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
class Holder { held: Positioned = wide as Node; }
const holder = new Holder();
console.log(String(holder.held.name ?? 'none'));
