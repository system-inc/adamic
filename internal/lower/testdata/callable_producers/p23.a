// A good callable union view whose producer returns a narrower literal type.
type Target = ((value: number) => number) | ((value: string) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
const produce = (value: number): 1 => 1;
const raw = { kind: 'Receiver' as const, value: produce };
const base: { readonly kind: 'Receiver' } = raw;
const view = base as Receiver;
const callable = view.value;
console.log(`${callable !== undefined}`);
