// #s03ndgw, a stop: Base declares a static named name, but every constructor has a name of its own, so Node
// reads Sub's own, Sub, never Base's B. Sub writes no static of that name, so the view's read stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly name: string; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly name: string = 'B';
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(viewed.value.name);
