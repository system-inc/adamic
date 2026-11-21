// A write to a class constructor the view admitted, through a type with no view contract of its own (Writer):
// no backend guards it, since StaticsAdmittedReceiver reads the receiver's own type, and it writes the static as
// Node does (#s03ndgw, J-7agaskd). The guard is for writes through the view, whose contract says what it holds.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; count: number; }
interface Writer { readonly tag: 'I'; count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
const writer: Writer = viewed.value;
console.log('before');
writer.count = 4;
console.log(String(Klass.count));
