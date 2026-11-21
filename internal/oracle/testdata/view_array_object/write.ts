// #s03ndgw, a stop: a write through the view to its optional field, which an array the view admitted has
// no slot for, stops at lowering (view_optional_writes.go). Node gives the array an a of its own.
interface Holder { readonly kind: 'h' | 'g'; }
interface Loose { a?: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: Loose; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
const inner = (holder as Boxed).inner;
console.log('before');
inner.a = 4;
console.log(String(inner.a));
