interface Node { readonly kind: number; }
interface Positioned { readonly kind: number; readonly name?: string; }
const wide = { kind: 1, name: 5 };
function f(p: Positioned): string { return String(p.name ?? 'none'); }
const nodes: (Node | undefined)[] = [wide];
const maybe: Node | undefined = nodes[0];
console.log(f(maybe!));
