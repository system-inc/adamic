// A keyed getter calls a method that reads a union-over-number field and compares it: the read's box, if any, is a temporary the getter frees.
interface Held { readonly kind: 'Holder'; readonly amount: number | string; }

class Wrapper {
	readonly held: Held;
	constructor(base: Held) { this.held = base; }
	get isOne(): boolean { return this.check(); }
	check(): boolean { return this.held.amount === 1; }
}
function read(value: object, key: string): unknown {
	return Reflect.get(Object(value), key, value);
}
const raw = { kind: 'Holder' as const, amount: 1 };
console.log(String(read(new Wrapper(raw), 'isOne')));
