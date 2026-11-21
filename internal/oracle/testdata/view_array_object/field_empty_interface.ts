// #s03ndgw: an array literal where a checked view says an object whose every field is optional and no
// array's is that object, as Node holds it: typeof says object, and its field reads undefined. Before,
// both backends stopped at the view's read, found array (#nazvcep's stop, lifted here).
interface Holder { readonly kind: 'h' | 'g'; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: { readonly a?: number }; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
console.log(String(inner.a));
