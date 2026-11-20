enum Kind { A, B, C }
interface Base { readonly kind: Kind }
interface XA extends Base { readonly kind: Kind.A; readonly a: string }
interface ZA extends Base { readonly kind: Kind.B; readonly a: string }
interface XB extends Base { readonly kind: Kind.C; readonly b: number }
interface YA extends Base { readonly kind: Kind.A; readonly a: number }
function guard(run: () => string): string { try { return run(); } catch (error) { return 'threw'; } }
type Inner = YA | XB;
type Outer = Inner | XA | ZA;
interface WA extends Base { readonly kind: Kind.A; readonly a: string }
function show(s: XA | WA | ZA): string { switch (s.kind) { case Kind.A: case Kind.B: return guard(() => String(s.a).toUpperCase() + typeof s.a); default: return 'other'; } }
const w: WA = { kind: Kind.A, a: 'q' };
console.log(show(w));
