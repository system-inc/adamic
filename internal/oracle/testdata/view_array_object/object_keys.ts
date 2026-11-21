// #s03ndgw, a stop: Object.keys through the view would read an array's indices, which the object the view
// says has none of, so it stops on both backends. Node prints 0.
interface Holder { readonly kind: 'h' | 'g'; }
interface Loose { readonly a?: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: Loose; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
const inner = (holder as Boxed).inner;
console.log('before');
console.log(Object.keys(inner).join(','));
