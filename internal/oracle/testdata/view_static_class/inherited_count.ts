// #s03ndgw, a stop: Sub owns its tag but only inherits count from Base, so the view's read stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface View extends Root { readonly value: number | Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : String(read.count));
