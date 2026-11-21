// Compiler's probe on #s03ndgw: a class held at a plain object-typed view field whose static the view names is missing.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: string; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass { static readonly tag = 'I'; }
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
const inner = viewed.value;
console.log(String(inner.count));
console.log(String(inner.count + 1));
