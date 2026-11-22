// #24kvb6k, a stop broader than Node (#4gm3k3c lifts it): Klass declares no extra, so Node reads Klass.extra as undefined, but Other,
// an unrelated class, declares a static method named extra. An optional field named like any static method the
// program declares admits no class, whichever class declares it, so both backends stop.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly extra?: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
}
class Other {
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
