// #s03ndgw: Sub owns its tag and inherits count from Base, held at a field typed Inner itself, so the read
// walks Sub's statics to Base's, as Node walks Sub's prototype, and reads 3.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.count));
