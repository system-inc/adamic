// A keyed getter returns a union-over-number field read through a checked view: natively the read boxes the number,
// and the descriptor helper hands that owned result to the read's consumer, which frees it.
interface Root { readonly kind: 'Holder'; }
interface Held extends Root { readonly amount: number | string; }
class Wrapper {
	readonly held: Held;
	constructor(base: Root) { this.held = base as Held; }
	get value(): number | string { return this.pick(); }
	pick(): number | string { return this.held.amount; }
}
function read(value: object, key: string): unknown {
	return Reflect.get(Object(value), key, value);
}
const raw = { kind: 'Holder' as const, amount: 1 };
console.log(String(read(new Wrapper(raw), 'value')));
