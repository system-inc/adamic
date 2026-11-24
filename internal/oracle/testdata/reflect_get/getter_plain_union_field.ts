// The view union pin's control: the same getter reads the same union field plain, with no view, and agrees with Node.
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
