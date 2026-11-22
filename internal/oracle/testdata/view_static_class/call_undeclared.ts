// #s03ndgw, a stop: Inner requires call and Klass declares only tag, so Node reads Function's call, a method
// no static holds, and the selection refuses Klass.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly call: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : typeof read.call);
