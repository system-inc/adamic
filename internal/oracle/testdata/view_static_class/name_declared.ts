// #s03ndgw: name is every constructor's own (Function's), but Klass declares its own static named name, which
// Node defines over the builtin, so the view reads Named, as Node does.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly name: string; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly name = 'Named';
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log(viewed.value.name);
