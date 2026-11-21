// #s03ndgw, a stop: the array a view admitted, written through a parameter's type, which no view checks.
// The view reads an array's absent fields and never writes them, and an array has no object's slots, so
// the write stops on both backends, after its value, before it writes. Node gives the array an a.
interface Holder { readonly kind: 'h' | 'g'; }
interface Loose { readonly a?: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: Loose; }
function set(target: { a?: number }): void {
	target.a = 4;
}
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
const inner = (holder as Boxed).inner;
console.log('before');
set(inner);
console.log(String(inner.a));
