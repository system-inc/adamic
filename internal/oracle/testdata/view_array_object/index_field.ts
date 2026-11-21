// #s03ndgw, a stop: '0' is an index an array holds as its own, so on the union road the selection admits
// no member for the array, where Node prints object.
interface Holder { readonly kind: 'h' | 'g'; }
interface Indexed { readonly '0'?: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: number | Indexed; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
