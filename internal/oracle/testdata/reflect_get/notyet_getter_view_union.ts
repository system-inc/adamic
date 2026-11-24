// A keyed getter calls a method that reads a union-over-number field through a checked view: natively the view read boxes it.
interface Root { readonly kind: 'Holder'; }
interface Held extends Root { readonly amount: number | string; }
class Wrapper {
	readonly held: Held;
	constructor(base: Root) { this.held = base as Held; }
	get isOne(): boolean { return this.check(); }
	check(): boolean { return this.held.amount === 1; }
}
function read(value: object, key: string): unknown {
	return Reflect.get(Object(value), key, value);
}
const raw = { kind: 'Holder' as const, amount: 1 };
console.log(String(read(new Wrapper(raw), 'isOne')));
