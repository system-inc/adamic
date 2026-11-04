enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
type N = A | B;
interface H { readonly n: N }
interface M { n: N }
function isA(n: N): n is A { return n.kind === K.A; }
const a: A = { kind: K.A, flag: F.X, a: 1 };
const b: B = { kind: K.B, b: 'x' };
const ha: H = { n: a };
const hb: H = { n: b };
class Lazy implements H { get n(): N { return a; } }
function show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }
