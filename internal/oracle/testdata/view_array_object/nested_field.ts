// #s03ndgw: a field whose own contract is an object, { c: number }, is absent from the array the view
// admitted like any other: it reads undefined or stops, and is never read raw as that object.
interface Holder { readonly kind: 'h' | 'g'; }
interface Nesting { readonly b?: { readonly c: number }; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: Nesting; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
const b = inner.b;
console.log(b === undefined ? 'absent' : String(b.c));
