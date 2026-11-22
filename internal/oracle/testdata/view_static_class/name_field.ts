// #s03ndgw, a stop: name is a property every constructor has as its own (Function's). Klass declares no
// static named name, so Node reads its builtin name, which no static holds, and the selection refuses it.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly name: string; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : read.name);
