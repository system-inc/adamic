// A keyed getter calls a method that reads a union-over-number field and tests its typeof: the read's box is a temporary the getter frees.
interface Held { readonly kind: 'Holder'; readonly amount: number | string; }

class Wrapper {
	readonly held: Held;
	constructor(base: Held) { this.held = base; }
	get isOne(): boolean { return this.check(); }
	check(): boolean { return typeof this.held.amount === "number"; }
}
function read(value: object, key: string): unknown {
	return Reflect.get(Object(value), key, value);
}
const raw: Held = { kind: 'Holder', amount: 1 };
console.log(String(read(new Wrapper(raw), 'isOne')));
