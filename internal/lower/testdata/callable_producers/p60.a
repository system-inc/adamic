type Target = ((value: number) => number) | ((value: string) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
function double(value: number): number { return value * 2; }
const raw = { kind: 'Receiver' as const, value: double };
const base: { readonly kind: 'Receiver' } = raw;
const view = base as Receiver;
const callable = view.value;
console.log(`${callable !== undefined}`);
