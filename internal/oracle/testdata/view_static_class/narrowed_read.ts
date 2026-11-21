// #s03ndgw, a stop: the union shape, narrowed by typeof before .count reads. The check lowering puts on the
// narrowed local (localRead, internal/lower/locals.go) asks typeof === 'object' of what the read put back,
// and a class answers 'function', so it stops there at run time. It awaits a follow-up that teaches that
// check a class constructor.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
const read = viewed.value;
if (typeof read !== 'number') {
	console.log(String(read.count));
}
