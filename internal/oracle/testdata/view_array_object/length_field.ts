// #s03ndgw, a stop: an array's length is its own, so the view's read keeps its stop, found array, where
// Node prints object.
interface Holder { readonly kind: 'h' | 'g'; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: { readonly length?: number }; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
