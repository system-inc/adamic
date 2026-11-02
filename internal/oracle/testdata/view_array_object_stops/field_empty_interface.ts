// An array literal where a checked view says an object stops at the view's read on both backends: natively an
// array is never an object, and the JavaScript backend marks only tuples as objects (#nazvcep). Node prints on.
interface Holder { readonly kind: 'h' | 'g'; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: { readonly a?: number }; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
console.log(String(inner.a));
