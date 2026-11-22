// #s03ndgw, a stop: Sub owns its tag, and count is a static getter Base declares, not a data static, so the
// view's read stops, as static_getter's does for a class's own getter.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface View extends Root { readonly value: number | Inner; }
class Base {
	static readonly tag = 'I' as const;
	static get count(): number {
		return 3;
	}
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
