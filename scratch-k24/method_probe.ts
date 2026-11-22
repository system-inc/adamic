// #24kvb6k's question: a static method named like an optional field of the view's contract.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static extra(): number {
		return 9;
	}
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
const extra = viewed.value.extra;
console.log(extra === undefined ? 'undefined' : typeof extra);
