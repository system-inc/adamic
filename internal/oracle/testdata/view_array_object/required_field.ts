// #s03ndgw, a stop: a required field, which no array has, keeps the union selection's stop on the union
// road, where Node prints object.
interface Holder { readonly kind: 'h' | 'g'; }
interface Demanding { readonly a: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: number | Demanding; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
console.log('before');
const inner = (holder as Boxed).inner;
console.log(typeof inner);
