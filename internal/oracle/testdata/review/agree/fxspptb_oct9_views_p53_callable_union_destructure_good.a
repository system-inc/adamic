// A good producer-certified callable union read by destructuring, then called.
type Target = ((value: number) => number) | ((value: number) => string);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
const raw = { kind: 'Receiver' as const, value: (value: number): string => `n${value}` };
const base: { readonly kind: 'Receiver' } = raw;
const view = base as Receiver;
const { value } = view;
console.log(`${value(5)}`);
