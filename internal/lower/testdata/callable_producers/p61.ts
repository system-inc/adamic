type Target = ((value: number) => number) | ((value: string) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
class Scale { constructor(readonly factor: number) {} apply(value: number): number { return value * this.factor; } }
const scale = new Scale(3);
const raw = { kind: 'Receiver' as const, value: (value: number): number => scale.apply(value) };
const base: { readonly kind: 'Receiver' } = raw;
const view = base as Receiver;
const callable = view.value;
console.log(`${callable !== undefined}`);
