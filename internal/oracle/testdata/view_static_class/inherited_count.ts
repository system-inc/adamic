// #s03ndgw, a stop past the selection: Sub owns its tag and inherits count from Base, which the union's
// selection admits, as Node reads Sub.count through Base. The read of count is through a local typeof
// narrowed, and the check lowering puts there asks typeof === 'object' of a class, so it stops as
// narrowed_read does, until #775qbxz. inherited_field reads the 3 at a field typed Inner itself.
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
