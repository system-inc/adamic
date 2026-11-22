// #24kvb6k, a Now's second road: a static getter named like an optional field. Node runs the getter and reads
// its value; it must never read as absent. A contract with an optional field admits no class.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static get extra(): number {
		return 4;
	}
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
console.log(String(viewed.value.extra));
