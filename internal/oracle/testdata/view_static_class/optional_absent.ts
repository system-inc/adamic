// #s03ndgw: an optional field the class doesn't declare is absent in Node, and reads undefined. Since #24kvb6k's Now,
// a contract with an optional field admits no class, so both backends stop; #24kvb6k re-admits it.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.count) + ' ' + String(viewed.value.extra));
