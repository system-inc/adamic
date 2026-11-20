enum Kind { A, B, C }
interface Base { readonly kind: Kind }
interface XA extends Base { readonly kind: Kind.A; readonly a: string }
interface ZA extends Base { readonly kind: Kind.B; readonly a: string }
interface XB extends Base { readonly kind: Kind.C; readonly b: number }
interface YA extends Base { readonly kind: Kind.A; readonly a: number }
function guard(run: () => string): string { try { return run(); } catch (error) { return 'threw'; } }
type Inner = YA | XB;
type Outer = Inner | XA | ZA;
function show(s: XA | YA): string { return String(s['a']) + typeof s['a']; }
const ya: YA = { kind: Kind.A, a: 3 };
console.log(show(ya));
