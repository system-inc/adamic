// #s03ndgw, a stop: push is Array's, so an array holds it whatever the object type says, and the view's
// read keeps its stop, found array, where Node prints object.
interface Holder { readonly kind: 'h' | 'g'; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: { readonly push?: number }; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
