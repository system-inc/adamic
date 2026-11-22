// #24kvb6k, a Now's third road: an optional field the class only inherits as a data static from its base.
// Node reads the base's value through the prototype chain; it must never read as absent.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Base {
	static readonly tag: 'I' = 'I';
	static readonly extra: number = 6;
}
class Sub extends Base {
	static override readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Sub };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.extra));
