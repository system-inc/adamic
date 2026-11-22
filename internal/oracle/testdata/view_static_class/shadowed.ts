// #s03ndgw: Sub writes its own count over Base's, so the read finds Sub's 7 first, as Node does, never Base's 3.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly count: number = 3;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
	static override readonly count = 7;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log(String(viewed.value.count));
