// #s03ndgw: two optional fields, neither an array's, each read absent from the array the view admitted.
interface Holder { readonly kind: 'h' | 'g'; }
interface Loose { readonly a?: number; readonly b?: string; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: Loose; }
const raw = { kind: 'h' as const, inner: [3, 4] };
const holder: Holder = raw;
const inner = (holder as Boxed).inner;
console.log(typeof inner);
console.log(String(inner.a));
console.log(inner.b === undefined ? 'no b' : inner.b);
