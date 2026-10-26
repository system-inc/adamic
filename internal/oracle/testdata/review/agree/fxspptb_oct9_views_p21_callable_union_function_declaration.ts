// A good callable union view whose producer is a function declaration.
type Target = ((value: number) => number) | ((value: string) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
function increment(value: number): number {
	return value + 1;
}
const raw = { kind: 'Receiver' as const, value: increment };
const base: { readonly kind: 'Receiver' } = raw;
const view = base as Receiver;
const callable = view.value;
console.log(`${callable !== undefined}`);
