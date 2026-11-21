// Compiler's probe on #s03ndgw: the view's static is only inherited.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: string; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base { static readonly tag = 'B'; static readonly count = 7; }
class Sub extends Base {}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
const inner = viewed.value;
console.log(inner.tag);
console.log(String(inner.count));
