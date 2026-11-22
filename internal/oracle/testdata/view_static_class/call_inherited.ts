// #s03ndgw: Inner requires call, a name every constructor has from Function's prototype. Base declares a data
// static named call and Sub inherits it, so Node reads Base's 5 before Function's, and so does the view.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly call: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag = 'I' as const;
	static readonly call = 5;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log(String(viewed.value.call));
