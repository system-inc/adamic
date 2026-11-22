// #s03ndgw: Inner's optional extra is a static only Base declares. Sub inherits it, so it's there, not absent,
// and the read finds Base's 4, as Node does.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly count: number = 3;
	static readonly extra = 4;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
	static override readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log(String(viewed.value.count) + ' ' + String(viewed.value.extra));
